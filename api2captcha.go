package api2captcha

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	BaseURL = "https://2captcha.com"
)

type (
	Request struct {
		Params map[string]string
		Files  map[string]string
	}

	Client struct {
		BaseURL          *url.URL
		ApiKey           string
		DefaultTimeout   int
		PollingInterval  int
		RecaptchaTimeout int

		httpClient *http.Client
	}

	ReCaptcha struct {
		SiteKey   string
		Url       string
		Invisible bool
		Version   string
		Action    string
		Score     float64
	}
)

var (
	ErrNetwork = errors.New("api2captcha: Network failure")
	ErrApi     = errors.New("api2captcha: API error")
	ErrTimeout = errors.New("api2captcha: Request timeout")
)

func NewClient(apiKey string) *Client {
	base, _ := url.Parse(BaseURL)
	return &Client{
		BaseURL:          base,
		ApiKey:           apiKey,
		DefaultTimeout:   120,
		PollingInterval:  5,
		RecaptchaTimeout: 600,
		httpClient:       &http.Client{},
	}
}

func (c *Client) res(req Request) (*string, error) {

	rel := &url.URL{Path: "/res.php"}
	uri := c.BaseURL.ResolveReference(rel)

	req.Params["key"] = c.ApiKey
	c.httpClient.Timeout = time.Duration(c.DefaultTimeout) * time.Second

	var resp *http.Response = nil

	values := url.Values{}
	for key, val := range req.Params {
		values.Add(key, val)
	}
	uri.RawQuery = values.Encode()

	var err error = nil
	resp, err = http.Get(uri.String())
	if err != nil {
		return nil, ErrNetwork
	}

	defer resp.Body.Close()
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	data := body.String()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrApi
	}

	if strings.HasPrefix(data, "ERROR_") {
		return nil, ErrApi
	}

	return &data, nil
}

func (c *Client) resAction(action string) (*string, error) {
	req := Request{
		Params: map[string]string{"action": action},
	}

	return c.res(req)
}

func (c *Client) Send(req Request) (string, error) {

	rel := &url.URL{Path: "/in.php"}
	uri := c.BaseURL.ResolveReference(rel)

	req.Params["key"] = c.ApiKey

	c.httpClient.Timeout = time.Duration(c.DefaultTimeout) * time.Second

	var resp *http.Response = nil
	if req.Files != nil && len(req.Files) > 0 {

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		for name, path := range req.Files {
			file, err := os.Open(path)
			if err != nil {
				return "", err
			}
			defer file.Close()

			part, err := writer.CreateFormFile(name, filepath.Base(path))
			if err != nil {
				return "", err
			}
			_, err = io.Copy(part, file)
		}

		for key, val := range req.Params {
			_ = writer.WriteField(key, val)
		}

		err := writer.Close()
		if err != nil {
			return "", err
		}

		request, err := http.NewRequest("POST", uri.String(), body)
		if err != nil {
			return "", err
		}

		request.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err = c.httpClient.Do(request)
		if err != nil {
			return "", ErrNetwork
		}
	} else {
		values := url.Values{}
		for key, val := range req.Params {
			values.Add(key, val)
		}

		var err error = nil
		resp, err = http.PostForm(uri.String(), values)
		if err != nil {
			return "", ErrNetwork
		}
	}

	defer resp.Body.Close()
	body := &bytes.Buffer{}
	_, err := body.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	data := body.String()

	if resp.StatusCode != http.StatusOK {
		return "", ErrApi
	}

	if strings.HasPrefix(data, "ERROR_") {
		return "", ErrApi
	}

	if !strings.HasPrefix(data, "OK|") {
		return "", ErrApi
	}

	return data[3:], nil
}

func (c *Client) Solve(req Request) (string, error) {

	id, err := c.Send(req)
	if err != nil {
		return "", err
	}

	timeout := c.DefaultTimeout
	if req.Params["method"] == "userrecaptcha" {
		timeout = c.RecaptchaTimeout
	}

	return c.WaitForResult(id, timeout, c.PollingInterval)
}

func (c *Client) WaitForResult(id string, timeout int, interval int) (string, error) {

	start := time.Now()
	now := start
	for now.Sub(start) < (time.Duration(timeout) * time.Second) {

		time.Sleep(time.Duration(interval) * time.Second)

		code, err := c.GetResult(id)
		if err == nil && code != nil {
			return *code, nil
		}

		// ignore network errors
		if err != nil && err != ErrNetwork {
			return "", err
		}

		now = time.Now()
	}

	return "", ErrTimeout
}

func (c *Client) GetResult(id string) (*string, error) {
	req := Request{
		Params: map[string]string{"action": "get", "id": id},
	}

	data, err := c.res(req)
	if err != nil {
		return nil, err
	}

	if *data == "CAPCHA_NOT_READY" {
		return nil, nil
	}

	if !strings.HasPrefix(*data, "OK|") {
		return nil, ErrApi
	}

	reply := (*data)[3:]
	return &reply, nil
}

func (c *Client) GetBalance() (float64, error) {
	data, err := c.resAction("getbalance")
	if err != nil {
		return 0.0, err
	}

	return strconv.ParseFloat(*data, 64)
}

func (req *Request) SetProxy(proxyType string, uri string) {
	req.Params["proxytype"] = proxyType
	req.Params["proxy"] = uri
}

func (c *ReCaptcha) ToRequest() Request {
	req := Request{
		Params: map[string]string{"method": "userrecaptcha"},
	}
	if c.SiteKey != "" {
		req.Params["googlekey"] = c.SiteKey
	}
	if c.Url != "" {
		req.Params["pageurl"] = c.Url
	}
	if c.Invisible {
		req.Params["invisible"] = "1"
	}
	if c.Version != "" {
		req.Params["version"] = c.Version
	}
	if c.Action != "" {
		req.Params["action"] = c.Action
	}
	if c.Score != 0 {
		req.Params["min_score"] = strconv.FormatFloat(c.Score, 'f', -1, 64)
	}

	return req
}
