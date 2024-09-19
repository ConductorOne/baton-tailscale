package connector

import (
	"context"
	"os"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
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

	cli, err := getClientForTesting(ctxTest, apiKey, tailnet)
	require.Nil(t, err)

	u := &userBuilder{
		resourceType: userResourceType,
		client:       cli,
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

	cli, err := getClientForTesting(ctxTest, apiKey, tailnet)
	require.Nil(t, err)

	r := &roleBuilder{
		resourceType: userResourceType,
		client:       cli,
	}
	res, _, _, err := r.List(ctxTest, nil, nil)
	require.Nil(t, err)
	require.NotNil(t, res)
}

func TestDeviceBuilderList(t *testing.T) {
	if apiKey == "" && tailnet == "" {
		t.Skip()
	}

	cli, err := getClientForTesting(ctxTest, apiKey, tailnet)
	require.Nil(t, err)

	d := &deviceBuilder{
		resourceType: deviceResourceType,
		client:       cli,
	}
	res, _, _, err := d.List(ctxTest, nil, nil)
	require.Nil(t, err)
	require.NotNil(t, res)
}
