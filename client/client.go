package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ecadlabs/tezos-grafana-datasource/storage"
)

type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (h *HTTPError) Error() string {
	return fmt.Sprintf("(%s) %s", h.Status, string(h.Body))
}

type Client struct {
	URL    string
	Chain  string
	Client *http.Client
}

func (c *Client) chain() string {
	if c.Chain != "" {
		return c.Chain
	}
	return "main"
}

func (c *Client) client() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return http.DefaultClient
}

func (c *Client) do(r *http.Request) (io.ReadCloser, error) {
	res, err := c.client().Do(r)
	if err != nil {
		return nil, err
	}

	if res.StatusCode/100 != 2 {
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, &HTTPError{
			StatusCode: res.StatusCode,
			Status:     res.Status,
			Body:       body,
		}
	}

	return res.Body, nil
}

func (c *Client) NewGetBlockHeaderRequest(ctx context.Context, blockID string) (*http.Request, error) {
	u := fmt.Sprintf("%s/chains/%s/blocks/%s/header", c.URL, c.chain(), blockID)
	return http.NewRequestWithContext(ctx, "GET", u, nil)
}

func (c *Client) GetBlockHeader(ctx context.Context, blockID string) (*storage.BlockHeader, error) {
	req, err := c.NewGetBlockHeaderRequest(ctx, blockID)
	if err != nil {
		return nil, err
	}
	res, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var v storage.BlockHeader
	dec := json.NewDecoder(res)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *Client) NewGetProtocolConstantsRequest(ctx context.Context) (*http.Request, error) {
	u := fmt.Sprintf("%s/chains/%s/blocks/head/context/constants", c.URL, c.chain())
	return http.NewRequestWithContext(ctx, "GET", u, nil)
}

func (c *Client) GetProtocolConstants(ctx context.Context) (*storage.ProtocolConstants, error) {
	req, err := c.NewGetProtocolConstantsRequest(ctx)
	if err != nil {
		return nil, err
	}
	res, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var v storage.ProtocolConstants
	dec := json.NewDecoder(res)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *Client) NewGetBlockOperationsRequest(ctx context.Context, blockID string) (*http.Request, error) {
	u := fmt.Sprintf("%s/chains/%s/blocks/%s/operations", c.URL, c.chain(), blockID)
	return http.NewRequestWithContext(ctx, "GET", u, nil)
}

func (c *Client) GetBlockOperations(ctx context.Context, blockID string) (storage.BlockOperations, error) {
	req, err := c.NewGetBlockOperationsRequest(ctx, blockID)
	if err != nil {
		return nil, err
	}
	res, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var v storage.BlockOperations
	dec := json.NewDecoder(res)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

/*
func newJSONRequest(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	var rd io.Reader
	if body != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
		rd = &buf
	}
	req, err := http.NewRequestWithContext(ctx, method, url, rd)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	return req, nil
}
*/

func (c *Client) NewGetMinimalValidTimeRequest(ctx context.Context, blockID string, priority, power int) (*http.Request, error) {
	u, err := url.Parse(fmt.Sprintf("%s/chains/%s/blocks/%s/minimal_valid_time", c.URL, c.chain(), blockID))
	if err != nil {
		return nil, err
	}
	u.RawQuery = url.Values{
		"priority":        []string{strconv.FormatInt(int64(priority), 10)},
		"endorsing_power": []string{strconv.FormatInt(int64(power), 10)},
	}.Encode()
	return http.NewRequestWithContext(ctx, "GET", u.String(), nil)
}

func (c *Client) GetMinimalValidTime(ctx context.Context, blockID string, priority, power int) (t time.Time, err error) {
	req, err := c.NewGetMinimalValidTimeRequest(ctx, blockID, priority, power)
	if err != nil {
		return
	}
	res, err := c.do(req)
	if err != nil {
		return
	}
	defer res.Close()

	dec := json.NewDecoder(res)
	dec.DisallowUnknownFields()
	if err = dec.Decode(&t); err != nil {
		return
	}
	return
}
