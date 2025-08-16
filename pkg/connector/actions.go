package connector

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "github.com/conductorone/baton-sdk/pb/c1/config/v1"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-tailscale/pkg/connector/client"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	SetDeviceAttributeActionID = "tailscale:set-device-attribute"
)

// Use the correct CustomActionManager interface methods
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
				{
					Name:        "expiry_value",
					DisplayName: "Expiry Value",
					Description: "The custom attribute value to set",
					IsRequired:  false,
					Field: &v1.Field_StringField{
						StringField: &v1.StringField{},
					},
				},
				{
					Name:        "comment_value",
					DisplayName: "Comment Value",
					Description: "The custom attribute value to set",
					IsRequired:  false,
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
	switch name {
	case SetDeviceAttributeActionID:
		return &v2.BatonActionSchema{
			Name:        SetDeviceAttributeActionID,
			DisplayName: "Set Device Attribute",
			Description: "Set a custom attribute on all devices for a specific user",
			Arguments: []*v1.Field{
				{
					Name:        "email",
					DisplayName: "User Email",
					Description: "Email address of the user whose device(s) will have their device posture attribute updated",
					IsRequired:  true,
					Field: &v1.Field_StringField{
						StringField: &v1.StringField{},
					},
				},
				{
					Name:        "attribute_key",
					DisplayName: "Attribute Key",
					Description: "The device posture attribute key to set",
					IsRequired:  true,
					Field: &v1.Field_StringField{
						StringField: &v1.StringField{},
					},
				},
				{
					Name:        "attribute_value",
					DisplayName: "Attribute Value",
					Description: "The device posture attribute value to set",
					IsRequired:  true,
					Field: &v1.Field_StringField{
						StringField: &v1.StringField{},
					},
				},
				{
					Name:        "comment_value",
					DisplayName: "Comment Value",
					Description: "A comment about the device posture attribute set",
					IsRequired:  true,
					Field: &v1.Field_StringField{
						StringField: &v1.StringField{},
					},
				},
			},
			ReturnTypes: []*v1.Field{
				{
					Name:        "success",
					DisplayName: "Success",
					Description: "Whether the device resource(s) device posture attribute was updated successfully",
					Field:       &v1.Field_BoolField{},
				},
			},
		}, nil, nil
	default:
		return nil, nil, fmt.Errorf("action schema not found: %s", name)
	}
}

func (c *Connector) InvokeAction(ctx context.Context, name string, args *structpb.Struct) (string, v2.BatonActionStatus, *structpb.Struct, annotations.Annotations, error) {
	switch name {
	case SetDeviceAttributeActionID:
		return c.performSetDeviceAttribute(ctx, args)
	default:
		return "", v2.BatonActionStatus_BATON_ACTION_STATUS_FAILED, nil, nil, fmt.Errorf("unsupported action: %s", name)
	}
}

func (c *Connector) GetActionStatus(ctx context.Context, id string) (v2.BatonActionStatus, string, *structpb.Struct, annotations.Annotations, error) {
	// For now, we'll return a simple implementation
	// In a real-world scenario, you might want to store action status in a database or cache

	// Check if the ID matches any known action patterns
	if strings.HasPrefix(id, "set-device-attribute-") {
		// This is a set device attribute action
		// For simplicity, we'll assume it completed successfully
		result := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"success": structpb.NewBoolValue(true),
				"message": structpb.NewStringValue("Device attribute update completed successfully"),
			},
		}
		return v2.BatonActionStatus_BATON_ACTION_STATUS_COMPLETE, "completed", result, nil, nil
	}

	// If we don't recognize the ID, return unknown status
	return v2.BatonActionStatus_BATON_ACTION_STATUS_UNKNOWN, "", nil, nil, fmt.Errorf("action status not found: %s", id)
}

func (c *Connector) performSetDeviceAttribute(ctx context.Context, args *structpb.Struct) (string, v2.BatonActionStatus, *structpb.Struct, annotations.Annotations, error) {
	// Extract input fields from structpb.Struct
	email := getStructValue(args, "email")
	attributeKey := getStructValue(args, "attribute_key")
	attributeValue := getStructValue(args, "attribute_value")
	expiryValue := getStructValue(args, "expiry_value")
	comment := getStructValue(args, "comment_value")

	if email == "" || attributeKey == "" || attributeValue == "" {
		return "", v2.BatonActionStatus_BATON_ACTION_STATUS_FAILED, nil, nil, fmt.Errorf("email, attribute_key, and attribute_value are required")
	}

	// Get all devices for the tailnet
	devices, _, err := c.client.GetDevices(ctx)
	if err != nil {
		return "", v2.BatonActionStatus_BATON_ACTION_STATUS_FAILED, nil, nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// Filter devices by user email
	var userDevices []client.Device
	for _, device := range devices {
		if device.User == email {
			userDevices = append(userDevices, device)
		}
	}

	if len(userDevices) == 0 {
		result := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"message": structpb.NewStringValue(fmt.Sprintf("No devices found for user: %s", email)),
			},
		}
		return "no-devices", v2.BatonActionStatus_BATON_ACTION_STATUS_COMPLETE, result, nil, nil
	}

	// Set custom attribute on each device
	var updatedDevices []string
	var errors []string

	// Parse expiry value if provided and convert to RFC3339 timestamp
	var expiryTimestamp string
	if expiryValue != "" {
		duration, err := parseDuration(expiryValue)
		if err != nil {
			return "", v2.BatonActionStatus_BATON_ACTION_STATUS_FAILED, nil, nil, fmt.Errorf("invalid expiry value '%s': %w", expiryValue, err)
		}
		expiryTime := time.Now().UTC().Add(duration)
		expiryTimestamp = expiryTime.Format(time.RFC3339)
	}

	for _, device := range userDevices {
		// POST /api/v2/device/{deviceId}/attributes/{attributeKey}
		err := c.client.SetDeviceAttribute(ctx, device.ID, "custom:"+attributeKey, attributeValue, expiryTimestamp, comment)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to set attribute on device %s (%s): %v", device.Name, device.ID, err))
		} else {
			updatedDevices = append(updatedDevices, fmt.Sprintf("%s (%s)", device.Name, device.ID))
		}
	}

	// Build result
	result := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"updated_devices": structpb.NewStringValue(strings.Join(updatedDevices, ", ")),
			"device_count":    structpb.NewNumberValue(float64(len(updatedDevices))),
		},
	}

	if len(errors) > 0 {
		result.Fields["errors"] = structpb.NewStringValue(strings.Join(errors, "; "))
	}

	return "completed", v2.BatonActionStatus_BATON_ACTION_STATUS_COMPLETE, result, nil, nil
}

// parseDuration parses duration strings like "30m", "1h", "1d" and returns time.Duration
func parseDuration(durationStr string) (time.Duration, error) {
	// Remove any whitespace
	durationStr = strings.TrimSpace(durationStr)

	// Check for negative values
	if strings.HasPrefix(durationStr, "-") {
		return 0, fmt.Errorf("negative duration not allowed: %s", durationStr)
	}

	// Handle days as a special case since time.ParseDuration doesn't support days
	if strings.HasSuffix(durationStr, "d") {
		days := strings.TrimSuffix(durationStr, "d")
		if days == "" {
			return 0, fmt.Errorf("invalid days format")
		}
		// Convert days to hours (24 hours per day)
		hours, err := time.ParseDuration(days + "h")
		if err != nil {
			return 0, fmt.Errorf("invalid days format: %w", err)
		}
		return hours * 24, nil
	}

	// For everything else (minutes, hours, etc.), use time.ParseDuration directly
	return time.ParseDuration(durationStr)
}

func getStructValue(args *structpb.Struct, fieldName string) string {
	if args == nil || args.Fields == nil {
		return ""
	}
	if field, exists := args.Fields[fieldName]; exists && field != nil {
		return field.GetStringValue()
	}
	return ""
}
