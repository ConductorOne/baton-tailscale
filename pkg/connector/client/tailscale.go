package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/conductorone/baton-tailscale/pkg/connutils"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const userAgent = "ConductorOne/tailscale-connector-0.2.0"

type Client struct {
	apiKey  string
	tailnet string
	baseUrl *url.URL
	wrapper *uhttp.BaseHttpClient
}

// Documenting api calls
// GET - https://api.tailscale.com/api/v2/tailnet/__TAILNETID__/users"
// GET - https://api.tailscale.com/api/v2/tailnet/__TAILNETID__/devices
// GET - https://api.tailscale.com/api/v2/tailnet/__TAILNETID__/user-invites
// POST - https://api.tailscale.com/api/v2/users/__USERID__/role

// New creates a new client.
func New(ctx context.Context, apiKey string, tailnet string) (*Client, error) {
	httpClient, err := uhttp.NewClient(
		ctx,
		uhttp.WithLogger(true, ctxzap.Extract(ctx)),
		uhttp.WithUserAgent(userAgent),
	)
	if err != nil {
		return nil, err
	}
	wrapper, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey:  apiKey,
		tailnet: tailnet,
		baseUrl: url,
		wrapper: wrapper,
	}, nil
}

func (c *Client) ListGroups(ctx context.Context) ([]Resource, *v2.RateLimitDescription, error) {
	response, _, ratelimitData, err := c.get(ctx)
	if err != nil {
		return nil, ratelimitData, err
	}

	groupNames := connutils.Unique(
		connutils.Convert(
			connutils.GetPatternFromHujson(
				response.Value,
				func(s string) bool {
					return strings.HasPrefix(s, groupPrefix)
				},
			),
			func(s string) string {
				return strings.TrimPrefix(s, groupPrefix)
			},
		),
	)
	groups := make([]Resource, 0)
	for _, groupName := range groupNames {
		groups = append(
			groups,
			Resource{
				DisplayName: groupName,
				Id:          groupPrefix + groupName,
			},
		)
	}

	return groups, ratelimitData, nil
}

func (c *Client) AddEmailToGroup(ctx context.Context, groupName string, email string) (bool, *v2.RateLimitDescription, error) {
	response, etag, ratelimitData, err := c.get(ctx)
	if err != nil {
		return false, ratelimitData, err
	}

	wasAdded, err := AddEmailToGroup(ctx, response, groupName, email)
	if err != nil {
		return false, nil, err
	}

	if !wasAdded {
		return false, ratelimitData, nil
	}

	response.Format()
	// hujson payload bytes
	postBody := response.Pack()
	_, ratelimitData, err = c.post(ctx, postBody, etag)

	return true, ratelimitData, err
}

func (c *Client) RemoveEmailFromGroup(ctx context.Context, groupName string, email string) (bool, *v2.RateLimitDescription, error) {
	response, etag, ratelimitData, err := c.get(ctx)
	if err != nil {
		return false, ratelimitData, err
	}

	wasRemoved, err := RemoveEmailFromGroup(ctx, response, groupName, email)
	if err != nil {
		return false, nil, err
	}

	if !wasRemoved {
		return false, ratelimitData, nil
	}

	response.Format()
	// hujson payload bytes
	postBody := response.Pack()
	_, ratelimitData, err = c.post(ctx, postBody, etag)

	return true, ratelimitData, err
}

func (c *Client) addEmailToRule(
	ctx context.Context,
	ruleHash string,
	ruleKey ruleKey,
	hashPrefix string,
	email string,
) (
	bool,
	*v2.RateLimitDescription,
	error,
) {
	response, etag, ratelimitData, err := c.get(ctx)
	if err != nil {
		return false, ratelimitData, err
	}

	hash := strings.TrimPrefix(ruleHash, fmt.Sprintf("%s:", hashPrefix))
	wasAdded, err := AddEmailToRule(ctx, response, ruleKey, hash, email)
	if err != nil {
		return false, nil, err
	}

	if !wasAdded {
		return false, ratelimitData, nil
	}

	response.Format()
	// hujson payload bytes
	postBody := response.Pack()
	_, ratelimitData, err = c.post(ctx, postBody, etag)

	return true, ratelimitData, err
}

func (c *Client) AddEmailToSSHRule(ctx context.Context, ruleHash string, email string) (bool, *v2.RateLimitDescription, error) {
	return c.addEmailToRule(ctx, ruleHash, RuleKeySSH, "ssh", email)
}

func (c *Client) AddEmailToACLRule(ctx context.Context, ruleHash string, email string) (bool, *v2.RateLimitDescription, error) {
	return c.addEmailToRule(ctx, ruleHash, RuleKeyACLs, "acl", email)
}

func (c *Client) removeEmailFromRule(
	ctx context.Context,
	ruleHash string,
	ruleKey ruleKey,
	hashPrefix string,
	email string,
) (
	bool,
	*v2.RateLimitDescription,
	error,
) {
	response, etag, ratelimitData, err := c.get(ctx)
	if err != nil {
		return false, ratelimitData, err
	}

	hash := strings.TrimPrefix(ruleHash, fmt.Sprintf("%s:", hashPrefix))
	wasRemoved, err := RemoveEmailFromRule(ctx, response, ruleKey, hash, email)
	if err != nil {
		return false, nil, err
	}

	if !wasRemoved {
		return false, ratelimitData, nil
	}

	response.Format()
	// hujson payload bytes
	postBody := response.Pack()
	_, ratelimitData, err = c.post(ctx, postBody, etag)

	return true, ratelimitData, err
}

func (c *Client) RemoveEmailFromSSHRule(ctx context.Context, ruleHash string, email string) (bool, *v2.RateLimitDescription, error) {
	return c.removeEmailFromRule(ctx, ruleHash, RuleKeySSH, "ssh", email)
}

func (c *Client) RemoveEmailFromACLRule(ctx context.Context, ruleHash string, email string) (bool, *v2.RateLimitDescription, error) {
	return c.removeEmailFromRule(ctx, ruleHash, RuleKeyACLs, "acl", email)
}

func (c *Client) ListGroupMemberships(ctx context.Context, groupName string) ([]string, *v2.RateLimitDescription, error) {
	response, _, ratelimitData, err := c.get(ctx)
	if err != nil {
		return nil, ratelimitData, err
	}
	emails, err := GetGroupRulesFromHujson(response.Value, groupName)
	if err != nil {
		return nil, nil, err
	}
	return emails, ratelimitData, nil
}

func (c *Client) listRules(ctx context.Context, key ruleKey, idPrefix string) ([]Resource, *v2.RateLimitDescription, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"listRules",
		zap.String("key", string(key)),
		zap.String("idPrefix", idPrefix),
	)

	target, _, ratelimitData, err := c.get(ctx)
	if err != nil {
		return nil, ratelimitData, err
	}

	rules, err := GetRulesFromHujson(target.Value, key)
	if err != nil {
		return nil, nil, err
	}

	output := make([]Resource, 0)
	for _, foundRule := range rules {
		action := foundRule.GetValueOfNamedMember("action")[0]
		dst := append(
			foundRule.GetValueOfNamedMember("ports"),
			foundRule.GetValueOfNamedMember("dst")...,
		)
		name := connutils.Truncate(strings.Join(dst, ", "), 64)
		newResource := Resource{
			Id:          fmt.Sprintf("%s:%s", idPrefix, foundRule.GetHash()),
			DisplayName: fmt.Sprintf("%s: %s", action, name),
		}
		output = append(output, newResource)
	}

	return output, ratelimitData, nil
}

func (c *Client) ListSSHRules(ctx context.Context) ([]Resource, *v2.RateLimitDescription, error) {
	return c.listRules(ctx, "ssh", "ssh")
}

func (c *Client) ListACLRules(ctx context.Context) ([]Resource, *v2.RateLimitDescription, error) {
	return c.listRules(ctx, "acls", "acl")
}

func (c *Client) listRuleEmails(
	ctx context.Context,
	ruleId string,
	key ruleKey,
	idPrefix string,
) (
	[]string,
	*v2.RateLimitDescription,
	error,
) {
	response, _, ratelimitData, err := c.get(ctx)
	if err != nil {
		return nil, ratelimitData, err
	}

	rules, err := GetRulesFromHujson(response.Value, key)
	if err != nil {
		return nil, nil, err
	}

	emails := make([]string, 0)
	for _, foundRule := range rules {
		hash := fmt.Sprintf("%s:%s", idPrefix, foundRule.GetHash())
		if hash != ruleId {
			continue
		}
		for _, email := range foundRule.GetValueOfNamedMember("src") {
			if connutils.IsValidEmail(email) {
				emails = append(emails, email)
			}
		}
		for _, email := range foundRule.GetValueOfNamedMember("users") {
			if connutils.IsValidEmail(email) {
				emails = append(emails, email)
			}
		}
	}

	return emails, ratelimitData, nil
}

func (c *Client) ListSSHEmails(ctx context.Context, ruleId string) ([]string, *v2.RateLimitDescription, error) {
	return c.listRuleEmails(ctx, ruleId, "ssh", "ssh")
}

func (c *Client) ListACLEmails(ctx context.Context, ruleId string) ([]string, *v2.RateLimitDescription, error) {
	return c.listRuleEmails(ctx, ruleId, "acls", "acl")
}

func WithAuthorizationBearerHeader(token string) uhttp.RequestOption {
	return uhttp.WithHeader("Authorization", "Bearer "+token)
}

func (c *Client) doRequest(ctx context.Context, path string, target interface{}) (*v2.RateLimitDescription, error) {
	request, err := c.wrapper.NewRequest(
		ctx,
		http.MethodGet,
		c.baseUrl.JoinPath(path),
		uhttp.WithAcceptJSONHeader(),
		WithAuthorizationBearerHeader(c.apiKey),
	)
	if err != nil {
		return nil, err
	}

	var ratelimitData v2.RateLimitDescription
	response, err := c.wrapper.Do(
		request,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&ratelimitData),
	)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	return &ratelimitData, nil
}

// GetUsers. Get all users. Only authenticated users may call this resource.
// https://tailscale.com/api#tag/users/GET/tailnet/{tailnet}/users
// The Tailscale API does not currently support pagination. All results are returned at once.
func (c *Client) GetUsers(ctx context.Context) ([]User, *v2.RateLimitDescription, error) {
	var userData UsersAPIData
	endpointUrl, err := url.JoinPath("tailnet", c.tailnet, "users")
	if err != nil {
		return nil, nil, err
	}

	ratelimitData, err := c.doRequest(ctx, endpointUrl, &userData)
	if err != nil {
		return nil, ratelimitData, err
	}

	return userData.Users, ratelimitData, nil
}

// GetUserInvites. Get all users invites. Only authenticated users may call this resource.
// https://tailscale.com/api#tag/userinvites/GET/tailnet/{tailnet}/user-invites
// The Tailscale API does not currently support pagination. All results are returned at once.
func (c *Client) GetUserInvites(ctx context.Context) (UserInvitesAPIData, *v2.RateLimitDescription, error) {
	var userInviteData UserInvitesAPIData
	endpointUrl, err := url.JoinPath("tailnet", c.tailnet, "user-invites")
	if err != nil {
		return nil, nil, err
	}

	ratelimitData, err := c.doRequest(ctx, endpointUrl, &userInviteData)
	if err != nil {
		return nil, ratelimitData, err
	}

	return userInviteData, ratelimitData, nil
}

// GetDevices. Get all devices. Only authenticated users may call this resource.
// https://tailscale.com/api#tag/devices/GET/tailnet/{tailnet}/devices
// The Tailscale API does not currently support pagination. All results are returned at once.
func (c *Client) GetDevices(ctx context.Context) ([]Device, *v2.RateLimitDescription, error) {
	var deviceData DevicesAPIData
	endpointUrl, err := url.JoinPath("tailnet", c.tailnet, "devices")
	if err != nil {
		return nil, nil, err
	}

	ratelimitData, err := c.doRequest(ctx, endpointUrl, &deviceData)
	if err != nil {
		return nil, ratelimitData, err
	}

	return deviceData.Devices, ratelimitData, nil
}

// UpdateUserRole. Updates user-roles
// https://tailscale.com/api#tag/users/POST/users/{userId}/role
func (c *Client) UpdateUserRole(ctx context.Context, userId, roleName string) error {
	var (
		body struct {
			Role string `json:"role"`
		}
		payload = []byte(fmt.Sprintf(`{"role": "%s"}`, roleName))
	)

	endpointUrl, err := url.JoinPath(c.baseUrl.String(), "users", userId, "role")
	if err != nil {
		return err
	}

	uri, err := url.Parse(endpointUrl)
	if err != nil {
		return err
	}

	err = json.Unmarshal(payload, &body)
	if err != nil {
		return err
	}

	req, err := c.wrapper.NewRequest(ctx,
		http.MethodPost,
		uri,
		uhttp.WithAcceptJSONHeader(),
		WithAuthorizationBearerHeader(c.apiKey),
		uhttp.WithJSONBody(body),
	)
	if err != nil {
		return err
	}

	resp, err := c.wrapper.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// SetDeviceAttribute sets a custom attribute on a device using the Tailscale API
// POST /api/v2/device/{deviceId}/attributes/{attributeKey}.
func (c *Client) SetDeviceAttribute(ctx context.Context, deviceID, attributeKey, attributeValue string, expiryTimestamp string, comment string) error {
	// Build the full URL for the device attribute endpoint
	deviceURL := c.baseUrl.JoinPath("device", deviceID, "attributes", attributeKey)

	// Create the request body with the attribute value
	requestBody := map[string]string{
		"value": attributeValue,
	}

	if expiryTimestamp != "" {
		requestBody["expiry"] = expiryTimestamp
	}

	if comment != "" {
		requestBody["comment"] = comment
	}

	req, err := c.wrapper.NewRequest(
		ctx,
		http.MethodPost,
		deviceURL,
		uhttp.WithAcceptJSONHeader(),
		WithAuthorizationBearerHeader(c.apiKey),
		uhttp.WithJSONBody(requestBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Make the request using the wrapper's Do method
	response, err := c.wrapper.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set device attribute: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to set device attribute: HTTP %d", response.StatusCode)
	}

	return nil
}

// DeleteDeviceAttribute deletes a custom attribute from a device using the Tailscale API
// DELETE /api/v2/device/{deviceId}/attributes/{attributeKey}.
func (c *Client) DeleteDeviceAttribute(ctx context.Context, deviceID, attributeKey string) error {
	// 200 with a null body is returned if the attribute is not found, so we can't check for that.
	// Build the full URL for the device attribute endpoint
	deviceURL := c.baseUrl.JoinPath("device", deviceID, "attributes", attributeKey)

	// Create the DELETE request
	req, err := c.wrapper.NewRequest(
		ctx,
		http.MethodDelete,
		deviceURL,
		uhttp.WithAcceptJSONHeader(),
		WithAuthorizationBearerHeader(c.apiKey),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Make the request using the wrapper's Do method
	response, err := c.wrapper.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete device attribute: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete device attribute: HTTP %d", response.StatusCode)
	}

	return nil
}
