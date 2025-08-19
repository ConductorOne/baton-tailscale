package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-tailscale/pkg/connector/client"
)

type deviceBuilder struct {
	resourceType           *v2.ResourceType
	client                 *client.Client
	ignoreEphemeralDevices bool
}

func (d *deviceBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return d.resourceType
}

func deviceResource(ctx context.Context, device *client.Device, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	return rs.NewResource(
		device.Name,
		deviceResourceType,
		device.ID,
		rs.WithParentResourceID(parentResourceID),
		rs.WithAppTrait(
			rs.WithAppProfile(
				map[string]interface{}{
					"device_id":   device.ID,
					"device_name": device.Name,
					"login":       device.User,
					"email":       device.User,
					"authorized":  device.Authorized,
				},
			),
		),
	)
}

func (d *deviceBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var rv []*v2.Resource
	devices, _, err := d.client.GetDevices(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	for _, device := range devices {
		deviceCopy := device
		
		// Skip ephemeral devices if configured to ignore them
		if d.ignoreEphemeralDevices && deviceCopy.IsEphemeral {
			continue
		}
		
		dr, err := deviceResource(ctx, &deviceCopy, parentResourceID)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, dr)
	}

	return rv, "", nil, nil
}

func (d *deviceBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (d *deviceBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newDeviceBuilder(client *client.Client, ignoreEphemeralDevices bool) *deviceBuilder {
	return &deviceBuilder{
		resourceType:           deviceResourceType,
		client:                 client,
		ignoreEphemeralDevices: ignoreEphemeralDevices,
	}
}
