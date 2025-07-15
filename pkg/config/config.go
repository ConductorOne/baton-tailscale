package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	TailnetField = field.StringField(
		"tailnet",
		field.WithDisplayName("Tailnet"),
		field.WithDescription("Tailscale Tailnet"),
		field.WithRequired(true),
	)

	ApiKeyField = field.StringField(
		"api-key",
		field.WithDisplayName("API Key"),
		field.WithDescription("Tailscale API Key"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)
	// ConfigurationFields defines the external configuration required for the connector to run.
	ConfigurationFields = []field.SchemaField{
		ApiKeyField,
		TailnetField,
	}

	Configurations     = field.NewConfiguration(ConfigurationFields)
	FieldRelationships = []field.SchemaFieldRelationship{}
)

//go:generate go run ./gen
var Config = field.NewConfiguration(
	ConfigurationFields,
	field.WithConstraints(FieldRelationships...),
	field.WithConnectorDisplayName("Tailscale v2"),
	field.WithHelpUrl("/docs/baton/tailscale-v2"),
	field.WithIconUrl("/static/app-icons/tailscale.svg"),
)
