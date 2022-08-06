package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"testing"

	"github.com/janikgar/keybase-go-bot/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetUrl(t *testing.T) {
	httpReq := mocks.NewRequests(t)

	cases := []struct {
		url              string
		status           int
		errorMatchString error
	}{
		{"https://api.ipify.org", 200, nil},
		{"https://foo.bar.baz", 0, errors.New("received status code")},
	}

	body, w := io.Pipe()
	go func() {
		fmt.Fprint(w, "1.1.1.1")
		w.Close()
	}()

	for _, c := range cases {
		httpReq.On("Get", c.url).Return(&http.Response{
			StatusCode: c.status,
			Body:       body,
		}, c.errorMatchString)

		response, err := getUrl(httpReq, c.url)

		if c.errorMatchString != nil {
			require.Contains(t, err.Error(), c.errorMatchString.Error())
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

	httpReq.On("Get", "https://api.ipify.org").Return(&http.Response{
		StatusCode: 200,
		Body:       body,
	}, nil)

	ipPattern := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)
	ipAddr, err := getIp(httpReq)
	if err != nil {
		t.FailNow()
	}

	require.True(t, ipPattern.Match([]byte(ipAddr)))
}
