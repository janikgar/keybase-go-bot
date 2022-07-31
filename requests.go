package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

// func getFromHass(url string) {
// 	header := make(map[string][]string)
// 	header["Authorization"] = []string{fmt.Sprintf("Bearer %s", hassApiKey)}

// 	req := http.Request{
// 		Method: "GET",
// 		Header: header,
// 	}

// 	fmt.Println(req)
// }

func getIp(httpReq Requests) (string, error) {
	log.Println("looking up...")
	ipResult, err := getUrl(httpReq, "https://api.ipify.org")
	if err != nil {
		return "", err
	}
	return ipResult, nil
}
