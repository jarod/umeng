package upush

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// RawClient 友盟push接口简单封装
type RawClient struct {
	appKey          string
	appMasterSecret string
	gatewayURL      string
	hc              *http.Client
}

// Client 友盟push接口易用封装
type Client struct {
	rc RawClient
}

// NewRawClient .
func NewRawClient(appKey, appMasterSecret string) *RawClient {
	return &RawClient{
		appKey:          appKey,
		appMasterSecret: appMasterSecret,
		gatewayURL:      httpsGatewayURL,
		hc:              &http.Client{},
	}
}

// NewClient .
func NewClient(appKey, appMasterSecret string) *Client {
	return &Client{
		rc: RawClient{
			appKey:          appKey,
			appMasterSecret: appMasterSecret,
			gatewayURL:      httpsGatewayURL,
			hc:              &http.Client{},
		},
	}
}

// SetHTTPS set use https api gateway or not. default true
func (c *RawClient) SetHTTPS(b bool) {
	if b {
		c.gatewayURL = httpsGatewayURL
	} else {
		c.gatewayURL = httpGatewayURL
	}
}

// SetHTTPS set use https api gateway or not. default true
func (c *Client) SetHTTPS(b bool) {
	c.rc.SetHTTPS(b)
}

func (c *RawClient) sendRequest(ctx context.Context, method, url string, p interface{}, result Result) error {
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	mysign := Sign(method, url, body, c.appMasterSecret)
	url = url + "?sign=" + mysign
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	res, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode > 400 { //  not include 400
		return errors.New(res.Status)
	}
	respData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	err = json.Unmarshal(respData, result)
	if err != nil {
		return err
	}
	return result.Error()
}
