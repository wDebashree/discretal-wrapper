package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"io"
	"net/http"
	"pubsubapi/env"
	"pubsubapi/errors"
	"pubsubapi/models"

	// "strings"

	"github.com/gin-gonic/gin"
)

func AddMaps(c *gin.Context) {
	// to test whether AddMaps requirements are met or not
	var testMaps []models.AddMapReq
	if err := c.ShouldBindJSON(&testMaps); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}
	mapdata, _ := json.Marshal(testMaps)

	urlMap := env.Env(envThingURL, sdkThingURL)
	urlMap = urlMap + "/maps"

	resp, err := httpReq(c, "POST", urlMap, mapdata, nil)
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

	var response models.RespError
	_ = json.Unmarshal(data, &response)

	errors.InfoHandler(c, "maps addition operation completed"+withOutput(response.Error), response, resp.StatusCode)
}

func GetMap(c *gin.Context) {
	// Retrieve the params for thing details
	var item models.ItemId
	if err := c.ShouldBindUri(&item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	urlMap := env.Env(envThingURL, sdkThingURL)
	urlMap = urlMap + "/maps/" + item.Id

	resp, err := httpReq(c, "GET", urlMap, nil, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), resp.StatusCode)
		return
	}

	// type Response struct {
	// 	*models.ViewMapRes `json:",omitempty"`
	// 	Error              string `json:"error,omitempty"`
	// }

	// var response Response
	// _ = json.Unmarshal(data, &response)

	type Response struct {
		*models.ViewMapRes `json:",omitempty"`
		Error              string `json:"error,omitempty"`
	}

	var response Response
	_ = json.Unmarshal(data, &response)

	readerUrl := env.Env(envReaderURL, sdkReaderURL)
	newURL := readerUrl + "/channels/data/" + item.Id
	resp, err = httpReq(c, "GET", newURL, nil, nil)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), resp.StatusCode)
		return
	}
	fmt.Println("==> ", string(data))

	var lst models.LastSentTime
	_ = json.Unmarshal(data, &lst)
	fmt.Printf("--> %+v \n", lst)

	lo := time.Unix(int64(lst.LatestTimes[response.ThingID]), 0)
	response.Status = !(time.Since(lo) > time.Duration(300)*time.Second)
	response.LastOnline = lo.UTC().Format("Jan 2, 2006, 3:04:05 PM")

	errors.InfoHandler(c, "map details retrieval done"+withOutput(response.Error), response, resp.StatusCode)
}

func GetMaps(c *gin.Context) {
	urlMap := env.Env(envThingURL, sdkThingURL)
	urlMap = urlMap + "/maps"

	var status string
	if v, ok := c.GetQuery("status"); ok {
		status = v
	}

	params := make(map[string]string)
	paramarr := []string{"limit", "offset", "name", "order", "dir", "email"}
	for _, prms := range paramarr {
		val, ok := c.GetQuery(prms)
		if ok {
			params[prms] = val
		}
	}

	resp, err := httpReq(c, "GET", urlMap, nil, params)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), resp.StatusCode)
		return
	}

	type Response struct {
		*models.MapsPageRes `json:",omitempty"`
		Error               string `json:"error,omitempty"`
	}

	// var response Response
	// _ = json.Unmarshal(data, &response)
	var maps Response
	_ = json.Unmarshal(data, &maps)

	// var maps models.MapsPageRes
	// _ = json.Unmarshal(data, &maps)

	readerUrl := env.Env(envReaderURL, sdkReaderURL)
	newURL := readerUrl + "/channels/data"
	resp, err = httpReq(c, "GET", newURL, nil, params)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("error occurred while reading response : %v", err), resp.StatusCode)
		return
	}
	var lst models.LastSentTime
	_ = json.Unmarshal(data, &lst)

	var response models.MapsPageRes
	var totalDevices uint64

	fmt.Println("total maps :- ", len(maps.Maps))
	for _, mp := range maps.Maps {
		// fmt.Println(mp.ThingID, " - ", time.Unix(int64(lst.LatestTimes[mp.ThingID]), 0).Format("Jan 2, 2006, 3:04:05 PM"), " - ", time.Now().Format("Jan 2, 2006, 3:04:05 PM"))
		// fmt.Println(mp.ThingID, " - ", time.Unix(int64(lst.LatestTimes[mp.ThingID]), 0).UTC().Format("Jan 2, 2006, 3:04:05 PM"), " - ", time.Now().UTC().Format("Jan 2, 2006, 3:04:05 PM"))

		lo := time.Unix(int64(lst.LatestTimes[mp.ThingID]), 0)
		mp.Status = !(time.Since(lo) > time.Duration(300)*time.Second)
		// mp.LastOnline = lo.String()
		mp.LastOnline = lo.UTC().Format("Jan 2, 2006, 3:04:05 PM")

		switch status {
		//	to list online devices
		case "online":
			if mp.Status {
				response.Maps = append(response.Maps, mp)
				totalDevices++
			}

		//	to list offline devices
		case "offline":
			if !mp.Status {
				response.Maps = append(response.Maps, mp)
				totalDevices++
			}

		//	to list all devices
		default:
			response.Maps = append(response.Maps, mp)
			totalDevices++
		}
	}
	response.Dir = maps.Dir
	response.Limit = maps.Limit
	response.Offset = maps.Offset
	response.Order = maps.Order
	// response.Total = maps.Total
	response.Total = totalDevices

	errors.InfoHandler(c, "map details retrieval done"+withOutput(maps.Error), response, resp.StatusCode)
}

func UpdateMap(c *gin.Context) {
	// Retrieve the params for thing details
	var item models.ItemId
	if err := c.ShouldBindUri(&item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	var updatemap models.UpdateMapReq
	if err := c.ShouldBindJSON(&updatemap); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}
	mapdata, _ := json.Marshal(updatemap)

	urlMap := env.Env(envThingURL, sdkThingURL)
	urlMap = urlMap + "/maps/" + item.Id

	resp, err := httpReq(c, "PUT", urlMap, mapdata, nil)
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

	var response models.RespError
	_ = json.Unmarshal(data, &response)

	errors.InfoHandler(c, "map updation completed"+withOutput(response.Error), response, resp.StatusCode)
}

func RemoveMap(c *gin.Context) {
	// Retrieve the params for thing details
	var item models.ItemId
	if err := c.ShouldBindUri(&item); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding uri : %v", err), http.StatusBadRequest)
		return
	}

	urlMap := env.Env(envThingURL, sdkThingURL)
	urlMap = urlMap + "/maps/" + item.Id

	resp, err := httpReq(c, "DELETE", urlMap, nil, nil)
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

	var response models.RespError
	_ = json.Unmarshal(data, &response)

	errors.InfoHandler(c, "map data removal completed"+withOutput(response.Error), response, resp.StatusCode)
}
