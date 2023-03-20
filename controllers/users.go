package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mainflux/mainflux/pkg/sdk/go"
	"io"
	"net/http"
	"pubsubapi/env"
	"pubsubapi/errors"
)

func GetUsers(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	urlUser := env.Env(envUserURL, sdkUserURL)
	req, _ := http.NewRequest("GET", urlUser+"/users", nil)
	req.Header.Add("Authorization", token)
	req.Header.Add("Content-Type", "application/json")

	type pageRes struct {
		Total   uint64 `json:"total"`
		Offset  uint64 `json:"offset"`
		Limit   uint64 `json:"limit"`
		IsAdmin bool   `json:"isadmin"`
	}
	type users struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	type respUsers struct {
		pageRes
		Users []users `json:"users"`
	}

	//add query parameters
	q := req.URL.Query()
	// q.Add("status", "flagged")
	params := []string{"offset", "limit", "email"}
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
		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	var rpUsers respUsers
	if resp.StatusCode != 200 {
		req, err := http.NewRequest("GET", urlUser+"/users/profile", nil)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
			return
		}
		req.Header.Add("Authorization", token)
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
			return
		}

		// type viewUserRes struct {
		// 	ID    string `json:"id"`
		// 	Email string `json:"email"`
		// }
		// var userRes viewUserRes
		var userRes users
		json.Unmarshal(data, &userRes)
		rpUsers.Limit = 1
		rpUsers.Offset = 0
		rpUsers.Total = 1
		rpUsers.IsAdmin = false
		rpUsers.Users = append(rpUsers.Users, userRes)
		// errors.InfoHandler(c, "user displayed successfully", userRes, http.StatusOK)
		// return
	} else {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
			return
		}
		json.Unmarshal(data, &rpUsers)
		rpUsers.IsAdmin = true
	}

	errors.InfoHandler(c, "users listed successfully", rpUsers, http.StatusOK)
}

func GetUserProfile(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	urlUser := env.Env(envUserURL, sdkUserURL)
	profileURL := urlUser + "/users/profile"
	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Add("Authorization", token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
		return
	}

	type viewUserRes struct {
		ID       string                 `json:"id"`
		Email    string                 `json:"email"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}
	var userRes viewUserRes
	json.Unmarshal(data, &userRes)
	errors.InfoHandler(c, "user profile generated successfully", userRes, http.StatusOK)
}

func UpdateUser(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	// urlUser := env.Env(envUserURL, sdkUserURL)
	// profileURL := urlUser + "/users/profile"

	type User struct {
		ID       string                 `json:"id,omitempty"`
		Email    string                 `json:"email,omitempty"`
		Groups   []string               `json:"groups,omitempty"`
		Password string                 `json:"password,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	// configure the sdk payload for forwarding request
	urlUser := env.Env(envUserURL, sdkUserURL)
	configUser := sdk.Config{
		BaseURL: urlUser,
	}
	mfsdkUser := sdk.NewSDK(configUser)

	var sdkUser = sdk.User{
		ID:       user.ID,
		Email:    user.Email,
		Groups:   user.Groups,
		Password: user.Password,
		Metadata: user.Metadata,
	}

	err = mfsdkUser.UpdateUser(sdkUser, token)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
		return
	}

	errors.InfoHandler(c, "user updated successfully", "user updated successfully", http.StatusOK)
}
