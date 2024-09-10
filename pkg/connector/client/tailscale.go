package client

import (
	"context"
	"fmt"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/conductorone/baton-tailscale/pkg/utils"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

const (
	userAgent = "ConductorOne/tailscale-connector-0.2.0"
)

type Client struct {
	apiKey  string
	tailnet string
	wrapper *uhttp.BaseHttpClient
}

// New creates a new client.
func New(ctx context.Context, apiKey string, tailnet string) (*Client, error) {
	httpClient, err := uhttp.NewClient(
		ctx,
		uhttp.WithLogger(
			true,
			ctxzap.Extract(ctx),
		),
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
		wrapper: wrapper,
	}, nil
}

func (c *Client) ListGroups(ctx context.Context) (
	[]Resource,
	*v2.RateLimitDescription,
	error,
) {
	var response tailscaleResponseTrimmed
	_, ratelimitData, err := c.get(ctx, &response)
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

func (c *Client) AddEmailToGroup(
	ctx context.Context,
	groupName string,
	email string,
) (
	bool,
	*v2.RateLimitDescription,
	error,
) {
	var response tailscaleResponse
	etag, ratelimitData, err := c.get(ctx, &response)
	if err != nil {
		return false, ratelimitData, err
	}

	wasAdded, err := AddEmailToGroup(ctx, response.Value, groupName, email)
	if err != nil {
		return false, nil, err
	}

	if !wasAdded {
		return false, ratelimitData, nil
	}

	response.Value.Format()
	postBody := strings.NewReader(response.Value.String())

	ratelimitData, err = c.post(ctx, postBody, etag)
	return true, ratelimitData, err
}

func (c *Client) RemoveEmailFromGroup(
	ctx context.Context,
	groupName string,
	email string,
) (
	bool,
	*v2.RateLimitDescription,
	error,
) {
	var response tailscaleResponse
	etag, ratelimitData, err := c.get(ctx, &response)
	if err != nil {
		return false, ratelimitData, err
	}

	wasRemoved, err := RemoveEmailFromGroup(ctx, response.Value, groupName, email)
	if err != nil {
		return false, nil, err
	}

	if !wasRemoved {
		return false, ratelimitData, nil
	}

	response.Value.Format()
	postBody := strings.NewReader(response.Value.String())

	ratelimitData, err = c.post(ctx, postBody, etag)
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
	var response tailscaleResponse
	etag, ratelimitData, err := c.get(ctx, &response)
	if err != nil {
		return false, ratelimitData, err
	}

	hash := strings.TrimPrefix(ruleHash, fmt.Sprintf("%s:", hashPrefix))
	wasAdded, err := AddEmailToRule(ctx, response.Value, ruleKey, hash, email)
	if err != nil {
		return false, nil, err
	}

	if !wasAdded {
		return false, ratelimitData, nil
	}

	response.Value.Format()
	postBody := strings.NewReader(response.Value.String())

	ratelimitData, err = c.post(ctx, postBody, etag)
	return true, ratelimitData, err
}

func (c *Client) AddEmailToSSHRule(
	ctx context.Context,
	ruleHash string,
	email string,
) (
	bool,
	*v2.RateLimitDescription,
	error,
) {
	return c.addEmailToRule(ctx, ruleHash, RuleKeySSH, "ssh", email)
}

func (c *Client) AddEmailToACLRule(
	ctx context.Context,
	ruleHash string,
	email string,
) (
	bool,
	*v2.RateLimitDescription,
	error,
) {
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
	var response tailscaleResponse
	etag, ratelimitData, err := c.get(ctx, &response)
	if err != nil {
		return false, ratelimitData, err
	}

	hash := strings.TrimPrefix(ruleHash, fmt.Sprintf("%s:", hashPrefix))
	wasRemoved, err := RemoveEmailFromRule(ctx, response.Value, ruleKey, hash, email)
	if err != nil {
		return false, nil, err
	}

	if !wasRemoved {
		return false, ratelimitData, nil
	}

	response.Value.Format()
	postBody := strings.NewReader(response.Value.String())

	ratelimitData, err = c.post(ctx, postBody, etag)
	return true, ratelimitData, err
}

func (c *Client) RemoveEmailFromSSHRule(
	ctx context.Context,
	ruleHash string,
	email string,
) (
	bool,
	*v2.RateLimitDescription,
	error,
) {
	return c.removeEmailFromRule(ctx, ruleHash, RuleKeySSH, "ssh", email)
}

func (c *Client) RemoveEmailFromACLRule(
	ctx context.Context,
	ruleHash string,
	email string,
) (
	bool,
	*v2.RateLimitDescription,
	error,
) {
	return c.removeEmailFromRule(ctx, ruleHash, RuleKeyACLs, "acl", email)
}

func (c *Client) ListGroupMemberships(
	ctx context.Context,
	groupName string,
) (
	[]string,
	*v2.RateLimitDescription,
	error,
) {
	var response tailscaleResponseTrimmed
	_, ratelimitData, err := c.get(ctx, &response)
	if err != nil {
		return nil, ratelimitData, err
	}
	emails, err := GetGroupRulesFromHujson(response.Value, groupName)
	if err != nil {
		return nil, nil, err
	}
	return emails, ratelimitData, nil
}

func (c *Client) listRules(
	ctx context.Context,
	key ruleKey,
	idPrefix string,
) (
	[]Resource,
	*v2.RateLimitDescription,
	error,
) {
	var response tailscaleResponseTrimmed
	_, ratelimitData, err := c.get(ctx, &response)
	if err != nil {
		return nil, ratelimitData, err
	}
	rules, err := GetRulesFromHujson(response.Value, key)
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

func (c *Client) ListSSHRules(ctx context.Context) (
	[]Resource,
	*v2.RateLimitDescription,
	error,
) {
	return c.listRules(ctx, "ssh", "ssh")
}

func (c *Client) ListACLRules(ctx context.Context) (
	[]Resource,
	*v2.RateLimitDescription,
	error,
) {
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
	var response tailscaleResponseTrimmed
	_, ratelimitData, err := c.get(ctx, &response)
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

func (c *Client) ListSSHEmails(ctx context.Context, ruleId string) (
	[]string,
	*v2.RateLimitDescription,
	error,
) {
	return c.listRuleEmails(ctx, ruleId, "ssh", "ssh")
}

func (c *Client) ListACLEmails(ctx context.Context, ruleId string) (
	[]string,
	*v2.RateLimitDescription,
	error,
) {
	return c.listRuleEmails(ctx, ruleId, "acls", "acl")
}
