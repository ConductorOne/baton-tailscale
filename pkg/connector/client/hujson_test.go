package client

import (
	"context"
	"strings"
	"testing"

	"github.com/conductorone/baton-tailscale/pkg/utils"
	"github.com/conductorone/baton-tailscale/test"
	"github.com/stretchr/testify/require"
	"github.com/tailscale/hujson"
)

func TestGetPatternFromHujson(t *testing.T) {
	val, err := hujson.Parse([]byte(test.FullJSONExample))
	require.Nil(t, err)

	emailsGeneric := utils.GetPatternFromHujson(val.Value, utils.IsValidEmail)
	groupNamesGeneric := utils.GetPatternFromHujson(val.Value, func(s string) bool {
		return strings.HasPrefix(s, "group:")
	})

	require.Nil(t, err)

	require.Equal(t, []string{
		"justin.gallardo@insulator.one",
		"bjorn.tipling@insulator.one",
		"logan.saso@insulator.one",
		"michael.burton@insulator.one",
		"santhosh.kumar@insulator.one",
		"john.degner@insulator.one",
		"john.degner@insulator.one",
		"logan.saso@insulator.one",
		"logan.saso@gmail.com",
	}, emailsGeneric)

	require.Equal(
		t,
		[]string{
			"group:devs",
			"group:moredevs",
			"group:security",
		},
		groupNamesGeneric,
	)
}

func TestAddEmailToSSHHujson(t *testing.T) {
	ctx := context.Background()
	val, err := hujson.Parse([]byte(test.MinimalSSHExample))
	require.Nil(t, err)

	// f5bbecb4d717767d25115f8c5addd91a8506309a4475285dfaf229d7d17afc02 is the hash of `checkautogroup:selfautogroup:nonrootroot`
	_, err = AddEmailToRule(
		ctx,
		&val,
		RuleKeySSH,
		"f5bbecb4d717767d25115f8c5addd91a8506309a4475285dfaf229d7d17afc02",
		"bonk.flambe@insulator.one",
	)
	require.Nil(t, err)

	require.Equal(t, val.String(), test.ExpectedSSHResult)
}
