package connector

import (
	"context"
	"fmt"
	"os"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-tailscale/pkg/connector/client"
	"github.com/stretchr/testify/require"
)

var (
	apiKey  = os.Getenv("BATON_API_KEY")
	tailnet = os.Getenv("BATON_TAILNET")
	ctxTest = context.Background()
)

func TestUserBuilderList(t *testing.T) {
	if apiKey == "" && tailnet == "" {
		t.Skip()
	}

	cliTest, err := getClientForTesting(ctxTest, apiKey, tailnet)
	require.Nil(t, err)

	u := &userBuilder{
		resourceType: userResourceType,
		client:       cliTest,
	}

	res, _, _, err := u.List(ctxTest, &v2.ResourceId{}, &pagination.Token{})
	require.Nil(t, err)
	require.NotNil(t, res)
}

func getClientForTesting(ctx context.Context, apiKey string, tailnet string) (*client.Client, error) {
	client, err := client.New(ctx, apiKey, tailnet)
	if err != nil {
		return nil, err
	}

	return client, err
}

func TestRoleBuilderList(t *testing.T) {
	if apiKey == "" && tailnet == "" {
		t.Skip()
	}

	cliTest, err := getClientForTesting(ctxTest, apiKey, tailnet)
	require.Nil(t, err)

	r := &roleBuilder{
		resourceType: userResourceType,
		client:       cliTest,
	}
	res, _, _, err := r.List(ctxTest, nil, nil)
	require.Nil(t, err)
	require.NotNil(t, res)
}

func TestDeviceBuilderList(t *testing.T) {
	if apiKey == "" && tailnet == "" {
		t.Skip()
	}

	cliTest, err := getClientForTesting(ctxTest, apiKey, tailnet)
	require.Nil(t, err)

	d := &deviceBuilder{
		resourceType: deviceResourceType,
		client:       cliTest,
	}
	res, _, _, err := d.List(ctxTest, nil, nil)
	require.Nil(t, err)
	require.NotNil(t, res)
}

func TestGetUserInvites(t *testing.T) {
	if apiKey == "" && tailnet == "" {
		t.Skip()
	}

	cliTest, err := getClientForTesting(ctxTest, apiKey, tailnet)
	require.Nil(t, err)

	u := &userBuilder{
		client: cliTest,
	}
	res, err := u.client.GetUserInvites(ctxTest)
	require.Nil(t, err)
	require.NotNil(t, res)
}

func TestAddUserRole(t *testing.T) {
	if apiKey == "" && tailnet == "" {
		t.Skip()
	}

	cliTest, err := getClientForTesting(ctxTest, apiKey, tailnet)
	require.Nil(t, err)

	u := &userBuilder{
		client: cliTest,
	}
	res, err := u.client.AddUserRole(ctxTest, "uYmTSnEi9711CNTRL", "billing-admin")
	require.Nil(t, err)
	require.NotNil(t, res)
}

func TestRoleGrant(t *testing.T) {
	var licenseEntitlement string
	if apiKey == "" && tailnet == "" {
		t.Skip()
	}

	cliTest, err := getClientForTesting(ctxTest, apiKey, tailnet)
	require.Nil(t, err)

	// --grant-entitlement role:billing-admin:members
	grantEntitlement := "role:billing-admin:members"
	// --grant-principal-type user
	grantPrincipalType := "user"
	// --grant-principal uYmTSnEi9711CNTRL
	grantPrincipal := "uYmTSnEi9711CNTRL"
	_, data, err := parseEntitlementID(grantEntitlement)
	require.Nil(t, err)
	require.NotNil(t, data)

	licenseEntitlement = data[2]
	resource, err := getRoleResourceForTesting(ctxTest)
	require.Nil(t, err)

	entitlement := getEntitlementForTesting(resource, grantPrincipalType, licenseEntitlement)
	l := &roleBuilder{
		client: cliTest,
	}
	_, err = l.Grant(ctxTest, &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: userResourceType.Id,
			Resource:     grantPrincipal,
		},
	}, entitlement)
	require.Nil(t, err)
}

func getRoleResourceForTesting(ctxTest context.Context) (*v2.Resource, error) {
	return roleResource(ctxTest, &client.Role{
		ID:   "billing-admin",
		Name: "billing-admin",
	}, nil)
}

func getEntitlementForTesting(resource *v2.Resource, resourceDisplayName, licenseEntitlement string) *v2.Entitlement {
	options := []ent.EntitlementOption{
		ent.WithGrantableTo(roleResourceType),
		ent.WithDisplayName(fmt.Sprintf("%s resource %s", resourceDisplayName, licenseEntitlement)),
		ent.WithDescription(fmt.Sprintf("%s of %s Tialscale role", licenseEntitlement, resourceDisplayName)),
	}

	return ent.NewAssignmentEntitlement(resource, licenseEntitlement, options...)
}
