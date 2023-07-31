package action

import (
	"context"
	"fmt"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var _ Action = (*DeleteByID)(nil)

// DeleteByID is a struct for running 'delete' command.
type DeleteByID struct {
	authOpts *gophercloud.AuthOptions
	provider *gophercloud.ProviderClient
	eo       *gophercloud.EndpointOpts
}

// Run is the main function for 'delete' command.
func (d *DeleteByID) Run(ctx context.Context) error {
	ao, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return err
	}
	d.authOpts = &ao
	d.provider, err = openstack.AuthenticatedClient(*d.authOpts)
	if err != nil {
		return err
	}
	d.eo = &gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	}

	idList, ok := ctx.Value("allArgs").([]string)
	if !ok {
		msg := "Image args list assertion failed"
		err := fmt.Errorf("%s", msg)
		return err
	}

	err = d.deleteImages(ctx, idList)
	return err
}

func (d *DeleteByID) deleteImages(ctx context.Context, idList []string) error {
	client, _ := openstack.NewImageServiceV2(d.provider, *d.eo)

	for _, id := range idList {
		result := images.Delete(client, id)
		log.Debug(result.Result)
		if result.Err != nil {
			return result.Err
		}
	}
	return nil
}

// Cmd returns 'delete' *cli.Command.
func (d *DeleteByID) Cmd() *cli.Command {
	return &cli.Command{
		Name:    "delete",
		Aliases: []string{"del"},
		Usage:   "Delete image by id",
		Action:  toCtx(d.Run),
	}
}
