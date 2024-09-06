package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	ApiKeyField = field.StringField(
		"api-key",
		field.WithRequired(true),
		field.WithDescription("Tailscale API Key"),
	)
	TailnetField = field.StringField(
		"tailnet",
		field.WithRequired(true),
		field.WithDescription("Tailscale Tailnet"),
	)

	// ConfigurationFields defines the external configuration required for the connector to run.
	ConfigurationFields = []field.SchemaField{
		ApiKeyField,
		TailnetField,
	}
	Configurations = field.NewConfiguration(ConfigurationFields)
)
