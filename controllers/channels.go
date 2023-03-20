package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	// "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	// sdk "github.com/mainflux/mainflux/pkg/sdk/go"

	"net/http"
	"pubsubapi/env"
	"pubsubapi/errors"
	"pubsubapi/models"
)

// CreateChannel	godoc
//
//	@Summary		Adds new channel
//	@Description	Creates new channel. User identified by the provided access token will be the channels owner.
//	@Tags			channels
//	@Produce		json
//	@Param			Request	body		models.ChannelReq	true	"JSON-formatted document describing the updated channel."
//
//	@Success		201		{object}	models.ChannelRes	"Channel created."
//	@Failure		400		"Failed due to malformed JSON."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/channels [post]
func CreateChannel(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// To test whether CreateChannel requirements are met or not
	var testChannel models.ChannelReq
	if err := c.ShouldBindJSON(&testChannel); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	// configure the sdk payload for forwarding request
	url := env.Env(envThingURL, sdkThingURL)
	url = url + "/channels"

	// To check whether channelname already exists or not
	channelsList, errcode, err := findChannels(c, url)
	if err != nil {
		if errcode > 0 {
			errors.ErrHandler(c, fmt.Errorf("error occurred while checking naming conflict : %v", err), errcode)
		} else {
			errors.ErrHandler(c, fmt.Errorf("error occurred while checking naming conflict : %v", err), http.StatusBadRequest)
		}
		return
	}

	var channel models.ChannelResAll
	for _, chn := range channelsList {
		if chn.Name == testChannel.Name {
			channel = chn
			break
		}
	}

	if channel.ID != "" {
		errors.InfoHandler(c, "channel name already exists", channel, http.StatusAlreadyReported)
		return
	}

	channeldata, _ := json.Marshal(testChannel)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(channeldata))
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()
	id := strings.TrimPrefix(resp.Header.Get("Location"), fmt.Sprintf("/%s/", "channels"))

	// TBD - connecting new channels with existing things automatically - below is the functionality
	errcode, err = connect(c, "", id)
	if err != nil {
		if errcode > 0 {
			errors.ErrHandler(c, fmt.Errorf("error occurred in connecting with things : %v", err), errcode)
		} else {
			errors.ErrHandler(c, fmt.Errorf("error occurred in connecting with things : %v", err), http.StatusBadRequest)
		}
		return
	}

	var channelresp = models.ChannelRes{
		ID:   id,
		Name: testChannel.Name,
	}

	errors.InfoHandler(c, "channel created successfully", channelresp, http.StatusCreated)
}

// GetChannels	godoc
//
//	@Summary		Retrieves channels
//	@Description	Retrieves a list of channels. Due to performance concerns, data is retrieved in subsets. The API things must ensure that the entire dataset is consumed either by making subsequent requests, or by increasing the subset size of the initial request.
//	@Tags			channels
//	@Produce		json
//
//	@Param			channelsParams	query		controllers.GetChannels.ChannelsParams	false	"channelsParams"
//
//	@Success		200				{object}	models.ChannelsList						"Data retrieved."
//	@Failure		400				"Failed due to malformed query parameters."
//	@Failure		401				"Missing or invalid access token provided."
//	@Failure		500				"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/channels [get]
func GetChannels(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	url := env.Env(envThingURL, sdkThingURL)
	req, err := http.NewRequest("GET", url+"/channels", nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	//add query parameters
	q := req.URL.Query()
	// params := []string{"offset", "limit", "disconnected", "order", "dir"}
	params := []string{"offset", "limit", "disconnected", "order", "dir", "email", "name"}
	for _, prms := range params {
		val, ok := c.GetQuery(prms)
		if ok {
			q.Add(prms, val)
		}
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		return
	}

	type pageRes struct {
		Total     uint64 `json:"total"`
		Offset    uint64 `json:"offset"`
		Limit     uint64 `json:"limit"`
		Order     string `json:"order,omitempty"`
		Direction string `json:"dir,omitempty"`
		IsAdmin   bool   `json:"isadmin,omitempty"`
	}
	type channelsPageRes struct {
		pageRes
		Channels []models.ChannelResAll `json:"channels"`
	}

	var allChannels channelsPageRes
	err = json.Unmarshal(data, &allChannels)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "channels retrieved successfully", allChannels, http.StatusOK)
}

// GetChannel	godoc
//
//	@Summary		Retrieves channel info
//	@Description	Retrieves the details of a channel
//	@Tags			channels
//	@Produce		json
//
//	@Param			name	path		string				true	"Unique channel name."
//
//	@Success		200		{object}	models.ChannelRes	"Data retrieved."
//	@Failure		400		"Failed due to malformed channel's ID."
//	@Failure		404		"Channel does not exist."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/channels/{name} [get]
func GetChannel(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for channel details
	var Item models.ItemId
	if err := c.ShouldBindUri(&Item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	// configure the sdk payload for forwarding request
	url := env.Env(envThingURL, sdkThingURL)
	url = url + "/channels/" + Item.Id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		return
	}
	if (resp.StatusCode != 200) && (resp.StatusCode != 201) {
		errors.ErrHandler(c, fmt.Errorf(`invalid channel ID`), http.StatusBadRequest)
		return
	}

	var channelres models.ChannelRes
	err = json.Unmarshal(data, &channelres)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "channel retrieved successfully", channelres, http.StatusOK)
}

// DeleteChannel	godoc
//
//	@Summary		Removes a channel
//	@Description	Removes a channel. The service will ensure that the subscribed apps and things are unsubscribed from the removed channel.
//	@Tags			channels
//	@Produce		json
//
//	@Param			name	path	string	true	"Unique channel name."
//
//	@Success		204		"Channel removed."
//	@Failure		400		"Failed due to malformed channel's ID."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/channels/{name} [delete]
func DeleteChannel(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for channel details
	var Item models.ItemId
	if err := c.ShouldBindUri(&Item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	// configure the sdk payload for forwarding request
	url := env.Env(envThingURL, sdkThingURL)
	url = url + "/channels/" + Item.Id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		return
	}
	if (resp.StatusCode != 200) && (resp.StatusCode != 201) {
		errors.ErrHandler(c, fmt.Errorf(`invalid channel ID`), http.StatusBadRequest)
		return
	}

	var channel models.ChannelRes
	err = json.Unmarshal(data, &channel)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	if channel.Name == "cloudToDevice" || channel.Name == "deviceToCloud" {
		errors.ErrHandler(c, fmt.Errorf(`default channels "cloudToDevice" or "deviceToCloud" cannot be deleted`), http.StatusBadRequest)
		return
	}

	//delete the channel by its ID
	req, err = http.NewRequest("DELETE", url, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	errors.InfoHandler(c, "channel deleted successfully", "channel deleted successfully", http.StatusNoContent)
}

func GetConnectedThings(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for channel details
	var Item models.ItemId
	if err := c.ShouldBindUri(&Item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	url := env.Env(envThingURL, sdkThingURL)
	// // configure the sdk payload for forwarding request
	// config := sdk.Config{
	// 	BaseURL: url,
	// }
	// mfsdk := sdk.NewSDK(config)

	// //retrieve channels by passing the name in argument
	// res, err := mfsdk.Channels(token, 0, 1, channelName.Name)
	// if err != nil {
	// 	errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	// 	return
	// }
	// if len(res.Channels) == 0 {
	// 	errors.ErrHandler(c, fmt.Errorf("channel name does not exist"), http.StatusNotFound)
	// 	return
	// }

	// chanID := res.Channels[0].ID
	// newURL := url + "/channels/" + chanID + "/things"
	newURL := url + "/channels/" + Item.Id + "/things"

	req, err := http.NewRequest("GET", newURL, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	//add query parameters
	q := req.URL.Query()
	params := []string{"offset", "limit", "disconnected", "order", "dir"}
	for _, prms := range params {
		val, ok := c.GetQuery(prms)
		if ok {
			q.Add(prms, val)
		}
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		return
	}

	type ConnThings struct {
		Total     uint64            `json:"total"`
		Offset    uint64            `json:"offset"`
		Limit     uint64            `json:"limit"`
		Order     string            `json:"order"`
		Direction string            `json:"direction"`
		Things    []models.ThingRes `json:"things"`
	}

	var connthings ConnThings
	err = json.Unmarshal(data, &connthings)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "connected things retrieved successfully", connthings, http.StatusOK)
}

func GetMessages(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for channel details
	var Item models.ItemId
	if err := c.ShouldBindUri(&Item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	// url := env.Env(envThingURL, sdkThingURL)
	// // configure the sdk payload for forwarding request
	// config := sdk.Config{
	// 	BaseURL: url,
	// }
	// mfsdk := sdk.NewSDK(config)

	// //retrieve channels by passing the name in argument
	// res, err := mfsdk.Channels(token, 0, 1, channelName.Name)
	// if err != nil {
	// 	errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	// 	return
	// }
	// if len(res.Channels) == 0 {
	// 	errors.ErrHandler(c, fmt.Errorf("channel name does not exist"), http.StatusNotFound)
	// 	return
	// }

	// chanID := res.Channels[0].ID
	readerurl := env.Env(envReaderURL, sdkReaderURL)
	// newURL := readerurl + "/channels/" + chanID + "/messages"
	newURL := readerurl + "/channels/" + Item.Id + "/messages"

	req, err := http.NewRequest("GET", newURL, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	//add query parameters
	q := req.URL.Query()
	params := []string{"offset", "limit", "format", "subtopic", "publisher", "protocol", "name", "v", "comparator", "vb", "vs", "vd", "from", "to"}
	for _, prms := range params {
		val, ok := c.GetQuery(prms)
		if ok {
			// if prms == "publisher" {
			// 	//retrieve thingID by passing the name in argument
			// 	res, err := mfsdk.Things(token, 0, 1, val)
			// 	if err != nil {
			// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
			// 		return
			// 	}
			// 	if len(res.Things) == 0 {
			// 		errors.ErrHandler(c, fmt.Errorf("publisher name does not exist"), http.StatusNotFound)
			// 		return
			// 	}
			// 	thingID := res.Things[0].ID
			// 	q.Add(prms, thingID)
			// } else {
			// 	q.Add(prms, val)
			// }
			q.Add(prms, val)
		}
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		return
	}

	type Message interface{}
	type ResMessages struct {
		Total       uint64    `json:"total"`
		Offset      uint64    `json:"offset"`
		Limit       uint64    `json:"limit"`
		Subtopic    string    `json:"subtopic,omitempty"`
		Publisher   string    `json:"publisher,omitempty"`
		Protocol    string    `json:"protocol,omitempty"`
		Name        string    `json:"name,omitempty"`
		Value       float64   `json:"v,omitempty"`
		Comparator  string    `json:"comparator,omitempty"`
		BoolValue   bool      `json:"vb,omitempty"`
		StringValue string    `json:"vs,omitempty"`
		DataValue   string    `json:"vd,omitempty"`
		From        float64   `json:"from,omitempty"`
		To          float64   `json:"to,omitempty"`
		Format      string    `json:"format,omitempty"`
		Messages    []Message `json:"messages,omitempty"`
	}

	var msgs ResMessages
	err = json.Unmarshal(data, &msgs)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "connected things retrieved successfully", msgs, http.StatusOK)
}

func SendMessages(c *gin.Context) {
	path := strings.SplitAfterN(c.Request.URL.Path, "/", 5)
	endurl := strings.Join(path[4:], "")

	// // check if auth token exists
	// token, err := getToken(c)
	// if err != nil {
	// 	errors.ErrHandler(c, err, http.StatusUnauthorized)
	// 	return
	// }

	// Retrieve the params for channel details
	var Item models.ItemId
	if err := c.ShouldBindUri(&Item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	// // configure the sdk payload for forwarding request
	// url := env.Env(envThingURL, sdkThingURL)
	// config := sdk.Config{
	// 	BaseURL: url,
	// }
	// mfsdk := sdk.NewSDK(config)

	// //retrieve channels by passing the name in argument
	// res, err := mfsdk.Channels(token, 0, 1, channelName.Name)
	// if err != nil {
	// 	errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	// 	return
	// }
	// if len(res.Channels) == 0 {
	// 	errors.ErrHandler(c, fmt.Errorf("channel name does not exist"), http.StatusNotFound)
	// 	return
	// }

	// chanID := res.Channels[0].ID
	writerurl := env.Env(envWriterURL, sdkWriterURL)
	// newURL := writerurl + "/channels/" + chanID + "/" + endurl
	newURL := writerurl + "/channels/" + Item.Id + "/" + endurl

	data, _ := io.ReadAll(c.Request.Body)

	type Msg map[string]interface{}
	var msgs []Msg
	_ = json.Unmarshal(data, &msgs)

	var thingid, thingkey, protocol string
	for _, msg := range msgs {
		for k, v := range msg {
			if k == "ukey" {
				thingkey = v.(string)
				delete(msg, k)
			}
			if k == "uid" {
				thingid = v.(string)
				delete(msg, k)
			}
			if k == "protocol" {
				protocol = v.(string)
				if protocol == "mqtt" {
					msg[k] = "mqtt_c"
				}
			}
		}
	}

	newdata, _ := json.Marshal(msgs)

	fmt.Println("thingID:", thingid)
	fmt.Println("thingKey:", thingkey)

	// if protocol == "mqtt" {
	// 	// Publish message to mqtt ->
	// 	mqttURL := env.Env(envMQTTURL, sdkMQTTURL)
	// 	opts := mqtt.NewClientOptions().AddBroker(mqttURL).SetClientID("myTestClient01").SetUsername(thingid).SetPassword(thingkey)
	// 	client := mqtt.NewClient(opts)
	// 	if token := client.Connect(); token.Wait() && token.Error() != nil {
	// 		errors.ErrHandler(c, fmt.Errorf("error at mqtt client connection : %v", token.Error()), http.StatusInternalServerError)
	// 		panic(token.Error())
	// 	}

	// 	// err = publish(c, client, chanID, newdata)
	// 	// if err != nil {
	// 	// 	errors.ErrHandler(c, fmt.Errorf("error occurred while publishing message : %v", err), http.StatusInternalServerError)
	// 	// 	return
	// 	// }
	// 	client.Disconnect(250)
	// 	errors.InfoHandler(c, "mqtt message sent successfully", "mqtt message sent successfully", http.StatusOK)

	// }

	fmt.Println("forward to http api -->")
	req, err := http.NewRequest("POST", newURL, bytes.NewBuffer(newdata))
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Thing "+thingkey)
	req.Header.Set("Content-Type", "application/json")

	httpclient := &http.Client{}
	resp, err := httpclient.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	fmt.Println("response status: ", resp.StatusCode)
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	fmt.Println("response data: ", string(data))
}
