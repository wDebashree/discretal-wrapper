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
	"strings"

	"github.com/gin-gonic/gin"
)

// CreateGroup	godoc
//
//	@Summary		Adds new group
//	@Description	Adds new group that will be owned by user identified using the provided access token.
//	@Tags			groups
//	@Produce		json
//	@Param			Request	body		models.GroupReq					true	"JSON-formatted document describing the new group."
//
//	@Success		201		{object}	controllers.CreateGroup.Resp	"Group created."
//	@Failure		400		"Failed due to malformed JSON."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/groups [post]
func CreateGroup(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// to test whether CreateGroup requirements are met or not
	var testGroup models.GroupReq
	if err := c.ShouldBindJSON(&testGroup); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	urlGroup := env.Env(envGroupURL, sdkGroupURL)
	urlGroup = urlGroup + "/groups"

	// To check whether groupname already exists or not
	groupsList, errcode, err := findGroups(c, urlGroup)
	if err != nil {
		if errcode > 0 {
			errors.ErrHandler(c, fmt.Errorf("error occurred while checking naming conflict : %v", err), errcode)
		} else {
			errors.ErrHandler(c, fmt.Errorf("error occurred while checking naming conflict : %v", err), http.StatusBadRequest)
		}
		return
	}

	var group models.GroupRes
	for _, grp := range groupsList {
		if grp.Name == testGroup.Name {
			group = grp
			break
		}
	}

	if group.ID != "" {
		errors.InfoHandler(c, "group name already exists", group, http.StatusAlreadyReported)
		return
	}

	groupdata, _ := json.Marshal(testGroup)
	req, err := http.NewRequest("POST", urlGroup, bytes.NewBuffer(groupdata))
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
	groupid := strings.TrimPrefix(resp.Header.Get("Location"), fmt.Sprintf("/%s/", "groups"))

	type Resp struct {
		ID string
	}
	groupresp := Resp{
		ID: groupid,
	}

	errors.InfoHandler(c, "group created successfully", groupresp, http.StatusCreated)
}

// GetGroups	godoc
//
//	@Summary		Retrieves groups
//	@Description	Retrieves a list of groups. Due to performance concerns, data is retrieved in subsets. The API things must ensure that the entire dataset is consumed either by making subsequent requests, or by increasing the subset size of the initial request.
//	@Tags			groups
//	@Produce		json
//
//	@Param			limit	query		integer				false	"Size of the subset to retrieve."
//	@Param			offset	query		integer				false	"Number of items to skip during retrieval."
//
//	@Success		200		{object}	models.GroupPageRes	"Data retrieved."
//	@Failure		400		"Failed due to malformed query parameters."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/groups [get]
func GetGroups(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	urlGroup := env.Env(envGroupURL, sdkGroupURL)
	urlGroup = urlGroup + "/groups"

	req, err := http.NewRequest("GET", urlGroup, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	//add query parameters
	q := req.URL.Query()
	params := []string{"offset", "limit"}
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

	var allGroups models.GroupPageRes
	err = json.Unmarshal(data, &allGroups)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "things retrieved successfully", allGroups, http.StatusOK)

}

// GetGroup	godoc
//
//	@Summary		Retrieves group info
//	@Description	Retrieves the details of a group
//	@Tags			groups
//	@Produce		json
//
//	@Param			id	path		string			true	"Unique group id."
//
//	@Success		200	{object}	models.GroupRes	"Data retrieved."
//	@Failure		400	"Failed due to malformed group's ID."
//	@Failure		404	"Group does not exist."
//	@Failure		401	"Missing or invalid access token provided."
//	@Failure		500	"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/groups/{id} [get]
func GetGroup(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for group details
	var Group models.GroupId
	if err := c.ShouldBindUri(&Group); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	urlGroup := env.Env(envGroupURL, sdkGroupURL)
	urlGroup = urlGroup + "/groups/" + Group.ID
	req, err := http.NewRequest("GET", urlGroup, nil)
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
		errors.ErrHandler(c, fmt.Errorf(`invalid group ID`), http.StatusBadRequest)
		return
	}

	var groupres models.GroupRes
	err = json.Unmarshal(data, &groupres)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "group retrieved successfully", groupres, http.StatusOK)
}

// DeleteGroup	godoc
//
//	@Summary		Removes a group
//	@Description	Removes a group. The service will ensure that the subscribed group relation is deleted as well.
//	@Tags			groups
//	@Produce		json
//
//	@Param			id	path	string	true	"Unique group id."
//
//	@Success		204	"Group removed."
//	@Failure		400	"Failed due to malformed group's ID."
//	@Failure		401	"Missing or invalid access token provided."
//	@Failure		500	"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/channels/{id} [delete]
func DeleteGroup(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for group details
	var Group models.GroupId
	if err := c.ShouldBindUri(&Group); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	// configure the sdk payload for forwarding request
	urlGroup := env.Env(envGroupURL, sdkGroupURL)
	urlGroup = urlGroup + "/groups/" + Group.ID

	//delete the channel by its ID
	req, err := http.NewRequest("DELETE", urlGroup, nil)
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
		data, _ := io.ReadAll(resp.Body)
		type Error struct {
			Err string `json:"error"`
		}
		var err Error
		er := json.Unmarshal(data, &err)
		if er != nil {
			errors.ErrHandler(c, fmt.Errorf(`group could not be deleted`), resp.StatusCode)
			return
		}
		fmt.Println(err.Err)
		errors.ErrHandler(c, fmt.Errorf(err.Err), resp.StatusCode)
		return
	}

	errors.InfoHandler(c, "group deleted successfully", "group deleted successfully", http.StatusNoContent)

}

// AddMembers	godoc
//
//	@Summary		Assign one or more things to the group.
//	@Description	Assign one or more things to the group.
//	@Tags			groups
//	@Produce		json
//
//	@Param			groupID	path	string				true	"Unique group id."
//	@Param			Request	body	models.AssignReq	true	"JSON-formatted document describing group IDs."
//
//	@Success		201		"Member(s) assigned."
//	@Failure		400		"Failed due to malformed JSON."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/groups/{groupID}/members [post]
func AddMembers(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for group and member details
	var group models.GroupId
	if err := c.ShouldBindUri(&group); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}
	var member models.AssignReq
	if err := c.ShouldBindJSON(&member); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	urlGroup := env.Env(envGroupURL, sdkGroupURL)
	urlGroup = urlGroup + "/groups/" + group.ID + "/members"

	groupdata, _ := json.Marshal(member)
	req, err := http.NewRequest("POST", urlGroup, bytes.NewBuffer(groupdata))
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

	if (resp.StatusCode != 200) && (resp.StatusCode != 201) {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		}
		var resperr models.RespError
		json.Unmarshal(data, &resperr)
		errors.ErrHandler(c, fmt.Errorf("%s", resperr.Error), http.StatusBadRequest)
		return
	}

	// // To update related metadata
	// var updateURL string
	// if member.Type == "things" {
	// 	updateURL = env.Env(envThingURL, sdkThingURL)
	// 	updateURL = updateURL + "/things/"
	// } else if member.Type == "users" {
	// 	updateURL = env.Env(envUserURL, sdkUserURL)
	// 	updateURL = updateURL + "/users/"
	// }

	// var addmember models.ThingReq
	// addmember.Metadata = make(map[string]interface{}, 1)
	// addmember.Metadata["addgroup"] = group.ID
	// addmemberdata, _ := json.Marshal(addmember)
	// fmt.Printf("updatemember --> %+v \n", addmember)

	// for _, memberid := range member.Members {
	// 	newURL := updateURL + memberid
	// 	req, err = http.NewRequest("PUT", newURL, bytes.NewBuffer(addmemberdata))
	// 	if err != nil {
	// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	req.Header.Add("Content-Type", "application/json")
	// 	req.Header.Add("Authorization", token)
	// 	client := &http.Client{}
	// 	_, err = client.Do(req)
	// 	if err != nil {
	// 		fmt.Println("err :: ", err)
	// 		errors.ErrHandler(c, fmt.Errorf("_request could not be processed by server : %v", err), http.StatusInternalServerError)
	// 	}
	// }

	errors.InfoHandler(c, "member added to group successfully", "member added to group successfully", http.StatusNoContent)
}

// DeleteMembers	godoc
//
//	@Summary		Remove one or more things from the group.
//	@Description	Remove one or more things from the group.
//	@Tags			groups
//	@Produce		json
//
//	@Param			id		path	string				true	"Unique group id."
//	@Param			Request	body	models.AssignReq	true	"JSON-formatted document describing group IDs."
//
//	@Success		204		"Group(s) unassigned."
//	@Failure		400		"Failed due to malformed JSON."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/groups/{groupID}/members [delete]
func DeleteMembers(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for group and member details
	var group models.GroupId
	if err := c.ShouldBindUri(&group); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}
	var member models.AssignReq
	if err := c.ShouldBindJSON(&member); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	urlGroup := env.Env(envGroupURL, sdkGroupURL)
	urlGroup = urlGroup + "/groups/" + group.ID + "/members"

	groupdata, _ := json.Marshal(member)
	req, err := http.NewRequest("DELETE", urlGroup, bytes.NewBuffer(groupdata))
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

	if (resp.StatusCode != 200) && (resp.StatusCode != 204) {
		errors.ErrHandler(c, fmt.Errorf(`member could not be removed from the group`), http.StatusBadRequest)
		return
	}

	// // To update related metadata

	// var updateURL string
	// if member.Type == "things" {
	// 	updateURL = env.Env(envThingURL, sdkThingURL)
	// 	updateURL = updateURL + "/things/"
	// } else if member.Type == "users" {
	// 	updateURL = env.Env(envUserURL, sdkUserURL)
	// 	updateURL = updateURL + "/users/"
	// }

	// var removemember models.ThingReq
	// removemember.Metadata = make(map[string]interface{}, 1)
	// removemember.Metadata["removegroup"] = group.ID
	// removememberdata, _ := json.Marshal(removemember)

	// for _, memberid := range member.Members {
	// 	newURL := updateURL + memberid
	// 	req, err = http.NewRequest("PUT", newURL, bytes.NewBuffer(removememberdata))
	// 	if err != nil {
	// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	req.Header.Add("Content-Type", "application/json")
	// 	req.Header.Add("Authorization", token)
	// 	client := &http.Client{}
	// 	_, err = client.Do(req)
	// 	if err != nil {
	// 		fmt.Println("err :: ", err)
	// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	// 	}
	// }

	errors.InfoHandler(c, "member removed from group successfully", "member removed from group successfully", http.StatusNoContent)
}

// GetMembers	godoc
//
//	@Summary		Retrieves connected members
//	@Description	Retrieves a list of members which belong to the group.
//	@Tags			groups
//	@Produce		json
//
//	@Param			id		path		string									true	"Unique thing id."
//	@Param			limit	query		integer									false	"Size of the subset to retrieve."
//	@Param			offset	query		integer									false	"Number of items to skip during retrieval."
//	@Param			type	query		string									true	"Member is of type users or things."
//
//	@Success		200		{object}	controllers.GetMembers.memberPageRes	"Data retrieved."
//	@Failure		400		"Failed due to malformed query parameters."
//	@Failure		401		"Missing or invalid access token provided."
//	@Failure		500		"Unexpected server-side error occurred."
//
//	@Security		BearerAuth
//
//	@Router			/groups/{groupID}/members [get]
func GetMembers(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// Retrieve the params for group details
	var Group models.GroupId
	if err := c.ShouldBindUri(&Group); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	urlGroup := env.Env(envGroupURL, sdkGroupURL)
	urlGroup = urlGroup + "/groups/" + Group.ID + "/members"

	req, err := http.NewRequest("GET", urlGroup, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	//add query parameters
	q := req.URL.Query()
	params := []string{"offset", "limit", "type"}
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

	if (resp.StatusCode != 200) && (resp.StatusCode != 201) {
		var err models.RespError
		json.Unmarshal(data, &err)
		errors.ErrHandler(c, fmt.Errorf("failed to process the request : %v", err.Error), resp.StatusCode)
		return
	}

	// type pageRes struct {
	// 	Limit  uint64 `json:"limit,omitempty"`
	// 	Offset uint64 `json:"offset,omitempty"`
	// 	Total  uint64 `json:"total"`
	// 	Level  uint64 `json:"level"`
	// 	Name   string `json:"name"`
	// }
	type memberPageRes struct {
		// pageRes
		models.GrpPageRes
		Type    string   `json:"type"`
		Members []string `json:"members"`
	}
	var membersres memberPageRes
	err = json.Unmarshal(data, &membersres)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while unmarshalling response : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "group's members retrieved successfully", membersres, http.StatusOK)
}
