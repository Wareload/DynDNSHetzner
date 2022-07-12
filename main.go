package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {
	log("DynDNSHetzner Client started")
	curIPv4 := "undefined" //default value, changes after first iteration
	for true {
		newIPv4, err := getPublicIPv4()
		if err != nil { //handles error while getting public ipv4
			log(err.Error())
			continue
		}
		if !(curIPv4 == newIPv4) {
			go updateHetznerRecord(&curIPv4, newIPv4)
		} //else skips
		time.Sleep(1 * time.Minute)
	}
}

// get public ipv4 by using the ipify api
func getPublicIPv4() (string, error) {
	ipifyUrl := "https://api.ipify.org?format=text"
	resp, err := http.Get(ipifyUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(ip), nil
}

//update Hetzner record with current ipv4, if the ipv4 address changes
func updateHetznerRecord(curIPv4 *string, newIPv4 string) {
	hetznerApiUrl := "https://dns.hetzner.com/api/v1/records/"
	apiKey, hasApiKey := os.LookupEnv("APIKEY")
	name, hasName := os.LookupEnv("NAME")
	record, hasRecord := os.LookupEnv("RECORD")
	zone, hasZone := os.LookupEnv("ZONE")
	if hasApiKey && hasName && hasRecord && hasZone {
		bodyObj := &body{
			ZoneId: zone,
			Type:   "A",
			Name:   name,
			Value:  newIPv4,
		}
		bodyJson, err := json.Marshal(bodyObj)
		if err != nil {
			log(err.Error())
			return
		}
		client := http.Client{}
		req, err := http.NewRequest(http.MethodPut, hetznerApiUrl+record, bytes.NewBuffer(bodyJson))
		if err != nil {
			log(err.Error())
			return
		}
		req.Header.Set("Auth-API-Token", apiKey)
		res, err := client.Do(req)
		if err != nil {
			log(err.Error())
			return
		}
		if res.StatusCode == 200 {
			log("Updated from " + *curIPv4 + " to " + newIPv4)
			*curIPv4 = newIPv4
		} else {
			log("Error while updating, code: " + res.Status)
		}
	}
}

//logs formatted to console with current timestamp
func log(msg string) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05") + "\t" + msg)
}

//structure for http request body
type body struct {
	ZoneId string `json:"zone_id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}
