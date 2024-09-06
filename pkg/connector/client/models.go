package client

import "github.com/tailscale/hujson"

type tailscaleResponseTrimmed struct {
	Value hujson.ValueTrimmed
}

type tailscaleResponse struct {
	Value hujson.Value
}

type Resource struct {
	Id          string `json:"id"`
	DisplayName string `json:"name"`
}
