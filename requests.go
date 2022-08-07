package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/yaml.v2"
)

var httpReq Requests

type Requests interface {
	Get(url string) (resp *http.Response, err error)
	NewRequest(method string, url string, body io.Reader) (*http.Request, error)
	Do(req *http.Request) (*http.Response, error)
}

type httpRequests struct{}

func (h *httpRequests) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

func (h *httpRequests) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func (h *httpRequests) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

func getUrl(httpReq Requests, url string) (string, error) {
	resp, err := httpReq.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting document: %s", err.Error())
	}
	if (200 <= resp.StatusCode) && (resp.StatusCode <= 299) {
		defer resp.Body.Close()
		bodyData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error opening content: %s", err.Error())
		}

		return string(bodyData), nil
	}
	return "", fmt.Errorf("error: received status code %d", resp.StatusCode)
}

func getFromHass(httpReq Requests, hassUrl string) (string, error) {
	header := make(map[string][]string)
	header["Authorization"] = []string{fmt.Sprintf("Bearer %s", hassApiKey)}

	req, err := httpReq.NewRequest("GET", hassUrl, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("error with Home Assistant request: %s", err.Error())
	}

	req.Header = header

	res, err := httpReq.Do(req)
	if err != nil {
		return "", fmt.Errorf("error with Home Assistant response: %s", err.Error())
	}

	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		responseBody, _ := io.ReadAll(res.Body)

		jsonResponse := make(map[string]interface{})
		if err = json.Unmarshal(responseBody, &jsonResponse); err != nil {
			return "", fmt.Errorf("error decoding response: %s", err)
		}

		yamlOutput, _ := yaml.Marshal(jsonResponse)
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
