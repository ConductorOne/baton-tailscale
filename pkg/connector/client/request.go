package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/tailscale/hujson"
	"go.uber.org/zap"
)

const (
	apiPathACL  = "/tailnet/%s/acl"
	baseUrl     = "https://api.tailscale.com/api/v2"
	contentType = "application/hujson"
	groupPrefix = "group:"
	ifMatch     = "If-Match"
)

// WithHujsonResponse TODO(marcos): Move this helper to baton-sdk.
func WithHujsonResponse(response interface{}) uhttp.DoOption {
	return func(resp *uhttp.WrapperResponse) error {
		responseContentType := resp.Header.Get("content-type")
		if !uhttp.IsJSONContentType(responseContentType) {
			return fmt.Errorf("unexpected content type for json response: %s", responseContentType)
		}
		if response == nil && len(resp.Body) == 0 {
			return nil
		}

		// TODO(marcos): The original parser would only read up to 256 kB.
		// Should we also error when that happens?
		parsed, err := hujson.Parse(resp.Body)
		if err != nil {
			return fmt.Errorf("tailscale-connector: %w", err)
		}
		response = parsed
		return nil
	}
}

// WithBody TODO(marcos): Clean this up and move it to baton-sdk.
func WithBody(body io.Reader) uhttp.RequestOption {
	return func() (io.ReadWriter, map[string]string, error) {
		var temp []byte
		_, err := body.Read(temp)
		if err != nil {
			return nil, nil, err
		}
		buffer := new(bytes.Buffer)
		return buffer, nil, nil
	}
}

func (c *Client) getACLUrl() (*url.URL, error) {
	return url.Parse(baseUrl + fmt.Sprintf(apiPathACL, c.tailnet))
}

func (c *Client) get(ctx context.Context, target interface{}) (
	string,
	*v2.RateLimitDescription,
	error,
) {
	return c.makeRequest(ctx, http.MethodGet, target, nil, "")
}

func (c *Client) post(ctx context.Context, requestBody io.Reader, etag string) (
	*v2.RateLimitDescription,
	error,
) {
	_, ratelimitData, err := c.makeRequest(
		ctx,
		http.MethodPost,
		nil,
		requestBody,
		etag,
	)
	return ratelimitData, err
}

func (c *Client) makeRequest(
	ctx context.Context,
	method string,
	target interface{},
	requestBody io.Reader,
	etag string,
) (
	string,
	*v2.RateLimitDescription,
	error,
) {
	logger := ctxzap.Extract(ctx)
	path, err := c.getACLUrl()
	if err != nil {
		return "", nil, err
	}

	options := []uhttp.RequestOption{
		uhttp.WithAccept(contentType),
		uhttp.WithContentType(contentType),
	}
	if requestBody != nil {
		options = append(options, WithBody(requestBody))
	}
	if etag != "" {
		options = append(options, uhttp.WithHeader(ifMatch, etag))
	}

	request, err := c.wrapper.NewRequest(
		ctx,
		method,
		path,
		options...,
	)
	if err != nil {
		return "", nil, err
	}

	request.SetBasicAuth(c.apiKey, "")

	ratelimitData := v2.RateLimitDescription{}
	response, err := c.wrapper.Do(
		request,
		uhttp.WithRatelimitData(&ratelimitData),
		WithHujsonResponse(target),
	)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusPreconditionFailed {
			logger.Error(
				"tailscale: precondition failed, the etag didn't match",
				zap.Int("status_code", response.StatusCode),
			)
			return "", &ratelimitData, errors.New("error updating tailscale grants: precondition failed")
		}

		return "", &ratelimitData, err
	}

	defer response.Body.Close()

	etag = response.Header.Get("etag")
	return etag, &ratelimitData, nil
}
