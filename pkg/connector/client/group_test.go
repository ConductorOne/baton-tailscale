package client

import (
	"context"
	"strings"
	"testing"

	"github.com/conductorone/baton-tailscale/pkg/connutils"
	"github.com/conductorone/baton-tailscale/test"
	"github.com/stretchr/testify/require"
	"github.com/tailscale/hujson"
)

func TestGetRuleFromHujson(t *testing.T) {
	val, err := hujson.Parse([]byte(test.FullJSONExample))
	require.Nil(t, err)

	sshRules, err := GetRulesFromHujson(val.Value, RuleKeySSH)
	require.Nil(t, err)

	aclRules, err := GetRulesFromHujson(val.Value, RuleKeyACLs)
	require.Nil(t, err)

	groupNamesGeneric := connutils.GetPatternFromHujson(val.Value, func(s string) bool {
		return strings.HasPrefix(s, "group:")
	})

	groupEmailArrays := map[string][]string{
		"group:devs": {
			"justin.gallardo@insulator.one",
			"bjorn.tipling@insulator.one",
			"logan.saso@insulator.one",
		},
		"group:moredevs": {
			"michael.burton@insulator.one",
			"santhosh.kumar@insulator.one",
			"john.degner@insulator.one",
		},
		"group:security": {},
	}

	for _, name := range groupNamesGeneric {
		groupRules, err := GetGroupRulesFromHujson(val.Value, name)
		require.Nil(t, err)
		emails := groupEmailArrays[name]
		require.Equal(t, emails, groupRules)
	}

	require.Len(t, aclRules, 3)
	require.Len(t, sshRules, 1)
}

func TestAddEmailToGroupHujson(t *testing.T) {
	ctx := context.Background()
	val, err := hujson.Parse([]byte(test.MinimalGroupsExample))
	require.Nil(t, err)

	_, err = AddEmailToGroup(ctx, &val, "group:devs", "bonk.flambe@insulator.one")
	require.Nil(t, err)

	require.Equal(t, val.String(), test.ExpectedGroupsResult)
}

func TestRemoveEmailFromGroupHujson(t *testing.T) {
	ctx := context.Background()
	val, err := hujson.Parse([]byte(test.ExpectedGroupsResult))
	require.Nil(t, err)

	_, err = RemoveEmailFromGroup(ctx, &val, "group:devs", "bonk.flambe@insulator.one")
	require.Nil(t, err)

	require.Equal(t, val.String(), test.MinimalGroupsExample)
}

func TestAddEmailToACLHujson(t *testing.T) {
	ctx := context.Background()
	val, err := hujson.Parse([]byte(test.MinimalACLExample))
	require.Nil(t, err)

	// 5737a0c593a5c4d6473966340e2d0261297a605d741689818c3691307df2a613 is the
	// hash of `acceptjohn.degner@insulator.onelogan.saso@insulator.one`
	_, err = AddEmailToRule(
		ctx,
		&val,
		RuleKeyACLs,
		"5737a0c593a5c4d6473966340e2d0261297a605d741689818c3691307df2a613",
		"bonk.flambe@insulator.one",
	)
	require.Nil(t, err)

	require.Equal(t, val.String(), test.ExpectedACLResult)
}

func TestRemoveEmailFromACLHujson(t *testing.T) {
	ctx := context.Background()
	val, err := hujson.Parse([]byte(test.ExpectedACLResult))
	require.Nil(t, err)

	// 5737a0c593a5c4d6473966340e2d0261297a605d741689818c3691307df2a613 is the
	// hash of `acceptjohn.degner@insulator.onelogan.saso@insulator.one`
	_, err = RemoveEmailFromRule(
		ctx,
		&val,
		RuleKeyACLs,
		"5737a0c593a5c4d6473966340e2d0261297a605d741689818c3691307df2a613",
		"bonk.flambe@insulator.one",
	)
	require.Nil(t, err)

	require.Equal(t, val.String(), test.MinimalACLExample)
}
