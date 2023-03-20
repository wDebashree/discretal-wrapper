package controllers

import (
	// "encoding/json"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	// "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	// sdk "github.com/mainflux/mainflux/pkg/sdk/go"

	// "io"
	"net/http"
	"pubsubapi/env"
	"pubsubapi/errors"
	"pubsubapi/models"
)

// CreateThing	godoc
//
//	@Summary		Adds new thing
//	@Description	Adds new thing to the list of things owned by user identified using the provided access token.
//	@Tags			things
//	@Produce		json
//	@Param			Request	body		models.ThingReq	true	"JSON-formatted document describing the new thing."
//
//	@Success		201		{object}	models.ThingRes	"Thing registered."
//	@Failure		400		"Failed due to malformed JSON."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things [post]
func CreateThing(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// to test whether CreateThing requirements are met or not
	var testThing models.ThingReq
	if err := c.ShouldBindJSON(&testThing); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	// configure the sdk payload for forwarding request
	url := env.Env(envThingURL, sdkThingURL)
	url = url + "/things"

	// To check whether channelname already exists or not
	thingsList, errcode, err := findThings(c, url)
	if err != nil {
		if errcode > 0 {
			errors.ErrHandler(c, fmt.Errorf("error occurred while checking naming conflict : %v", err), errcode)
		} else {
			errors.ErrHandler(c, fmt.Errorf("error occurred while checking naming conflict : %v", err), http.StatusBadRequest)
		}
		return
	}

	var thing models.ThingRes
	for _, thng := range thingsList {
		if thng.Name == testThing.Name {
			thing = thng
			break
		}
	}

	if thing.ID != "" {
		errors.InfoHandler(c, "thing name already exists", thing, http.StatusAlreadyReported)
		return
	}

	thingdata, _ := json.Marshal(testThing)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(thingdata))
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
	id := strings.TrimPrefix(resp.Header.Get("Location"), fmt.Sprintf("/%s/", "things"))

	errcode, err = connect(c, id, "")
	if err != nil {
		if errcode > 0 {
			errors.ErrHandler(c, fmt.Errorf("error occurred in connecting with channels : %v", err), errcode)
		} else {
			errors.ErrHandler(c, fmt.Errorf("error occurred in connecting with channels : %v", err), http.StatusBadRequest)
		}
		return
	}

	req, err = http.NewRequest("GET", url+"/"+id, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		return
	}

	var thingresp models.ThingRes
	err = json.Unmarshal(data, &thingresp)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "thing created successfully", thingresp, http.StatusCreated)
}

// GetThings	godoc
//
//	@Summary		Retrieves things
//	@Description	Retrieves a list of things. Due to performance concerns, data is retrieved in subsets. The API things must ensure that the entire dataset is consumed either by making subsequent requests, or by increasing the subset size of the initial request.
//	@Tags			things
//	@Produce		json
//
//	@Param			thingsParams	query		controllers.GetThings.ThingsParams	false	"thingsParams"
//
//	@Success		200				{object}	models.ThingsList					"Data retrieved."
//	@Failure		400				"Failed due to malformed query parameters."
//	@Failure		401				"Missing or invalid access token provided."
//	@Failure		500				"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things [get]
func GetThings(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	url := env.Env(envThingURL, sdkThingURL)
	req, err := http.NewRequest("GET", url+"/things", nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	//add query parameters
	q := req.URL.Query()
	// params := []string{"offset", "limit", "disconnected", "order", "dir"}
	params := []string{"offset", "limit", "disconnected", "order", "dir", "name"}
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
	type thingsPageRes struct {
		pageRes
		Things []models.ThingResAll `json:"things"`
	}

	var allThings thingsPageRes
	err = json.Unmarshal(data, &allThings)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "things retrieved successfully", allThings, http.StatusOK)
}

// GetThing	godoc
//
//	@Summary		Retrieves thing info
//	@Description	Retrieves the details of a thing
//	@Tags			things
//	@Produce		json
//
//	@Param			name	path		string			true	"Unique thing name."
//
//	@Success		200		{object}	models.ThingRes	"Data retrieved."
//	@Failure		400		"Failed due to malformed thing's ID."
//	@Failure		404		"Thing does not exist."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things/{name} [get]
func GetThing(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for thing details
	var Item models.ItemId
	if err := c.ShouldBindUri(&Item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	// configure the sdk payload for forwarding request
	url := env.Env(envThingURL, sdkThingURL)
	url = url + "/things/" + Item.Id
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
		errors.ErrHandler(c, fmt.Errorf(`invalid thing ID`), http.StatusBadRequest)
		return
	}

	var thingres models.ThingRes
	err = json.Unmarshal(data, &thingres)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "thing retrieved successfully", thingres, http.StatusOK)
}

func UpdateThing(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for thing's new details
	var thingid models.ItemId
	if err := c.ShouldBindUri(&thingid); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}
	var thingdata models.ThingReq
	if err := c.ShouldBindJSON(&thingdata); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	// configure the sdk payload for forwarding request
	url := env.Env(envThingURL, sdkThingURL)
	url = url + "/things/" + thingid.Id

	data, _ := json.Marshal(thingdata)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
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
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		data, _ := io.ReadAll(resp.Body)
		var err models.RespError
		json.Unmarshal(data, &err)
		errors.ErrHandler(c, fmt.Errorf("thing could not be updated"), resp.StatusCode)
		return
	}

	errors.InfoHandler(c, "thing updated successfully", "thing updated successfully", http.StatusCreated)
}

// DeleteThing	godoc
//
//	@Summary		Removes a thing
//	@Description	Removes a thing. The service will ensure that the removed thing is disconnected from all of the existing channels.
//	@Tags			things
//	@Produce		json
//
//	@Param			name	path	string	true	"Unique thing name."
//
//	@Success		204		"Thing removed."
//	@Failure		400		"Failed due to malformed thing's ID."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things/{name} [delete]
func DeleteThing(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for thing details
	var Item models.ItemId
	if err := c.ShouldBindUri(&Item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	// configure the sdk payload for forwarding request
	url := env.Env(envThingURL, sdkThingURL)
	url = url + "/things/" + Item.Id
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
		errors.ErrHandler(c, fmt.Errorf(`invalid thing ID`), http.StatusBadRequest)
		return
	}

	var thing models.ThingRes
	err = json.Unmarshal(data, &thing)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	if thing.Name == "cloudEnd" {
		errors.ErrHandler(c, fmt.Errorf(`default thing "cloudEnd" cannot be deleted`), http.StatusBadRequest)
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

	errors.InfoHandler(c, "thing deleted successfully", "thing deleted successfully", http.StatusNoContent)
}

func GetConnectedChannels(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for thing details
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
	// res, err := mfsdk.Things(token, 0, 1, thingName.Name)
	// if err != nil {
	// 	errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	// 	return
	// }
	// if len(res.Things) == 0 {
	// 	errors.ErrHandler(c, fmt.Errorf("thing name does not exist"), http.StatusNotFound)
	// 	return
	// }

	// thingID := res.Things[0].ID
	// newURL := url + "/things/" + thingID + "/channels"
	newURL := url + "/things/" + Item.Id + "/channels"

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

	type ConnChannels struct {
		Total     uint64              `json:"total"`
		Offset    uint64              `json:"offset"`
		Limit     uint64              `json:"limit"`
		Order     string              `json:"order"`
		Direction string              `json:"direction"`
		Channels  []models.ChannelRes `json:"channels"`
	}

	var connchannels ConnChannels
	err = json.Unmarshal(data, &connchannels)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "connected things retrieved successfully", connchannels, http.StatusOK)
}

func GetConnectedGroups(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for thing details
	var Item models.ItemId
	if err := c.ShouldBindUri(&Item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	urlGroup := env.Env(envGroupURL, sdkGroupURL)
	newURL := urlGroup + "/members/" + Item.Id + "/groups"

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

	type viewGroupRes struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	type pageRes struct {
		Limit  uint64 `json:"limit,omitempty"`
		Offset uint64 `json:"offset,omitempty"`
		Total  uint64 `json:"total"`
		Level  uint64 `json:"level"`
		Name   string `json:"name"`
	}
	type groupPageRes struct {
		pageRes
		Groups []viewGroupRes `json:"groups"`
	}

	var conngroups groupPageRes
	err = json.Unmarshal(data, &conngroups)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "connected groups retrieved successfully", conngroups, http.StatusOK)
}

func AssignGroups(c *gin.Context) {
	path := strings.SplitAfter(c.Request.URL.Path, "/api/")
	grouptype := strings.Split(path[1], "/")[0]

	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for thing details
	var Item models.ItemId
	if err := c.ShouldBindUri(&Item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}
	var groups models.AssignGroupReq
	if err := c.ShouldBindJSON(&groups); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}
	groups.Type = grouptype
	groupsdata, _ := json.Marshal(groups)

	urlGroup := env.Env(envGroupURL, sdkGroupURL)
	newURL := urlGroup + "/members/" + Item.Id + "/groups"

	req, err := http.NewRequest("POST", newURL, bytes.NewBuffer(groupsdata))
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

	// data, _ := io.ReadAll(resp.Body)
	// fmt.Println("data --> ", string(data))

	if (resp.StatusCode != 200) && (resp.StatusCode != 204) {
		fmt.Println()
		errors.ErrHandler(c, fmt.Errorf(`group could not be added to the member`), http.StatusBadRequest)
		return
	}

	// var updateURL string
	// if groups.Type == "things" {
	// 	updateURL = env.Env(envThingURL, sdkThingURL)
	// 	updateURL = updateURL + "/things/" + Item.Id
	// } else if groups.Type == "users" {
	// 	updateURL = env.Env(envUserURL, sdkUserURL)
	// 	updateURL = updateURL + "/users/" + Item.Id
	// }

	// var updatemember models.ThingReq
	// updatemember.Metadata = make(map[string]interface{}, 1)
	// updatemember.Metadata["addgroup"] = groups.Groups
	// addmemberdata, _ := json.Marshal(updatemember)
	// fmt.Printf("updatemember --> %+v \n", updatemember)

	// req, err = http.NewRequest("PUT", updateURL, bytes.NewBuffer(addmemberdata))
	// if err != nil {
	// 	errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	// }
	// req.Header.Add("Content-Type", "application/json")
	// req.Header.Add("Authorization", token)
	// client = &http.Client{}
	// _, err = client.Do(req)
	// if err != nil {
	// 	errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	// }

	errors.InfoHandler(c, "groups added to member successfully", "groups added to member successfully", http.StatusNoContent)
}

func UnassignGroups(c *gin.Context) {
	path := strings.SplitAfter(c.Request.URL.Path, "/api/")
	grouptype := strings.Split(path[1], "/")[0]

	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for thing details
	var Item models.ItemId
	if err := c.ShouldBindUri(&Item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}
	var groups models.AssignGroupReq
	if err := c.ShouldBindJSON(&groups); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}
	groups.Type = grouptype
	groupsdata, _ := json.Marshal(groups)

	urlGroup := env.Env(envGroupURL, sdkGroupURL)
	newURL := urlGroup + "/members/" + Item.Id + "/groups"

	req, err := http.NewRequest("DELETE", newURL, bytes.NewBuffer(groupsdata))
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

	if (resp.StatusCode != 200) && (resp.StatusCode != 204) {
		errors.ErrHandler(c, fmt.Errorf(`group could not be removed from the member`), http.StatusBadRequest)
		return
	}

	// var updateURL string
	// if groups.Type == "things" {
	// 	updateURL = env.Env(envThingURL, sdkThingURL)
	// 	updateURL = updateURL + "/things/" + Item.Id
	// } else if groups.Type == "users" {
	// 	updateURL = env.Env(envUserURL, sdkUserURL)
	// 	updateURL = updateURL + "/users/" + Item.Id
	// }

	// var updatemember models.ThingReq
	// updatemember.Metadata = make(map[string]interface{}, 1)
	// updatemember.Metadata["removegroup"] = groups.Groups
	// addmemberdata, _ := json.Marshal(updatemember)

	// req, err = http.NewRequest("PUT", updateURL, bytes.NewBuffer(addmemberdata))
	// if err != nil {
	// 	errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	// }
	// req.Header.Add("Content-Type", "application/json")
	// req.Header.Add("Authorization", token)
	// client = &http.Client{}
	// resp, err = client.Do(req)
	// if err != nil {
	// 	errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), resp.StatusCode)
	// }

	errors.InfoHandler(c, "groups removed from member successfully", "groups removed from member successfully", http.StatusNoContent)
}
