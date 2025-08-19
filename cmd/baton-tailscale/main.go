package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	cfg "github.com/conductorone/baton-tailscale/pkg/config"
	"github.com/conductorone/baton-tailscale/pkg/connector"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var (
	version       = "dev"
	connectorName = "baton-tailscale"
)

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		connectorName,
		getConnector,
		cfg.Configurations,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, tsc *cfg.Tailscale) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	err := field.Validate(cfg.Config, tsc)
	if err != nil {
		return nil, err
	}

	cb, err := connector.New(
		ctx,
		tsc.ApiKey,
		tsc.Tailnet,
		tsc.IgnoreEphemeralDevices,
	)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	conn, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return conn, nil
}
