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

// ActionResult stores the result of an action for later retrieval
type ActionResult struct {
	Status  v2.BatonActionStatus
	Message string
	Result  *structpb.Struct
}

const (
	SetDevicesPostureAttributeActionID    = "tailscale:set-device-posture-attribute-on-user-devices"
	DeleteDevicesPostureAttributeActionID = "tailscale:delete-device-posture-attribute-on-user-devices"
)

var SetUsersDevicesPostureAttributeSchema = &v2.BatonActionSchema{
	Name:        SetDevicesPostureAttributeActionID,
	DisplayName: "Set Device Attribute",
	Description: "Set a device posture attribute on all devices for a specific user",
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
			Description: "(Optional) Comment about the device posture attribute set",
			IsRequired:  false,
			Field: &v1.Field_StringField{
				StringField: &v1.StringField{},
			},
		},
		{
			Name:        "expiry_value",
			DisplayName: "Expiry Value",
			Description: "(Optional) Expiry time for the device posture attribute",
			IsRequired:  false,
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
		{
			Name:        "updated_devices",
			DisplayName: "Updated Devices",
			Description: "The devices that had their device posture attribute updated",
			Field:       &v1.Field_StringField{},
		},
		{
			Name:        "device_count",
			DisplayName: "Device Count",
			Description: "The number of devices that had their device posture attribute updated",
			Field:       &v1.Field_IntField{},
		},
		{
			Name:        "total_devices",
			DisplayName: "Total Devices",
			Description: "The total number of devices that were checked",
			Field:       &v1.Field_IntField{},
		},
	},
}

var RemoveUsersDevicesPostureAttributeSchema = &v2.BatonActionSchema{
	Name:        DeleteDevicesPostureAttributeActionID,
	DisplayName: "Remove Device Attribute",
	Description: "Remove a device posture attribute on all devices for a specific user",
	Arguments: []*v1.Field{
		{
			Name:        "email",
			DisplayName: "User Email",
			Description: "Email address of the user whose device(s) will have their device posture attribute removed",
			IsRequired:  true,
			Field: &v1.Field_StringField{
				StringField: &v1.StringField{},
			},
		},
		{
			Name:        "attribute_key",
			DisplayName: "Attribute Key",
			Description: "The device posture attribute key to remove",
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
			Description: "Whether the device resource(s) device posture attribute was removed successfully",
			Field:       &v1.Field_BoolField{},
		},
		{
			Name:        "deleted_devices",
			DisplayName: "Deleted Devices",
			Description: "The devices that had their device posture attribute deleted",
			Field:       &v1.Field_StringField{},
		},
		{
			Name:        "device_count",
			DisplayName: "Device Count",
			Description: "The number of devices that had their device posture attribute deleted",
			Field:       &v1.Field_IntField{},
		},
		{
			Name:        "total_devices",
			DisplayName: "Total Devices",
			Description: "The total number of devices that were checked",
			Field:       &v1.Field_IntField{},
		},
	},
}

// Use the correct CustomActionManager interface methods.
func (c *Connector) ListActionSchemas(ctx context.Context) ([]*v2.BatonActionSchema, annotations.Annotations, error) {
	schemas := []*v2.BatonActionSchema{
		SetUsersDevicesPostureAttributeSchema,
		RemoveUsersDevicesPostureAttributeSchema,
	}
	return schemas, nil, nil
}

func (c *Connector) GetActionSchema(ctx context.Context, name string) (*v2.BatonActionSchema, annotations.Annotations, error) {
	switch name {
	case SetDevicesPostureAttributeActionID:
		return SetUsersDevicesPostureAttributeSchema, nil, nil
	case DeleteDevicesPostureAttributeActionID:
		return RemoveUsersDevicesPostureAttributeSchema, nil, nil
	default:
		return nil, nil, fmt.Errorf("action schema not found: %s", name)
	}
}

func (c *Connector) InvokeAction(ctx context.Context, name string, args *structpb.Struct) (string, v2.BatonActionStatus, *structpb.Struct, annotations.Annotations, error) {
	switch name {
	case SetDevicePostureAttributeActionID:
		return c.performSetDeviceAttribute(ctx, args)
	default:
		return "", v2.BatonActionStatus_BATON_ACTION_STATUS_FAILED, nil, nil, fmt.Errorf("unsupported action: %s", name)
	}
}

func (c *Connector) GetActionStatus(ctx context.Context, id string) (v2.BatonActionStatus, string, *structpb.Struct, annotations.Annotations, error) {
	// Check if we have a stored result for this action ID
	c.actionResultsMutex.Lock()
	defer c.actionResultsMutex.Unlock()

	if result, exists := c.actionResults[id]; exists {
		// Remove the result after successful retrieval to prevent memory leaks
		delete(c.actionResults, id)
		return result.Status, result.Message, result.Result, nil, nil
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
	comment := getStructValue(args, "comment")

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
		expiryTime := time.Now().Add(duration)
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

	// Generate a unique action ID that includes the result
	actionID := fmt.Sprintf("set-device-attribute-%s-%d", email, time.Now().Unix())

	// Store the result for later retrieval
	c.storeActionResult(actionID, v2.BatonActionStatus_BATON_ACTION_STATUS_COMPLETE, "completed", result)

	return actionID, v2.BatonActionStatus_BATON_ACTION_STATUS_COMPLETE, result, nil, nil
}

// storeActionResult stores the result of an action for later retrieval
func (c *Connector) storeActionResult(actionID string, status v2.BatonActionStatus, message string, result *structpb.Struct) {
	c.actionResultsMutex.Lock()
	defer c.actionResultsMutex.Unlock()

	c.actionResults[actionID] = &ActionResult{
		Status:  status,
		Message: message,
		Result:  result,
	}
}

// parseDuration parses duration strings like "30m", "1h", "1d" and returns time.Duration
func parseDuration(durationStr string) (time.Duration, error) {
	// Remove any whitespace
	durationStr = strings.TrimSpace(durationStr)

	// Handle common duration formats
	switch {
	case strings.HasSuffix(durationStr, "m"):
		minutes := strings.TrimSuffix(durationStr, "m")
		if minutes == "" {
			return 0, fmt.Errorf("invalid minutes format")
		}
		return time.ParseDuration(minutes + "m")
	case strings.HasSuffix(durationStr, "h"):
		hours := strings.TrimSuffix(durationStr, "h")
		if hours == "" {
			return 0, fmt.Errorf("invalid hours format")
		}
		return time.ParseDuration(hours + "h")
	case strings.HasSuffix(durationStr, "d"):
		days := strings.TrimSuffix(durationStr, "d")
		if days == "" {
			return 0, fmt.Errorf("invalid days format")
		}
		// Convert days to hours since time.ParseDuration doesn't support days
		// Parse as hours and multiply by 24
		hours, err := time.ParseDuration(days + "h")
		if err != nil {
			return 0, fmt.Errorf("invalid days format: %w", err)
		}
		return hours * 24, nil
	default:
		return 0, fmt.Errorf("unsupported duration format: %s (use m, h, or d suffix)", durationStr)
	}
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
