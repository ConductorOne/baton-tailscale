package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	resourceSDK "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-tailscale/pkg/connector/client"
	"github.com/conductorone/baton-tailscale/pkg/connutils"
)

const (
	entitlementName = "member"
)

type groupBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

func groupResource(group client.Resource, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	return resourceSDK.NewGroupResource(
		group.DisplayName,
		groupResourceType,
		group.Id,
		nil,
		resourceSDK.WithParentResourceID(parentResourceID),
	)
}

func (o *groupBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

func (o *groupBuilder) List(
	ctx context.Context,
	parentID *v2.ResourceId,
	_ *pagination.Token,
) (
	[]*v2.Resource,
	string,
	annotations.Annotations,
	error,
) {
	groups, ratelimitData, err := o.client.ListGroups(ctx)
	outputAnnotations := connutils.WithRatelimitAnnotations(ratelimitData)
	if err != nil {
		return nil, "", outputAnnotations, err
	}

	output := make([]*v2.Resource, 0)
	for _, group := range groups {
		newGroupResource, err := groupResource(group, parentID)
		if err != nil {
			continue
		}
		output = append(output, newGroupResource)
	}
	return output, "", outputAnnotations, nil
}

func (o *groupBuilder) Entitlements(
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
			fmt.Sprintf("%s Group Member", resource.DisplayName),
		),
		entitlement.WithDescription(
			fmt.Sprintf("Is member of the %s group in Tailscale", resource.DisplayName),
		),
	)

	return []*v2.Entitlement{membership}, "", nil, nil
}

func userIDsToGrants(resource *v2.Resource, userIDs []string) []*v2.Grant {
	output := make([]*v2.Grant, 0)
	for _, userID := range userIDs {
		userRes := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     userID,
			},
		}

		output = append(
			output,
			grant.NewGrant(
				resource,
				entitlementName,
				userRes.Id,
			),
		)
	}
	return output
}

func (o *groupBuilder) Grants(
	ctx context.Context,
	resource *v2.Resource,
	_ *pagination.Token,
) (
	[]*v2.Grant,
	string,
	annotations.Annotations,
	error,
) {
	var grants []*v2.Grant
	var userIDs []string
	var users []client.User

	users, _, err := o.client.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	emails, ratelimitData, err := o.client.ListGroupMemberships(ctx, resource.Id.Resource)
	outputAnnotations := connutils.WithRatelimitAnnotations(ratelimitData)
	if err != nil {
		return nil, "", outputAnnotations, err
	}

	userInvites, _, err := o.client.GetUserInvites(ctx)
	if err != nil {
		return nil, "", nil, err
	}
	for _, userInvite := range userInvites {
		users = append(users, client.User{
			ID:        userInvite.ID,
			LoginName: userInvite.Email,
		})
	}

	userIDs = GetUserIDsFromUserEmails(users, emails)
	grants = append(grants, userIDsToGrants(resource, userIDs)...)

	return grants, "", outputAnnotations, nil
}

func (o *groupBuilder) Grant(
	ctx context.Context,
	principal *v2.Resource,
	entitlement *v2.Entitlement,
) (annotations.Annotations, error) {
	userTrait, err := resourceSDK.GetUserTrait(principal)
	if err != nil {
		return nil, fmt.Errorf("tailscale-connector: Failed to get user trait from user: %w", err)
	}

	wasAdded, ratelimitData, err := o.client.AddEmailToGroup(ctx, entitlement.Resource.Id.Resource, userTrait.GetLogin())
	outputAnnotations := connutils.WithRatelimitAnnotations(ratelimitData)
	if err != nil {
		return outputAnnotations, err
	}

	if !wasAdded {
		outputAnnotations.Append(&v2.GrantAlreadyExists{})
	}

	return outputAnnotations, nil
}

func (o *groupBuilder) Revoke(
	ctx context.Context,
	grant *v2.Grant,
) (annotations.Annotations, error) {
	userTrait, err := resourceSDK.GetUserTrait(grant.GetPrincipal())
	if err != nil {
		return nil, fmt.Errorf("tailscale-connector: Failed to get user trait from user: %w", err)
	}
	wasRevoked, ratelimitData, err := o.client.RemoveEmailFromGroup(
		ctx,
		grant.Entitlement.Resource.Id.Resource,
		userTrait.GetLogin(),
	)
	outputAnnotations := connutils.WithRatelimitAnnotations(ratelimitData)
	if err != nil {
		return outputAnnotations, err
	}

	if !wasRevoked {
		outputAnnotations.Append(&v2.GrantAlreadyRevoked{})
	}

	return outputAnnotations, nil
}

func newGroupBuilder(client *client.Client) *groupBuilder {
	return &groupBuilder{
		resourceType: groupResourceType,
		client:       client,
	}
}
