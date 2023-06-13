package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"net/http"
	"pubsubapi/env"
	"pubsubapi/errors"
	"pubsubapi/models"

	"github.com/gin-gonic/gin"
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
	// to test whether CreateThing requirements are met or not
	var testThing models.ThingReq
	if err := c.ShouldBindJSON(&testThing); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	url := env.Env(envThingURL, sdkThingURL)
	url = url + "/things"

	// To check whether thingname already exists or not
	thingsList, errcode, err := findThings(c, url, nil)
	if err != nil {
		errcodeHandler(c, "error occurred while checking naming conflict", errcode, err)
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

	// If thingname does not exist, send request to add it
	thingdata, _ := json.Marshal(testThing)
	resp, err := httpReq(c, "POST", url, thingdata, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		return
	}

	var resperr models.RespError
	_ = json.Unmarshal(data, &resperr)
	if resperr.Error != "" {
		errors.InfoHandler(c, "thing creation completed"+withOutput(resperr.Error), resperr, resp.StatusCode)
		return
	}

	// Get the new thing ID along with user details to connect with channels
	id := strings.TrimPrefix(resp.Header.Get("Location"), fmt.Sprintf("/%s/", "things"))
	user, err := getUser(c)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred in connecting with things : %v", err), http.StatusBadRequest)
	}

	// connect new thing to all existing channels
	errcode, err = connect(c, id, "", user.Email)
	if err != nil {
		errcodeHandler(c, "error occurred in connecting with channels", errcode, err)
		return
	}

	// get the newly created id details
	resp, err = httpReq(c, "GET", url+"/"+id, nil, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		return
	}

	type Response struct {
		*models.ThingRes `json:",omitempty"`
		Error            string `json:"error,omitempty"`
	}
	var response Response
	_ = json.Unmarshal(data, &response)

	errors.InfoHandler(c, "thing creation completed"+withOutput(response.Error), response, resp.StatusCode)
}

// GetThings	godoc
//
//	@Summary		Retrieves things
//	@Description	Retrieves a list of things. Due to performance concerns, data is retrieved in subsets. The API things must ensure that the entire dataset is consumed either by making subsequent requests, or by increasing the subset size of the initial request.
//	@Tags			things
//	@Produce		json
//
//	@Param			limit			query		integer					false	"Size of the subset to retrieve."
//	@Param			offset			query		integer					false	"Number of items to skip during retrieval."
//	@Param			name			query		string					false	"Unique thing name."
//	@Param			order			query		string					false	"Entity to be sorted on."
//	@Param			dir				query		string					false	"Asc or Desc sorting."
//	@Param			disconnected	query		bool					false	"Disconnected true or false."
//	@Param			email			query		string					false	"Email ID of selected user."
//	@Param			gids			query		array					false	"Array of group IDs."
//
//	@Success		200				{object}	models.ThingsPageRes	"Data retrieved."
//	@Failure		400				"Failed due to malformed query parameters."
//	@Failure		401				"Missing or invalid access token provided."
//	@Failure		500				"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things [get]
func GetThings(c *gin.Context) {
	var data []byte
	params := make(map[string]string)

	var ids []string
	if gids, ok := c.GetQuery("gids"); ok {
		ids = strings.Split(gids, ",")

		// if groupname, ok := c.GetQuery("grpname"); ok {
		// 	urlGroup := env.Env(envGroupURL, sdkGroupURL)
		// 	urlGroup = urlGroup + "/groups"

		// 	params["name"] = groupname

		// 	resp, err := httpReq(c, "GET", urlGroup, nil, params)
		// 	if err != nil {
		// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		// 		return
		// 	}
		// 	defer resp.Body.Close()

		// 	data, err = io.ReadAll(resp.Body)
		// 	if err != nil {
		// 		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		// 		return
		// 	}

		// 	var allGroups models.GroupPageRes
		// 	err = json.Unmarshal(data, &allGroups)
		// 	if err != nil {
		// 		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		// 		return
		// 	}

		// 	// var ids []string
		// 	for _, gp := range allGroups.Groups {
		// 		ids = append(ids, gp.ID)
		// 	}
		// }

		var payload = models.GroupIDs{
			IDs: ids,
		}
		groupdata, _ := json.Marshal(payload)

		urlThings := env.Env(envThingURL, sdkThingURL)
		urlThings = urlThings + "/groups"

		// delete(params, "name") // deleting the groupname passed in previous endpoint, in order to reuse the same map table

		paramarr := []string{"limit", "offset", "name", "order", "dir"}
		for _, prms := range paramarr {
			val, ok := c.GetQuery(prms)
			if ok {
				params[prms] = val
			}
		}

		newresp, err := httpReq(c, "GET", urlThings, groupdata, params)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
			return
		}

		data, err = io.ReadAll(newresp.Body)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
			return
		}
		defer newresp.Body.Close()

	} else {
		url := env.Env(envThingURL, sdkThingURL)
		url = url + "/things"

		// delete(params, "name") // deleting the groupname passed in previous endpoint, in order to reuse the same map table

		paramarr := []string{"offset", "limit", "disconnected", "order", "dir", "email", "name"}
		for _, prms := range paramarr {
			val, ok := c.GetQuery(prms)
			if ok {
				params[prms] = val
			}
		}

		resp, err := httpReq(c, "GET", url, nil, params)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
			return
		}

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()
	}

	var allThings models.ThingsPageRes
	err := json.Unmarshal(data, &allThings)
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
//	@Param			id	path		string			true	"Unique thing id."
//
//	@Success		200	{object}	models.ThingRes	"Data retrieved."
//	@Failure		400	"Failed due to malformed thing's ID."
//	@Failure		404	"Thing does not exist."
//	@Failure		401	"Missing or invalid access token provided."
//	@Failure		500	"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things/{id} [get]
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

// UpdateThing	godoc
//
//	@Summary		Updates thing info
//	@Description	Updates the details of a thing
//	@Tags			things
//	@Produce		json
//
//	@Param			id		path	string			true	"Unique thing id."
//	@Param			Request	body	models.ThingReq	true	"JSON-formatted document describing the updated thing."
//
//	@Success		200		"Thing updated."
//	@Failure		400		"Failed due to malformed thing's ID."
//	@Failure		404		"Thing does not exist."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things/{id} [put]
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
//	@Param			id	path	string	true	"Unique thing id."
//
//	@Success		204	"Thing removed."
//	@Failure		400	"Failed due to malformed thing's ID."
//	@Failure		401	"Missing or invalid access token provided."
//	@Failure		500	"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things/{id} [delete]
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

	data, _ = io.ReadAll(resp.Body)

	var response models.RespError
	_ = json.Unmarshal(data, &response)

	var withOutput string
	if response.Error != "" {
		withOutput = " with errors - " + response.Error
	} else {
		withOutput = " without errors."
	}
	errors.InfoHandler(c, "thing deletion operation completed"+withOutput, response, resp.StatusCode)
}

// GetConnectedChannels	godoc
//
//	@Summary		Retrieves connected channels
//	@Description	Retrieves a list of channels that are connected to the thing. Due to performance concerns, data is retrieved in subsets.
//	@Tags			things
//	@Produce		json
//
//	@Param			id				path		string											true	"Unique thing id."
//	@Param			limit			query		integer											false	"Size of the subset to retrieve."
//	@Param			offset			query		integer											false	"Number of items to skip during retrieval."
//	@Param			order			query		string											false	"Entity to be sorted on."
//	@Param			dir				query		string											false	"Asc or Desc sorting."
//	@Param			disconnected	query		bool											false	"Disconnected true or false."
//
//	@Success		200				{object}	controllers.GetConnectedChannels.ConnChannels	"Data retrieved."
//	@Failure		400				"Failed due to malformed query parameters."
//	@Failure		401				"Missing or invalid access token provided."
//	@Failure		500				"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things/{id}/channels [get]
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

// GetConnectedGroups	godoc
//
//	@Summary		Retrieves connected groups
//	@Description	Retrieves a list of groups to which the thing is a member.
//	@Tags			things
//	@Produce		json
//
//	@Param			id				path		string											true	"Unique thing id."
//	@Param			limit			query		integer											false	"Size of the subset to retrieve."
//	@Param			offset			query		integer											false	"Number of items to skip during retrieval."
//	@Param			order			query		string											false	"Entity to be sorted on."
//	@Param			dir				query		string											false	"Asc or Desc sorting."
//	@Param			disconnected	query		bool											false	"Disconnected true or false."
//
//	@Success		200				{object}	controllers.GetConnectedGroups.SuccessResponse	"Data retrieved."
//	@Failure		400				"Failed due to malformed query parameters."
//	@Failure		401				"Missing or invalid access token provided."
//	@Failure		500				"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things/{id}/groups [get]
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

	type SuccessResponse struct {
		models.GrpPageRes
		Groups []models.ViewGroupRes `json:"groups"`
	}
	type Response struct {
		*SuccessResponse `json:",omitempty"`
		Error            string `json:"error,omitempty"`
	}

	var conngroups Response
	err = json.Unmarshal(data, &conngroups)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "connected groups retrieved successfully", conngroups, http.StatusOK)
}

// AssignGroups	godoc
//
//	@Summary		Assign thing to one or more groups
//	@Description	Assign thing to one or more groups.
//	@Tags			things
//	@Produce		json
//
//	@Param			id		path	string					true	"Unique thing id."
//	@Param			Request	body	models.AssignGroupReq	true	"JSON-formatted document describing group IDs."
//
//	@Success		201		"Group(s) assigned."
//	@Failure		400		"Failed due to malformed JSON."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things/{id}/groups [post]
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		return
	}
	var response models.RespError
	_ = json.Unmarshal(data, &response)

	var withOutput string
	if response.Error != "" {
		withOutput = " with errors - " + response.Error
	} else {
		withOutput = " without errors."
	}

	errors.InfoHandler(c, "assign groups completed"+withOutput, response, resp.StatusCode)
}

// UnassignGroups	godoc
//
//	@Summary		Unassign thing from one or more groups
//	@Description	Unassign thing from one or more groups.
//	@Tags			things
//	@Produce		json
//
//	@Param			id		path	string					true	"Unique thing id."
//	@Param			Request	body	models.AssignGroupReq	true	"JSON-formatted document describing group IDs."
//
//	@Success		204		"Group(s) unassigned."
//	@Failure		400		"Failed due to malformed JSON."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/things/{id}/groups [delete]
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), http.StatusBadRequest)
		return
	}
	var response models.RespError
	_ = json.Unmarshal(data, &response)

	var withOutput string
	if response.Error != "" {
		withOutput = " with errors - " + response.Error
	} else {
		withOutput = " without errors."
	}

	errors.InfoHandler(c, "unassign groups completed"+withOutput, response, resp.StatusCode)
}
