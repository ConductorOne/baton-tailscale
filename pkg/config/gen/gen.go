package main

import (
	"github.com/conductorone/baton-sdk/pkg/config"
	cfg "github.com/conductorone/baton-tailscale/pkg/config"
)

func main() {
	config.Generate("tailscale", cfg.Config)
}
