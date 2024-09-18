package main

import (
	"testing"

	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/test"
	"github.com/conductorone/baton-sdk/pkg/ustrings"
)

func TestConfigs(t *testing.T) {
	test.ExerciseTestCasesFromExpressions(
		t,
		field.NewConfiguration(ConfigurationFields),
		nil,
		ustrings.ParseFlags,
		[]test.TestCaseFromExpression{
			{
				"",
				false,
				"empty",
			},
			{
				"--api-key 1",
				false,
				"tailnet missing",
			},
			{
				"--tailnet 1",
				false,
				"API key missing",
			},
			{
				"--api-key 1 --tailnet 1",
				true,
				"all",
			},
		},
	)
}
