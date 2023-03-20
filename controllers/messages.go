package controllers

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"net/http"
	"pubsubapi/errors"
	// "pubsubapi/models"
	"strings"
	"time"
)

func PublishMessage(c *gin.Context) {

	var messageReqParams struct {
		ThingID   string `form:"thingid" binding:"required"`
		ThingKey  string `form:"thingkey" binding:"required"`
		ChannelID string `form:"channelid" binding:"required"`
	}

	if err := c.ShouldBindQuery(&messageReqParams); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding query : %v", err), http.StatusBadRequest)
		return
	}

	opts := mqtt.NewClientOptions().AddBroker(":1883").SetClientID("myTestClient01").SetUsername(messageReqParams.ThingID).SetPassword(messageReqParams.ThingKey)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		errors.ErrHandler(c, fmt.Errorf("error at mqtt client connection : %v", token.Error()), http.StatusInternalServerError)
		panic(token.Error())
	}

	messages, err := c.GetRawData()
	if err != nil {
		return
	}

	err = publish(c, client, messageReqParams.ChannelID, messages)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while publishing message : %v", err), http.StatusInternalServerError)
		return
	}
	client.Disconnect(250)
	errors.InfoHandler(c, "mqtt message sent successfully", "mqtt message sent successfully", http.StatusOK)
}

func publish(c *gin.Context, client mqtt.Client, channelID string, messages []byte) error {
	channel := "channels/" + channelID + "/messages"
	token := client.Publish(channel, 2, true, standardizeSpaces(string(messages)))
	// token := client.Publish(channel, 0, true, string(messages))
	token.Wait()
	// time.Sleep(time.Second)
	return nil
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func subscribe(c *gin.Context, client mqtt.Client, channelID string) error {
	channel := "channels/" + channelID + "/messages"
	token := client.Subscribe(channel, 0, nil)
	token.Wait()
	time.Sleep(time.Second)
	return nil
}
