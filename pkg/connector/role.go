package connector

import (
	"context"
	"crypto/sha256"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-tailscale/pkg/connector/client"
)

type roleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

var roles = []string{"owner", "admin", "member", "billing-admin", "it-admin", "network-admin", "auditor"}

func (r *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return r.resourceType
}

// roleResource creates a new connector resource for a Zendesk role.
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

// List always returns an empty slice, we don't sync users.
func (r *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var rv []*v2.Resource
	for _, role := range roles {
		hash := sha256.New()
		_, err := hash.Write([]byte(role))
		if err != nil {
			return nil, "", nil, err
		}

		ur, err := roleResource(ctx, &client.Role{
			ID:   fmt.Sprintf("%x", hash.Sum(nil)),
			Name: role,
		}, parentResourceID)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, ur)
	}

	return rv, "", nil, nil
}

// Entitlements always returns an empty slice for users.
func (r *roleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (r *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newRoleBuilder(client *client.Client) *roleBuilder {
	return &roleBuilder{
		resourceType: roleResourceType,
		client:       client,
	}
}
