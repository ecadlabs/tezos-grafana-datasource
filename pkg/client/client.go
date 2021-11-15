package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ecadlabs/jtree"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model/block"
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

func (c *Client) GetBlockHeader(ctx context.Context, blockID string) (*block.Header, error) {
	req, err := c.NewGetBlockHeaderRequest(ctx, blockID)
	if err != nil {
		return nil, fmt.Errorf("getBlockHeader: %w", err)
	}
	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("getBlockHeader: %w", err)
	}
	defer res.Close()

	var v block.Header
	dec := jtree.NewDecoder(res)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("getBlockHeader: %w", err)
	}
	return &v, nil
}

func (c *Client) NewGetBlockRequest(ctx context.Context, blockID string) (*http.Request, error) {
	u := fmt.Sprintf("%s/chains/%s/blocks/%s", c.URL, c.chain(), blockID)
	return http.NewRequestWithContext(ctx, "GET", u, nil)
}

func (c *Client) GetBlock(ctx context.Context, blockID string) (*block.Block, error) {
	req, err := c.NewGetBlockRequest(ctx, blockID)
	if err != nil {
		return nil, fmt.Errorf("getBlock: %w", err)
	}
	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("getBlock: %w", err)
	}
	defer res.Close()

	var v block.Block
	dec := jtree.NewDecoder(res)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("getBlock: %w", err)
	}
	return &v, nil
}

func (c *Client) NewGetProtocolConstantsRequest(ctx context.Context) (*http.Request, error) {
	u := fmt.Sprintf("%s/chains/%s/blocks/head/context/constants", c.URL, c.chain())
	return http.NewRequestWithContext(ctx, "GET", u, nil)
}

func (c *Client) GetProtocolConstants(ctx context.Context) (*model.ProtocolConstants, error) {
	req, err := c.NewGetProtocolConstantsRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("getProtocolConstants: %w", err)
	}
	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("getProtocolConstants: %w", err)
	}
	defer res.Close()

	var v model.ProtocolConstants
	dec := jtree.NewDecoder(res)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("getProtocolConstants: %w", err)
	}
	return &v, nil
}

func (c *Client) NewGetBlockOperationsRequest(ctx context.Context, blockID string) (*http.Request, error) {
	u := fmt.Sprintf("%s/chains/%s/blocks/%s/operations", c.URL, c.chain(), blockID)
	return http.NewRequestWithContext(ctx, "GET", u, nil)
}

func (c *Client) GetBlockOperations(ctx context.Context, blockID string) (block.Operations, error) {
	req, err := c.NewGetBlockOperationsRequest(ctx, blockID)
	if err != nil {
		return nil, fmt.Errorf("getBlockOperations: %w", err)
	}
	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("getBlockOperations: %w", err)
	}
	defer res.Close()

	var v block.Operations
	dec := jtree.NewDecoder(res)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("getBlockOperations: %w", err)
	}
	return v, nil
}

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

func (c *Client) GetMinimalValidTime(ctx context.Context, blockID string, priority, power int) (time.Time, error) {
	req, err := c.NewGetMinimalValidTimeRequest(ctx, blockID, priority, power)
	if err != nil {
		return time.Time{}, fmt.Errorf("getMinimalValidTime: %w", err)
	}
	res, err := c.do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("getMinimalValidTime: %w", err)
	}
	defer res.Close()

	dec := jtree.NewDecoder(res)
	dec.DisallowUnknownFields()
	var t time.Time
	if err = dec.Decode(&t); err != nil {
		return time.Time{}, fmt.Errorf("getMinimalValidTime: %w", err)
	}
	return t, nil
}

func (c *Client) NewGetMonitorHeadsRequest(ctx context.Context) (*http.Request, error) {
	u := fmt.Sprintf("%s/monitor/heads/%s", c.URL, c.chain())
	return http.NewRequestWithContext(ctx, "GET", u, nil)
}

func (c *Client) GetMonitorHeads(ctx context.Context) (headerCh <-chan *block.ShellHeader, errorsCh <-chan error, err error) {
	req, err := c.NewGetMonitorHeadsRequest(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("getMonitorHeads: %w", err)
	}
	res, err := c.do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("getMonitorHeads: %w", err)
	}

	hdrCh := make(chan *block.ShellHeader, 100)
	errCh := make(chan error, 1)

	go func() {
		defer (func() {
			res.Close()
			close(hdrCh)
			close(errCh)
		})()

		dec := jtree.NewDecoder(res)
		dec.DisallowUnknownFields()
		for {
			var v block.ShellHeader
			if err := dec.Decode(&v); err != nil {
				if err != io.EOF {
					errCh <- err
				}
				return
			}
			select {
			case hdrCh <- &v:
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		}
	}()
	return hdrCh, errCh, nil
}
