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
)

func TestConnector(t *testing.T) {
	ctx := context.Background()

	if apiKey == "" && tailnet == "" {
		t.Skip()
	}

	cliTest, err := client.New(ctx, apiKey, tailnet, false)
	require.Nil(t, err)

	u := &userBuilder{
		resourceType: userResourceType,
		client:       cliTest,
	}

	r := &roleBuilder{
		resourceType: userResourceType,
		client:       cliTest,
	}

	t.Run("user builder should fetch a list of users", func(t *testing.T) {
		res, _, _, err := u.List(ctx, &v2.ResourceId{}, &pagination.Token{})
		require.Nil(t, err)
		require.NotNil(t, res)
	})

	t.Run("role builder should fetch a list of roles", func(t *testing.T) {
		res, _, _, err := r.List(ctx, nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)
	})
}
