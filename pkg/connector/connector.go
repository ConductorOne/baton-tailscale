package connector

import (
	"context"
	"io"

	v1 "github.com/conductorone/baton-sdk/pb/c1/config/v1"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-tailscale/pkg/connector/client"
	"google.golang.org/protobuf/types/known/structpb"
)

type Connector struct {
	client *client.Client
}

const (
	SetDeviceAttributeActionID = "tailscale:set-device-attribute"
)

// Implement CustomActionManager interface
func (c *Connector) ListActionSchemas(ctx context.Context) ([]*v2.BatonActionSchema, annotations.Annotations, error) {
	schemas := []*v2.BatonActionSchema{
		{
			Name:        SetDeviceAttributeActionID,
			DisplayName: "Set Device Attribute",
			Description: "Set a custom attribute on all devices for a specific user",
			Arguments: []*v1.Field{
				{
					Name:        "email",
					DisplayName: "User Email",
					Description: "Email address of the user whose devices to update",
					IsRequired:  true,
					Field: &v1.Field_StringField{
						StringField: &v1.StringField{},
					},
				},
				{
					Name:        "attribute_key",
					DisplayName: "Attribute Key",
					Description: "The custom attribute key to set",
					IsRequired:  true,
					Field: &v1.Field_StringField{
						StringField: &v1.StringField{},
					},
				},
				{
					Name:        "attribute_value",
					DisplayName: "Attribute Value",
					Description: "The custom attribute value to set",
					IsRequired:  true,
					Field: &v1.Field_StringField{
						StringField: &v1.StringField{},
					},
				},
			},
		},
	}
	return schemas, nil, nil
}

func (c *Connector) GetActionSchema(ctx context.Context, name string) (*v2.BatonActionSchema, annotations.Annotations, error) {
	// Implementation for getting a specific action schema
	return nil, nil, nil
}

func (c *Connector) InvokeAction(ctx context.Context, name string, args *structpb.Struct) (string, v2.BatonActionStatus, *structpb.Struct, annotations.Annotations, error) {
	// Implementation for invoking the action
	return "", v2.BatonActionStatus_BATON_ACTION_STATUS_PENDING, nil, nil, nil
}

func (c *Connector) GetActionStatus(ctx context.Context, id string) (v2.BatonActionStatus, string, *structpb.Struct, annotations.Annotations, error) {
	// Implementation for getting action status
	return v2.BatonActionStatus_BATON_ACTION_STATUS_PENDING, "", nil, nil, nil
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newACLRuleBuilder(d.client),
		newGroupBuilder(d.client),
		newSSHRuleBuilder(d.client),
		newUserBuilder(d.client),
		newRoleBuilder(d.client),
		newDeviceBuilder(d.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Tailscale Connector",
		Description: "Connector Syncing Tailscale users, groups, and roles",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, apiKey string, tailnet string) (*Connector, error) {
	client, err := client.New(ctx, apiKey, tailnet)
	if err != nil {
		return nil, err
	}
	return &Connector{client: client}, nil
}
