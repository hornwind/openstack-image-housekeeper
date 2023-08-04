package main

import (
	"fmt"
	"os"

	"github.com/hornwind/openstack-image-keeper/pkg/action"
	log "github.com/hornwind/openstack-image-keeper/pkg/logging"
	"github.com/urfave/cli/v2"
)

//nolintlint:gochecknoglobals // we need a global list of commands
var commands = []*cli.Command{
	new(action.List).Cmd(),
	new(action.DeleteByID).Cmd(),
	new(action.CleanupByName).Cmd(),
	new(action.Publication).Cmd(),
	version(),
}

func main() {
	log := log.GetLogger()
	c := CreateApp()

	defer recoverPanic()

	if err := c.Run(os.Args); err != nil {
		log.Fatal(err) //nolint:gocritic // we try to recover panics, not regular command errors
	}
}

func recoverPanic() {
	log := log.GetLogger()
	if r := recover(); r != nil {
		switch r.(type) {
		case CommandNotFoundError:
			log.Error(r)
			log.Logger.Exit(127)
		default:
			log.Panic(r)
		}
	}
}

// CreateApp creates *cli.App with all commands.
func CreateApp() *cli.App {
	c := cli.NewApp()

	c.Suggest = true
	c.Usage = "Image management for openstack"
	c.Description = "This tool helps you housekeeping your openstack images.!\n"
	c.Commands = commands
	c.CommandNotFound = command404

	return c
}

// CommandNotFoundError is return when CLI command is not found.
type CommandNotFoundError struct {
	Command string
}

func (e CommandNotFoundError) Error() string {
	return fmt.Sprintf("Command %q not found", e.Command)
}

func command404(c *cli.Context, s string) {
	err := CommandNotFoundError{
		Command: s,
	}
	panic(err)
}

func version() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"ver"},
		Usage:   "Show shorts version",
		Action: func(c *cli.Context) error {
			fmt.Println("develop") //nolint:forbidigo // we need to use fmt.Println here

			return nil
		},
	}
}
