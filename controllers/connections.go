package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pubsubapi/env"
	"pubsubapi/errors"
	"pubsubapi/models"

	"github.com/gin-gonic/gin"
)

const (
	// sdkThingURL = "http://discretal-things:8182"
	// sdkUserURL  = "http://discretal-users:8180"
	sdkThingURL  = "http://localhost:8182"
	sdkUserURL   = "http://localhost:8180"
	sdkGroupURL  = "http://localhost:8189"
	sdkReaderURL = "http://localhost:8905"
	sdkWriterURL = "http://localhost:8185"
	sdkMQTTURL   = "localhost:1883"

	envThingURL  = "DC_THING_URL"
	envUserURL   = "DC_USER_URL"
	envGroupURL  = "DC_GROUP_URL"
	envReaderURL = "DC_READER_URL"
	envWriterURL = "DC_WRITER_URL"
	envMQTTURL   = "DC_MQTT_URL"
)

// connect one-to-many
func connect(c *gin.Context, thingid, channelid, owner string) (int, error) {
	fmt.Println("connect function...")
	connecturl := env.Env(envThingURL, sdkThingURL)
	connecturl = connecturl + "/connect"
	thingsurl := env.Env(envThingURL, sdkThingURL)
	thingsurl = thingsurl + "/things"
	channelsurl := env.Env(envThingURL, sdkThingURL)
	channelsurl = channelsurl + "/channels"

	param := make(map[string]string)
	param["email"] = owner
	fmt.Printf("param for connect -> %+v \n", param)

	var payload models.Connect
	if thingid == "" {
		fmt.Println("In case of channels...")
		// loop through all things to conn
		things, errcode, err := findThings(c, thingsurl, param)
		if err != nil {
			return errcode, err
		}

		// If no things yet
		if len(things) == 0 {
			return 0, nil
		}

		var thingsids []string
		for _, t := range things {
			thingsids = append(thingsids, t.ID)
		}
		payload = models.Connect{
			Channel_ids: []string{channelid},
			Thing_ids:   thingsids,
		}
	}
	if channelid == "" {
		fmt.Println("In case of things...")
		// loop through all channels to conn
		channels, errcode, err := findChannels(c, channelsurl, param)
		if err != nil {
			return errcode, err
		}

		// If no channels yet
		if len(channels) == 0 {
			return 0, nil
		}
		fmt.Println("Total channels :- ", len(channels))
		fmt.Println("retrieved channels -> ", channels)

		var channelids []string
		for _, ch := range channels {
			fmt.Println(ch.ID, " - ", ch.Name, " - ")
			channelids = append(channelids, ch.ID)
		}
		payload = models.Connect{
			Channel_ids: channelids,
			Thing_ids:   []string{thingid},
		}
	}

	data, _ := json.Marshal(payload)
	fmt.Println("data :- ", string(data))

	resp, err := httpReq(c, "POST", connecturl, data, nil)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("could not be connected : %v", err)
	}
	defer resp.Body.Close()

	fmt.Println("--> ", resp.StatusCode)
	respdata, _ := io.ReadAll(resp.Body)
	fmt.Println("==> ", string(respdata))
	var resperr models.RespError
	json.Unmarshal(respdata, &resperr)
	fmt.Println("--> ", resperr.Error)
	// if resp.StatusCode != http.StatusOK {
	// 	return fmt.Errorf("could not be connected : %v", err)
	// }
	return 0, nil
}

func getToken(c *gin.Context) (string, error) {
	if c.Request.Header.Get("Authorization") == "" {
		return "", fmt.Errorf("missing authorization key")
	}
	return c.Request.Header["Authorization"][0], nil
}

// connect one-to-one
// func connect(c *gin.Context, thingid, channelid string) (int, error) {
// 	connecturl := env.Env(envConnectURL, defConnectURL)
// 	var payload = models.Connect{
// 		Channel_ids: []string{channelid},
// 		Thing_ids:   []string{thingid},
// 	}

// 	data, _ := json.Marshal(payload)

// 	resp, err := httpReq(c, "POST", connecturl, data)
// 	if err != nil {
// 		return http.StatusInternalServerError, fmt.Errorf("could not be connected : %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// if resp.StatusCode != http.StatusOK {
// 	// 	return fmt.Errorf("could not be connected : %v", err)
// 	// }
// 	return 0, nil
// }

func httpReq(c *gin.Context, method, url string, data []byte, params map[string]string) (*http.Response, error) {
	var req *http.Request
	var err error

	if data == nil {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create new request : %v", err)
	}

	if c.Request.Header.Get("Authorization") == "" {
		return nil, fmt.Errorf("missing authorization key")
	}

	token := c.Request.Header["Authorization"][0]
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	//add query parameters, if any
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response : %v", err)
	}
	return resp, nil
}

func findThings(c *gin.Context, url string, params map[string]string) ([]models.ThingRes, int, error) {
	resp, err := httpReq(c, "GET", url, nil, params)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("request could not be processed by server : %v", err)
	}
	defer resp.Body.Close()

	if !(resp.StatusCode == 200 || resp.StatusCode == 201) {
		return nil, resp.StatusCode, fmt.Errorf("request could not be processed by server : unauthorized access")
	}

	var thingsList models.ThingsList

	json.NewDecoder(resp.Body).Decode(&thingsList)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to Decode the thingsList : %v", err)
	}

	return thingsList.Things, 0, nil
}

func findChannels(c *gin.Context, url string, params map[string]string) ([]models.ChannelResAll, int, error) {
	resp, err := httpReq(c, "GET", url, nil, params)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("request could not be processed by server : %v", err)
	}
	defer resp.Body.Close()

	var channelsList models.ChannelsList

	json.NewDecoder(resp.Body).Decode(&channelsList)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to Decode the channelsList : %v", err)
	}

	return channelsList.Channels, 0, nil
}

func findGroups(c *gin.Context, url string) ([]models.GroupRes, int, error) {
	resp, err := httpReq(c, "GET", url, nil, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("request could not be processed by server : %v", err)
	}
	defer resp.Body.Close()

	if !(resp.StatusCode == 200 || resp.StatusCode == 201) {
		return nil, resp.StatusCode, fmt.Errorf("request could not be processed by server : unauthorized access")
	}

	var groupsList models.GroupsList

	json.NewDecoder(resp.Body).Decode(&groupsList)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to Decode the groupsList : %v", err)
	}

	return groupsList.Groups, 0, nil
}

func getUser(c *gin.Context) (models.User, error) {
	url := env.Env(envUserURL, sdkUserURL)
	url = url + "/users/profile"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return models.User{}, fmt.Errorf("request could not be processed by server : %v", err)
	}

	if c.Request.Header.Get("Authorization") == "" {
		return models.User{}, fmt.Errorf("missing authorization key")
	}
	token := c.Request.Header["Authorization"][0]

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.User{}, fmt.Errorf("request could not be processed by server : %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.User{}, fmt.Errorf("error occurred while reading response : %v", err)
	}

	var u models.User
	if err := json.Unmarshal(data, &u); err != nil {
		return models.User{}, err
	}

	return u, nil
}

func withOutput(op string) string {
	if op != "" {
		return " with errors - " + op
	} else {
		return " without errors."
	}
}

func errcodeHandler(c *gin.Context, errstr string, errcode int, err error) {
	if errcode > 0 {
		errors.ErrHandler(c, fmt.Errorf("%v : %v", errstr, err), errcode)
	} else {
		errors.ErrHandler(c, fmt.Errorf("%v : %v", errstr, err), http.StatusBadRequest)
	}
}
