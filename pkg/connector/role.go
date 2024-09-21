package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-tailscale/pkg/connector/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type roleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

// Standard roles: Owner, Admin, Member
// Advanced roles: Billing admin, IT admin, Network admin, Auditor
// https://tailscale.com/kb/1138/user-roles
var roles = []string{"owner", "admin", "member", "billing-admin", "it-admin", "network-admin", "auditor"}

func (r *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return r.resourceType
}

// roleResource creates a new connector resource for a role.
func roleResource(ctx context.Context, role *client.Role, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"role_id":   role.ID,
		"role_name": role.Name,
	}

	roleTraitOptions := []rs.RoleTraitOption{
		rs.WithRoleProfile(profile),
	}

	ret, err := rs.NewRoleResource(
		role.Name,
		roleResourceType,
		role.ID,
		roleTraitOptions,
		rs.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var rv []*v2.Resource
	for _, role := range roles {
		ur, err := roleResource(ctx, &client.Role{
			ID:   role,
			Name: role,
		}, parentResourceID)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, ur)
	}

	return rv, "", nil, nil
}

func (r *roleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return []*v2.Entitlement{
		ent.NewAssignmentEntitlement(
			resource,
			"member",
			ent.WithGrantableTo(userResourceType),
			ent.WithDisplayName(fmt.Sprintf("%s Role Member", resource.DisplayName)),
			ent.WithDescription(fmt.Sprintf("Member of %s Role", resource.DisplayName)),
		),
	}, "", nil, nil
}

func (r *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var rv []*v2.Grant
	roleName := resource.Id.Resource
	users, _, err := r.client.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	for _, user := range users {
		if roleName != user.Role {
			continue
		}

		principalID := &v2.ResourceId{ResourceType: userResourceType.Id, Resource: user.ID}
		gr := grant.NewGrant(resource, "member", principalID)
		rv = append(rv, gr)
	}

	return rv, "", nil, nil
}

func (r *roleBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	if principal.Id.ResourceType != userResourceType.Id {
		l.Warn(
			"baton-tailscale: only users can be granted role membership",
			zap.String("principal_type", principal.Id.ResourceType),
			zap.String("principal_id", principal.Id.Resource),
		)
		return nil, fmt.Errorf("baton-tailscale: only users can be granted role membership")
	}

	userID := principal.Id.Resource
	roleName := entitlement.Resource.Id.Resource
	isOk, err := r.client.UpdateUserRole(ctx, userID, roleName)
	if err != nil {
		return nil, fmt.Errorf("tailscale-connector: failed to add user role: %s", err.Error())
	}

	if isOk {
		l.Info("Role Membership has been created.",
			zap.String("userID", userID),
			zap.String("roleName", roleName),
		)
	}

	return nil, nil
}

func (r *roleBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	principal := grant.Principal
	entitlement := grant.Entitlement
	if principal.Id.ResourceType != userResourceType.Id {
		l.Warn(
			"tailscale-connector: only users can have user membership revoked",
			zap.String("principal_id", principal.Id.String()),
			zap.String("principal_type", principal.Id.ResourceType),
		)

		return nil, fmt.Errorf("tailscale-connector: only users can have user membership revoked")
	}

	userID := principal.Id.Resource
	roleName := entitlement.Resource.Id.Resource
	// users on a tailnet are members by default.
	isOk, err := r.client.UpdateUserRole(ctx, userID, "member")
	if err != nil {
		return nil, fmt.Errorf("tailscale-connector: failed to revoke user role: %s", err.Error())
	}

	if isOk {
		l.Info("Role Membership has been revoked.",
			zap.String("userID", userID),
			zap.String("roleName", roleName),
		)
	}

	return nil, nil
}

func newRoleBuilder(client *client.Client) *roleBuilder {
	return &roleBuilder{
		resourceType: roleResourceType,
		client:       client,
	}
}
