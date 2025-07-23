package client

import (
	"context"
	"strings"
	"testing"

	"github.com/conductorone/baton-tailscale/pkg/connutils"
	"github.com/conductorone/baton-tailscale/test"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tailscale/hujson"
)

type HujsonSuite struct {
	suite.Suite
	ctx context.Context
}

func (u *HujsonSuite) SetupTest() {
	u.ctx = context.Background()
}

func TestHujsonSuite(t *testing.T) {
	suite.Run(t, new(HujsonSuite))
}

func (u *HujsonSuite) TestGetRuleFromHujson() {
	val, err := hujson.Parse([]byte(test.FullJSONExample))
	require.Nil(u.T(), err)

	sshRules, err := GetRulesFromHujson(val.Value, RuleKeySSH)
	require.Nil(u.T(), err)

	aclRules, err := GetRulesFromHujson(val.Value, RuleKeyACLs)
	require.Nil(u.T(), err)

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
		require.Nil(u.T(), err)
		emails := groupEmailArrays[name]
		require.Equal(u.T(), emails, groupRules)
	}

	require.Len(u.T(), aclRules, 3)
	require.Len(u.T(), sshRules, 1)
}

func (u *HujsonSuite) TestAddEmailToGroupHujson() {
	val, err := hujson.Parse([]byte(test.MinimalGroupsExample))
	require.Nil(u.T(), err)

	_, err = AddEmailToGroup(u.ctx, &val, "group:devs", "bonk.flambe@insulator.one")
	require.Nil(u.T(), err)

	require.Equal(u.T(), val.String(), test.ExpectedGroupsResult)
}

func (u *HujsonSuite) TestRemoveEmailFromGroupHujson() {
	val, err := hujson.Parse([]byte(test.ExpectedGroupsResult))
	require.Nil(u.T(), err)

	_, err = RemoveEmailFromGroup(u.ctx, &val, "group:devs", "bonk.flambe@insulator.one")
	require.Nil(u.T(), err)

	require.Equal(u.T(), val.String(), test.MinimalGroupsExample)
}

func (u *HujsonSuite) TestAddEmailToACLHujson() {
	val, err := hujson.Parse([]byte(test.MinimalACLExample))
	require.Nil(u.T(), err)

	// 5737a0c593a5c4d6473966340e2d0261297a605d741689818c3691307df2a613 is the
	// hash of `acceptjohn.degner@insulator.onelogan.saso@insulator.one`
	_, err = AddEmailToRule(
		u.ctx,
		&val,
		RuleKeyACLs,
		"5737a0c593a5c4d6473966340e2d0261297a605d741689818c3691307df2a613",
		"bonk.flambe@insulator.one",
	)
	require.Nil(u.T(), err)

	require.Equal(u.T(), val.String(), test.ExpectedACLResult)
}

func (u *HujsonSuite) TestRemoveEmailFromACLHujson() {
	val, err := hujson.Parse([]byte(test.ExpectedACLResult))
	require.Nil(u.T(), err)

	// 5737a0c593a5c4d6473966340e2d0261297a605d741689818c3691307df2a613 is the
	// hash of `acceptjohn.degner@insulator.onelogan.saso@insulator.one`
	_, err = RemoveEmailFromRule(
		u.ctx,
		&val,
		RuleKeyACLs,
		"5737a0c593a5c4d6473966340e2d0261297a605d741689818c3691307df2a613",
		"bonk.flambe@insulator.one",
	)
	require.Nil(u.T(), err)

	require.Equal(u.T(), val.String(), test.MinimalACLExample)
}
