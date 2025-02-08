package rj_http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

/*
wrap http.Client
*/
type Client struct {
	httpClient *http.Client
}

type ClientOptions func(*Client)

func WithTimeout(timeout time.Duration) ClientOptions {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func NewHttpClient(options ...ClientOptions) *Client {
	c := &Client{
		httpClient: &http.Client{},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

type RequestOption func(*http.Request)

/*
send json req
*/
func (c *Client) doRequest(ctx context.Context, url string, method string, body any, requestOptions ...RequestOption) (response []byte, err error) {
	if c == nil {
		return nil, fmt.Errorf("http client not init")
	}

	var bodyReader io.Reader

	if body != nil {
		bodyByte, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyByte)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if bodyReader != nil {
		req.Header.Set("content-type", "application/json")
	}

	for _, option := range requestOptions {
		option(req)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *Client) Get(ctx context.Context, url string, requestOptions ...RequestOption) (response []byte, err error) {
	if c == nil {
		return nil, fmt.Errorf("http client not init")
	}
	return c.doRequest(ctx, url, http.MethodGet, nil, requestOptions...)
}

func (c *Client) Post(ctx context.Context, url string, body any, requestOptions ...RequestOption) (response []byte, err error) {
	if c == nil {
		return nil, fmt.Errorf("http client not init")
	}
	return c.doRequest(ctx, url, http.MethodPost, body, requestOptions...)
}

func UrlWithParams(baseUrl string, parms map[string]string) (string, error) {
	base, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}

	q := base.Query()
	for k, v := range parms {
		q.Add(k, v)
	}

	base.RawQuery = q.Encode()

	return base.String(), nil
}
