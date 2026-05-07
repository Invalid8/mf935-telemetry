package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const baseURL = "http://192.168.0.1"

type RouterClient struct {
	http    *http.Client
	cookies []*http.Cookie
}

func New() *RouterClient {
	return &RouterClient{
		http: &http.Client{Timeout: 8 * time.Second},
	}
}

func (c *RouterClient) GetCmds(fields []string) (map[string]string, error) {
	u := fmt.Sprintf(
		"%s/goform/goform_get_cmd_process?isTest=false&multi_data=1&cmd=%s",
		baseURL,
		url.QueryEscape(strings.Join(fields, ",")),
	)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("GetCmds: build request: %w", err)
	}

	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetCmds: do request: %w", err)
	}
	defer resp.Body.Close()

	c.storeCookies(resp.Cookies())

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetCmds: read body: %w", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("GetCmds: unmarshal: %w — body: %s", err, body)
	}

	return result, nil
}

func (c *RouterClient) Post(params map[string]string) (map[string]string, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	req, err := http.NewRequest(
		"POST",
		baseURL+"/goform/goform_set_cmd_process",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("Post: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Post: do request: %w", err)
	}
	defer resp.Body.Close()

	c.storeCookies(resp.Cookies())

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("Post: decode response: %w", err)
	}

	return result, nil
}

func (c *RouterClient) GetCookie(name string) string {
	for _, ck := range c.cookies {
		if ck.Name == name {
			return ck.Value
		}
	}
	return ""
}

func (c *RouterClient) storeCookies(incoming []*http.Cookie) {
	for _, inc := range incoming {
		found := false
		for i, existing := range c.cookies {
			if existing.Name == inc.Name {
				c.cookies[i] = inc
				found = true
				break
			}
		}
		if !found {
			c.cookies = append(c.cookies, inc)
		}
	}
}
