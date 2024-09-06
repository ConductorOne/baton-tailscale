package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	resourceSDK "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-tailscale/pkg/connector/client"
	"github.com/conductorone/baton-tailscale/pkg/utils"
)

type sshRuleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

func sshRuleResource(sshRule client.Resource, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	return resourceSDK.NewResource(
		sshRule.DisplayName,
		sshRuleResourceType,
		sshRule.Id,
		resourceSDK.WithParentResourceID(parentResourceID),
	)
}

func (o *sshRuleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

func (o *sshRuleBuilder) List(
	ctx context.Context,
	parentID *v2.ResourceId,
	_ *pagination.Token,
) (
	[]*v2.Resource,
	string,
	annotations.Annotations,
	error,
) {
	rules, ratelimitData, err := o.client.ListSSHRules(ctx)
	outputAnnotations := utils.WithRatelimitAnnotations(ratelimitData)
	if err != nil {
		return nil, "", outputAnnotations, err
	}

	output := make([]*v2.Resource, 0)
	for _, rule := range rules {
		newResource, err := sshRuleResource(rule, parentID)
		if err != nil {
			return nil, "", outputAnnotations, err
		}
		output = append(output, newResource)
	}
	return output, "", outputAnnotations, nil
}

func (o *sshRuleBuilder) Entitlements(
	_ context.Context,
	resource *v2.Resource,
	_ *pagination.Token,
) (
	[]*v2.Entitlement,
	string,
	annotations.Annotations,
	error,
) {
	membership := entitlement.NewAssignmentEntitlement(
		resource,
		entitlementName,
		entitlement.WithDisplayName(
			fmt.Sprintf("%s SSH Rule Member", resource.DisplayName),
		),
		entitlement.WithDescription(
			fmt.Sprintf("Is matched against the %s SSH Rule in Tailscale", resource.DisplayName),
		),
	)

	return []*v2.Entitlement{membership}, "", nil, nil
}

func (o *sshRuleBuilder) Grants(
	ctx context.Context,
	resource *v2.Resource,
	_ *pagination.Token,
) (
	[]*v2.Grant,
	string,
	annotations.Annotations,
	error,
) {
	emails, ratelimitData, err := o.client.ListSSHEmails(ctx, resource.Id.Resource)
	outputAnnotations := utils.WithRatelimitAnnotations(ratelimitData)
	if err != nil {
		return nil, "", outputAnnotations, err
	}

	grants := emailToGrants(resource, emails)
	return grants, "", outputAnnotations, nil
}

func (o *sshRuleBuilder) Grant(
	ctx context.Context,
	principal *v2.Resource,
	entitlement *v2.Entitlement,
) (annotations.Annotations, error) {
	wasAdded, ratelimitData, err := o.client.AddEmailToSshRule(ctx, entitlement.Id, principal.Id.Resource)
	outputAnnotations := utils.WithRatelimitAnnotations(ratelimitData)
	if err != nil {
		return outputAnnotations, err
	}

	if !wasAdded {
		outputAnnotations.Append(&v2.GrantAlreadyExists{})
	}

	return outputAnnotations, nil
}

func (o *sshRuleBuilder) Revoke(
	ctx context.Context,
	grant *v2.Grant,
) (annotations.Annotations, error) {
	wasRevoked, ratelimitData, err := o.client.RemoveEmailFromSshRule(
		ctx,
		grant.Entitlement.Id,
		grant.Principal.Id.Resource,
	)
	outputAnnotations := utils.WithRatelimitAnnotations(ratelimitData)
	if err != nil {
		return outputAnnotations, err
	}

	if !wasRevoked {
		outputAnnotations.Append(&v2.GrantAlreadyRevoked{})
	}

	return outputAnnotations, nil
}

func newSshRuleBuilder(client *client.Client) *sshRuleBuilder {
	return &sshRuleBuilder{
		resourceType: userResourceType,
		client:       client,
	}
}
