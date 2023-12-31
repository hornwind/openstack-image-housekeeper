package action

import (
	"context"

	log "github.com/hornwind/openstack-image-keeper/pkg/logging"
	"github.com/urfave/cli/v2"
)

// Action is an interface for all actions.
type Action interface {
	Run(context.Context) error
	Cmd() *cli.Command
}

// toCtx is a wrapper for urfave v2.
func toCtx(a func(context.Context) error) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		ctx := getContextWithFlags(c)

		return a(ctx)
	}
}

func getContextWithFlags(c *cli.Context) context.Context {
	log := log.GetLogger()
	ctx := c.Context
	for _, flagName := range c.FlagNames() {
		g := c.Value(flagName)
		log.WithField("name", flagName).WithField("value", g).Trace("adding flag to action context.Context")
		// log.GetLogger().GetLoggerWithField("name", flagName).WithField("value", g).Trace("adding flag to action context.Context")
		ctx = context.WithValue(ctx, flagName, g) //nolint:staticcheck // weird issue, we won't have any collisions with strings
	}

	ctx = context.WithValue(ctx, "cli", c) //nolint:staticcheck // same

	ctx = context.WithValue(ctx, "firstArg", c.Args().First()) //nolint:staticcheck // same
	ctx = context.WithValue(ctx, "allArgs", c.Args().Slice())  //nolint:staticcheck // same
	return ctx
}
