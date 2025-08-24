package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/segmentio/ksuid"
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

func (c *Client) getACLUrl() (*url.URL, error) {
	aclUrl, err := url.Parse(baseUrl + fmt.Sprintf(apiPathACL, c.tailnet))
	if err != nil {
		return nil, fmt.Errorf("tailscale-connector: error parsing acl url: %w", err)
	}
	uid := ksuid.New().String()
	q := aclUrl.Query()
	q.Set("baton-disable-cache", uid)
	aclUrl.RawQuery = q.Encode()
	return aclUrl, nil
}

func (c *Client) get(ctx context.Context) (
	*hujson.Value,
	string,
	*v2.RateLimitDescription,
	error,
) {
	return c.makeRequest(ctx, http.MethodGet, nil, "")
}

func (c *Client) post(ctx context.Context, requestBody []byte, etag string) (
	*hujson.Value,
	*v2.RateLimitDescription,
	error,
) {
	value, _, ratelimitData, err := c.makeRequest(
		ctx,
		http.MethodPost,
		requestBody,
		etag,
	)
	return value, ratelimitData, err
}

func (c *Client) makeRequest(
	ctx context.Context,
	method string,
	requestBody []byte,
	etag string,
) (
	*hujson.Value,
	string,
	*v2.RateLimitDescription,
	error,
) {
	logger := ctxzap.Extract(ctx)
	path, err := c.getACLUrl()
	if err != nil {
		return nil, "", nil, err
	}

	options := []uhttp.RequestOption{
		uhttp.WithAccept(contentType),
		uhttp.WithContentType(contentType),
	}

	if requestBody != nil {
		options = append(options, uhttp.WithBody(requestBody))
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
		return nil, "", nil, err
	}

	request.SetBasicAuth(c.apiKey, "")

	ratelimitData := v2.RateLimitDescription{}
	response, err := c.wrapper.Do(
		request,
		uhttp.WithRatelimitData(&ratelimitData),
	)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusPreconditionFailed {
			logger.Error(
				"tailscale: precondition failed, the etag didn't match",
				zap.Int("status_code", response.StatusCode),
			)
			return nil, "", &ratelimitData, errors.New("error updating tailscale grants: precondition failed")
		}

		return nil, "", &ratelimitData, err
	}

	// Ensure response is not nil before deferring close
	if response == nil {
		return nil, "", nil, errors.New("received nil response from HTTP client")
	}

	defer response.Body.Close()

	responseContentType := response.Header.Get("content-type")
	// We expect the content type to be `application/hujson`.
	if !uhttp.IsJSONContentType(responseContentType) {
		return nil, "", nil, fmt.Errorf("unexpected content type for json response: %s", responseContentType)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", nil, err
	}

	// TODO(marcos): The original parser would only read up to 256 kB.
	// Should we also error when that happens?
	parsed, err := hujson.Parse(bodyBytes)
	if err != nil {
		return nil, "", nil, fmt.Errorf("tailscale-connector: %w", err)
	}

	etag = response.Header.Get("etag")
	return &parsed, etag, &ratelimitData, nil
}
