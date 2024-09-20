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
	"github.com/conductorone/baton-tailscale/pkg/utils"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const userAgent = "ConductorOne/tailscale-connector-0.2.0"

type Client struct {
	apiKey  string
	tailnet string
	baseUrl string
	wrapper *uhttp.BaseHttpClient
}

// GET - https://api.tailscale.com/api/v2/tailnet/example.com/users"
// GET - https://api.tailscale.com/api/v2/tailnet/example.com/devices
// GET - https://api.tailscale.com/api/v2/tailnet/example.com/user-invites

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

	return &Client{
		apiKey:  apiKey,
		tailnet: tailnet,
		baseUrl: "https://api.tailscale.com/api/v2",
		wrapper: wrapper,
	}, nil
}

func (c *Client) ListGroups(ctx context.Context) ([]Resource, *v2.RateLimitDescription, error) {
	response, _, ratelimitData, err := c.get(ctx)
	if err != nil {
		return nil, ratelimitData, err
	}

	groupNames := utils.Unique(
		utils.Convert(
			utils.GetPatternFromHujson(
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
	postBody := strings.NewReader(response.String())
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
	postBody := strings.NewReader(response.String())
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
	postBody := strings.NewReader(response.String())
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
	postBody := strings.NewReader(response.String())
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
		name := utils.Truncate(strings.Join(dst, ", "), 64)
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
			if utils.IsValidEmail(email) {
				emails = append(emails, email)
			}
		}
		for _, email := range foundRule.GetValueOfNamedMember("users") {
			if utils.IsValidEmail(email) {
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

// GetUsers. Get all users. Only authenticated users may call this resource.
// https://tailscale.com/api#tag/users/GET/tailnet/{tailnet}/users
// The Tailscale API does not currently support pagination. All results are returned at once.
func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	var userData UsersAPIData
	endpointUrl, err := url.JoinPath(c.baseUrl, "tailnet", c.tailnet, "users")
	if err != nil {
		return nil, err
	}

	uri, err := url.Parse(endpointUrl)
	if err != nil {
		return nil, err
	}

	req, err := c.wrapper.NewRequest(ctx,
		http.MethodGet,
		uri,
		uhttp.WithAcceptJSONHeader(),
		WithAuthorizationBearerHeader(c.apiKey),
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.wrapper.Do(req, uhttp.WithJSONResponse(&userData))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return userData.Users, nil
}

// GetUserInvites. Get all users invites. Only authenticated users may call this resource.
// https://tailscale.com/api#tag/userinvites/GET/tailnet/{tailnet}/user-invites
// The Tailscale API does not currently support pagination. All results are returned at once.
func (c *Client) GetUserInvites(ctx context.Context) (UserInvitesAPIData, error) {
	var userInviteData UserInvitesAPIData
	endpointUrl, err := url.JoinPath(c.baseUrl, "tailnet", c.tailnet, "user-invites")
	if err != nil {
		return nil, err
	}

	uri, err := url.Parse(endpointUrl)
	if err != nil {
		return nil, err
	}

	req, err := c.wrapper.NewRequest(ctx,
		http.MethodGet,
		uri,
		uhttp.WithAcceptJSONHeader(),
		WithAuthorizationBearerHeader(c.apiKey),
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.wrapper.Do(req, uhttp.WithJSONResponse(&userInviteData))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return userInviteData, nil
}

// GetDevices. Get all devices. Only authenticated users may call this resource.
// https://tailscale.com/api#tag/devices/GET/tailnet/{tailnet}/devices
// The Tailscale API does not currently support pagination. All results are returned at once.
func (c *Client) GetDevices(ctx context.Context) ([]Device, error) {
	var deviceData DevicesAPIData
	endpointUrl, err := url.JoinPath(c.baseUrl, "tailnet", c.tailnet, "devices")
	if err != nil {
		return nil, err
	}

	uri, err := url.Parse(endpointUrl)
	if err != nil {
		return nil, err
	}

	req, err := c.wrapper.NewRequest(ctx,
		http.MethodGet,
		uri,
		uhttp.WithAcceptJSONHeader(),
		WithAuthorizationBearerHeader(c.apiKey),
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.wrapper.Do(req, uhttp.WithJSONResponse(&deviceData))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return deviceData.Devices, nil
}

func (c *Client) AddUserRole(ctx context.Context, userId, roleName string) (bool, error) {
	var (
		body struct {
			Role string `json:"role"`
		}
		payload = []byte(fmt.Sprintf(`{"role": "%s"}`, roleName))
	)

	endpointUrl, err := url.JoinPath(c.baseUrl, "users", userId, "role")
	if err != nil {
		return false, err
	}

	uri, err := url.Parse(endpointUrl)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(payload, &body)
	if err != nil {
		return false, err
	}

	req, err := c.wrapper.NewRequest(ctx,
		http.MethodPost,
		uri,
		uhttp.WithAcceptJSONHeader(),
		WithAuthorizationBearerHeader(c.apiKey),
		uhttp.WithJSONBody(body),
	)
	if err != nil {
		return false, err
	}

	resp, err := c.wrapper.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}
