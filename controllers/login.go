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

	// sdk "github.com/mainflux/mainflux/pkg/sdk/go"

	// "time"

	"github.com/gin-gonic/gin"
)

// LoginUser	godoc
//
//	@Summary		User authentication
//	@Description	Generates an access token when provided with proper credentials.
//	@Tags			users
//	@Produce		json
//	@Param			Request	body		models.LoginUserReq	true	"JSON-formatted document describing the user details for login"
//
//	@Success		200		{object}	models.LoginUserRes	"User authenticated."
//	@Failure		400		"Failed due to malformed JSON."
//	@Failure		500		"Unexpected server-side error occurred."
//	@Router			/login [post]
func LoginUser(c *gin.Context) {
	var user models.LoginUserReq
	if err := c.ShouldBindJSON(&user); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	// configure the sdk payload for forwarding request
	urlUser := env.Env(envUserURL, sdkUserURL)

	// var sdkUser = sdk.User{
	// 	Email:    user.Email,
	// 	Password: user.Password,
	// }

	// body, _ := json.Marshal(sdkUser)
	body, _ := json.Marshal(user)
	tokenreq, err := http.NewRequest("POST", urlUser+"/tokens", bytes.NewBuffer(body))
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	tokenreq.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	tokenresp, err := client.Do(tokenreq)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer tokenresp.Body.Close()

	data, err := io.ReadAll(tokenresp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
		return
	}

	var loginToken models.LoginUserRes
	json.Unmarshal(data, &loginToken)

	// if authorization header already exists, delete it; else new token cannot be set as the new auth header
	if c.Request.Header.Get("Authorization") != "" {
		c.Request.Header.Del("Authorization")
	}
	c.Request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", loginToken.Token))

	// checking for channels
	// configure the sdk payload for forwarding request
	url := env.Env(envThingURL, sdkThingURL)
	urlChan := url + "/channels"

	channels, _, err := findChannels(c, urlChan)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred in channels lookup : %v", err), http.StatusNotFound)
		return
	}

	tfound, ffound := false, false
	var chansToAdd []string

	// var channel models.Channel
	for _, chn := range channels {
		if chn.Name == "deviceToCloud" {
			tfound = true
			break
		}
	}
	for _, chn := range channels {
		if chn.Name == "cloudToDevice" {
			ffound = true
			break
		}
	}
	if !tfound {
		chansToAdd = append(chansToAdd, "deviceToCloud")
	}
	if !ffound {
		chansToAdd = append(chansToAdd, "cloudToDevice")
	}

	for _, chnName := range chansToAdd {
		chanData := models.ChannelReq{
			Name: chnName,
		}
		channeldata, _ := json.Marshal(chanData)
		newreq, err := http.NewRequest("POST", urlChan, bytes.NewBuffer(channeldata))
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
			return
		}
		newreq.Header.Add("Content-Type", "application/json")
		newreq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", loginToken.Token))
		// newreq.Header.Add("Authorization", loginToken.Token)

		client := &http.Client{}
		resp, err := client.Do(newreq)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		id := strings.TrimPrefix(resp.Header.Get("Location"), fmt.Sprintf("/%s/", "channels"))

		if id == "" {
			errors.ErrHandler(c, fmt.Errorf("failed to create %s channel : %v", chnName, err), http.StatusBadRequest)
		}
		fmt.Println("id:- ", id)
	}

	// checking for things
	// configure the sdk payload for forwarding request
	urlThing := url + "/things"
	things, _, err := findThings(c, urlThing)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred in channels lookup : %v", err), http.StatusNotFound)
		return
	}

	found := false
	for _, tng := range things {
		if tng.Name == "cloudEnd" {
			found = true
			break
		}
	}
	if !found {
		tngData := models.ThingReq{
			Name: "cloudEnd",
		}
		thingdata, _ := json.Marshal(tngData)
		newreq, err := http.NewRequest("POST", urlThing, bytes.NewBuffer(thingdata))
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
			return
		}
		newreq.Header.Add("Content-Type", "application/json")
		newreq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", loginToken.Token))

		client := &http.Client{}
		resp, err := client.Do(newreq)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		id := strings.TrimPrefix(resp.Header.Get("Location"), fmt.Sprintf("/%s/", "things"))

		if id == "" {
			errors.ErrHandler(c, fmt.Errorf("failed to create cloudEnd thing : %v", err), http.StatusBadRequest)
		}
		fmt.Println("id:- ", id)

		errcode, err := connect(c, id, "")
		if err != nil {
			if errcode > 0 {
				errors.ErrHandler(c, fmt.Errorf("error occurred in connecting with default channels : %v", err), errcode)
			} else {
				errors.ErrHandler(c, fmt.Errorf("error occurred in connecting with default channels : %v", err), http.StatusBadRequest)
			}
			return
		}
	}

	errors.InfoHandler(c, "logged in successfully", loginToken, http.StatusOK)
}
