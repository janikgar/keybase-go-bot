package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/yaml.v2"
)

var httpReq Requests

type Requests interface {
	Get(url string) (resp *http.Response, err error)
}

type httpRequests struct{}

func (h *httpRequests) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

func getUrl(httpReq Requests, url string) (string, error) {
	resp, err := httpReq.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting IP: %s", err.Error())
	}
	if (200 <= resp.StatusCode) || (resp.StatusCode <= 299) {
		defer resp.Body.Close()
		bodyData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error opening content: %s", err.Error())
		}

		return string(bodyData), nil
	}
	return "", fmt.Errorf("error: received status code %d", resp.StatusCode)
}

func getFromHass(hassUrl string) (string, error) {
	header := make(map[string][]string)
	header["Authorization"] = []string{fmt.Sprintf("Bearer %s", hassApiKey)}

	req, err := http.NewRequest("GET", hassUrl, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("error with Home Assistant request: %s", err.Error())
	}

	req.Header = header

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error with Home Assistant response: %s", err.Error())
	}

	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		responseBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", fmt.Errorf("error with Home Assistant response: %s", err.Error())
		}

		jsonResponse := make(map[string]interface{})
		if err = json.Unmarshal(responseBody, &jsonResponse); err != nil {
			return "", fmt.Errorf("error decoding response: %s", err)
		}

		yamlOutput, err := yaml.Marshal(jsonResponse)
		if err != nil {
			return "", fmt.Errorf("error assembling YAML: %s", err)
		}

		return fmt.Sprintf("HASS says: \n```\n%s\n```", string(yamlOutput)), nil
	}
	return res.Status, nil
}

func getIp(httpReq Requests) (string, error) {
	log.Println("looking up...")
	ipResult, err := getUrl(httpReq, "https://api.ipify.org")
	if err != nil {
		return "", err
	}
	return ipResult, nil
}
