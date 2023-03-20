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

	errors.InfoHandler(c, "group deleted successfully", "group deleted successfully", http.StatusNoContent)

}

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
		errors.ErrHandler(c, fmt.Errorf(`member could not be added to the group`), http.StatusBadRequest)
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
		errors.ErrHandler(c, fmt.Errorf("failed to process the request : %v", err), resp.StatusCode)
		return
	}

	type pageRes struct {
		Limit  uint64 `json:"limit,omitempty"`
		Offset uint64 `json:"offset,omitempty"`
		Total  uint64 `json:"total"`
		Level  uint64 `json:"level"`
		Name   string `json:"name"`
	}
	type memberPageRes struct {
		pageRes
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
