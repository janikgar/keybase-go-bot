package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/janikgar/keybase-go-bot/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetUrl(t *testing.T) {
	cases := []struct {
		url                  string
		status               int
		expectedRequestError error
		expectedHTTPError    error
	}{
		{
			"https://api.ipify.org",
			200,
			nil,
			nil,
		},
		{
			"https://api.ipify.org",
			200,
			errors.New("requestError"),
			nil,
		},
		{
			"https://foo.bar.baz",
			404,
			nil,
			errors.New("error: received status code 404"),
		},
	}

	body, w := io.Pipe()
	go func() {
		fmt.Fprint(w, "1.1.1.1")
		w.Close()
	}()

	for _, c := range cases {
		httpReq := mocks.NewRequests(t)

		httpReq.On("Get", c.url).Return(&http.Response{
			StatusCode: c.status,
			Body:       body,
		}, c.expectedRequestError)

		response, err := getUrl(httpReq, c.url)

		if c.status != 200 {
			require.Equal(t, err.Error(), c.expectedHTTPError.Error())
			require.Equal(t, "", response)
		} else if c.expectedRequestError != nil {
			require.Contains(t, err.Error(), c.expectedRequestError.Error())
			require.Equal(t, "", response)
		} else {
			require.NotEmpty(t, response)
		}
	}
}

func TestGetIp(t *testing.T) {
	httpReq := mocks.NewRequests(t)

	body, w := io.Pipe()
	go func() {
		fmt.Fprint(w, "1.1.1.1")
		w.Close()
	}()

	cases := []struct {
		httpError error
	}{
		{nil},
		{errors.New("no such domain")},
	}

	for _, c := range cases {
		httpReq.On("Get", "https://api.ipify.org").Return(&http.Response{
			StatusCode: 200,
			Body:       body,
		}, c.httpError)

		ipPattern := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)
		ipAddr, err := getIp(httpReq)
		if c.httpError != nil {
			require.NotNil(t, err)
		} else {
			require.True(t, ipPattern.Match([]byte(ipAddr)))
		}

	}

}

func TestGetFromHass(t *testing.T) {
	hassUrl := "http://home-assistant.home.lan:8123/api/"
	hassUrlAsUrl, _ := url.Parse(hassUrl)

	hassReq := &http.Request{
		Method: "GET",
		URL:    hassUrlAsUrl,
	}

	body, bodyWrite := io.Pipe()
	go func() {
		fmt.Fprintf(bodyWrite, `{"hello":"world"}`)
		bodyWrite.Close()
	}()

	cases := []struct {
		expectedOutput        string
		expectedRequestError  error
		expectedResponseError error
		truncateResponse      bool
		statusCode            int
		expectedFinalError    error
	}{
		{
			"",
			nil,
			errors.New("responseError"),
			false,
			200,
			errors.New("error with Home Assistant response: responseError"),
		},
		{
			"",
			errors.New("requestError"),
			nil,
			false,
			200,
			errors.New("error with Home Assistant request: requestError"),
		},
		{
			"",
			nil,
			nil,
			true,
			200,
			errors.New("error decoding response: unexpected end of JSON input"),
		},
		{
			"",
			nil,
			nil,
			false,
			200,
			errors.New(""),
		},
		{
			"404 error",
			nil,
			nil,
			false,
			404,
			nil,
		},
	}

	for _, c := range cases {
		httpReq := mocks.NewRequests(t)
		httpReq.On("NewRequest", "GET", hassUrl, http.NoBody).Return(
			hassReq,
			c.expectedRequestError,
		).Maybe()

		if c.truncateResponse {
			io.ReadAll(body)
		}

		statusText := "200 OK"
		if c.statusCode != 200 {
			statusText = fmt.Sprintf("%d error", c.statusCode)
		}

		httpReq.On("Do", hassReq).Return(
			&http.Response{
				Status:     statusText,
				StatusCode: c.statusCode,
				Body:       body,
			},
			c.expectedResponseError,
		).Maybe()

		output, err := getFromHass(httpReq, hassUrl)
		require.Equal(t, c.expectedOutput, output)
		if c.expectedFinalError != nil {
			require.Contains(t, err.Error(), c.expectedFinalError.Error())
		} else {
			require.Nil(t, err)
		}
	}
}
