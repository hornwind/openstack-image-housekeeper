package action

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	gh "github.com/hornwind/openstack-image-keeper/pkg/git-history"
	log "github.com/hornwind/openstack-image-keeper/pkg/logging"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

var _ Action = (*CleanupByName)(nil)

// CleanupByName is a struct for running 'cleanup' command.
type CleanupByName struct {
	savedImages       map[string]images.Image
	imagesForDeletion map[string]images.Image
	loglevel          string
	scanDepth         int
	dryRun            bool
}

var (
	tplOutput = `Saved images:
{{- range $key, $val := .savedImages }}
  {{ $val.ID }}
{{- end }}

Images for deletion:
{{- range $key, $val := .imagesForDeletion }}
  {{ $val.ID }}
{{- end }}
{{- print "\n" }}
`
)

// Run is the main function for 'cleanup' command.
func (c *CleanupByName) Run(ctx context.Context) error {
	log := log.GetLogger()
	if err := log.SetLogLevel(c.loglevel); err != nil {
		return err
	}

	c.savedImages = make(map[string]images.Image, 0)
	c.imagesForDeletion = make(map[string]images.Image, 0)

	imageName, ok := ctx.Value("firstArg").(string)
	if !ok || imageName == "" {
		msg := "Image name arg assertion failed"
		err := fmt.Errorf("%s", msg)
		return err
	}
	listOpts := &images.ListOpts{
		Owner: os.Getenv("OS_PROJECT_ID"),
		Name:  imageName,
	}

	client, err := c.getImageServiceClient()
	if err != nil {
		return err
	}

	log.Infof("Dry-run %t", c.dryRun)
	if err = c.buildLists(imageName, client, listOpts); err != nil {
		return err
	}

	if !c.dryRun {
		log.Infof("Running cleanup for %s", imageName)
		return c.cleanupImages(ctx, client)
	}

	return nil
}

func (c *CleanupByName) getImageServiceClient() (*gophercloud.ServiceClient, error) {
	ao, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}
	authOpts := &ao
	provider, err := openstack.AuthenticatedClient(*authOpts)
	if err != nil {
		return nil, err
	}
	eo := &gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	}

	return openstack.NewImageServiceV2(provider, *eo)
}

func (c *CleanupByName) buildLists(name string, client *gophercloud.ServiceClient, listOpts *images.ListOpts) error {
	log := log.GetLogger()

	allPages, err := images.List(client, *listOpts).AllPages()
	if err != nil {
		log.Error(err)
		return err
	}

	imgs, _ := images.ExtractImages(allPages)
	if len(imgs) < 1 {
		return nil
	}

	commits, err := gh.GetNCommitsFromHead(c.scanDepth)
	if err != nil {
		return err
	}

	return c.filterImagesByCommitAndTime(imgs, commits)
}

func (c *CleanupByName) filterImagesByCommitAndTime(imgs []images.Image, commits []string) error {
	log := log.GetLogger()
	latestImg := images.Image{}
	currentCommitImg := images.Image{}

	for step, i := range imgs {
		if i.Visibility == "public" {
			c.savedImages[i.ID] = i
			continue
		}
		if latestImg.ID == "" {
			log.Debugf("latest image on step %d is %s", step, i.ID)
			c.savedImages[i.ID] = i
			latestImg = i
		}

		// latest for current commit
		for _, tag := range i.Tags {

			if slices.Contains(commits, tag) {
				if currentCommitImg.ID == "" {
					log.Debugf("latest image on step %d is %s", step, i.ID)
					currentCommitImg = i
					c.savedImages[i.ID] = i
					continue
				}

				if i.CreatedAt.After(currentCommitImg.CreatedAt) {
					log.Debugln(i.ID, "after", currentCommitImg.ID)
					delete(c.savedImages, currentCommitImg.ID)
					c.imagesForDeletion[currentCommitImg.ID] = currentCommitImg
					c.savedImages[i.ID] = i
					currentCommitImg = i
					log.Debugf("latest image on step %d is %s", step, i.ID)
					continue
				}
				if i.CreatedAt.Before(currentCommitImg.CreatedAt) {
					log.Debugln(i.ID, "before", currentCommitImg.ID)
					delete(c.savedImages, i.ID)
					c.imagesForDeletion[i.ID] = i
					log.Debugf("latest image on step %d is %s", step, currentCommitImg.ID)
					continue
				}

				if i.CreatedAt == currentCommitImg.CreatedAt {
					tagCommitIdx := slices.Index(commits, tag)

					for _, t := range currentCommitImg.Tags {
						currentCommitImgCommitTagIdx := slices.Index(commits, t)
						if currentCommitImgCommitTagIdx == -1 {
							continue
						}

						if tagCommitIdx < currentCommitImgCommitTagIdx {
							log.Debugln(i.ID, "after", currentCommitImg.ID)
							delete(c.savedImages, currentCommitImg.ID)
							c.imagesForDeletion[currentCommitImg.ID] = currentCommitImg
							c.savedImages[i.ID] = i
							currentCommitImg = i
							log.Debugf("latest image on step %d is %s", step, i.ID)
							continue
						}
						if tagCommitIdx > currentCommitImgCommitTagIdx {
							log.Debugln(i.ID, "before", currentCommitImg.ID)
							delete(c.savedImages, i.ID)
							c.imagesForDeletion[i.ID] = i
							log.Debugf("latest image on step %d is %s", step, currentCommitImg.ID)
							continue
						}
						if tagCommitIdx == currentCommitImgCommitTagIdx {
							c.savedImages[i.ID] = i
							continue
						}
					}
				}
			}
		}

		if i.CreatedAt.After(latestImg.CreatedAt) {
			log.Debugln(i.ID, "after", latestImg.ID)
			delete(c.savedImages, latestImg.ID)
			c.imagesForDeletion[latestImg.ID] = latestImg
			c.savedImages[i.ID] = i
			latestImg = i
			log.Debugf("latest image on step %d is %s", step, i.ID)
			continue
		}
		if i.CreatedAt.Before(latestImg.CreatedAt) {
			log.Debugln(i.ID, "before", latestImg.ID)
			delete(c.savedImages, i.ID)
			c.imagesForDeletion[i.ID] = i
			log.Debugf("latest image on step %d is %s", step, latestImg.ID)
			continue
		}
		if i.CreatedAt == latestImg.CreatedAt {
			latestImg = i
			c.savedImages[i.ID] = i
		}
	}

	val := make(map[string]interface{}, 6)
	val["savedImages"] = c.savedImages
	val["imagesForDeletion"] = c.imagesForDeletion
	template.Must(template.New("Output").Parse(tplOutput)).Execute(os.Stdout, val) //nolint:errcheck

	return nil
}

func (c *CleanupByName) cleanupImages(ctx context.Context, client *gophercloud.ServiceClient) error {
	log := log.GetLogger()
	for _, img := range c.imagesForDeletion {
		result := images.Delete(client, img.ID)
		log.Infof("Delete image %s", img.ID)
		if result.Err != nil {
			return result.Err
		}
	}

	return nil
}

// Cmd returns 'cleanup' *cli.Command.
func (c *CleanupByName) Cmd() *cli.Command {
	return &cli.Command{
		Name:   "cleanup",
		Usage:  "Cleanup images by name",
		Flags:  c.flags(),
		Action: toCtx(c.Run),
	}
}

// flags return flag set of CLI urfave.
func (c *CleanupByName) flags() []cli.Flag {
	self := []cli.Flag{
		flagScanDepth(&c.scanDepth),
		flagDryRun(&c.dryRun),
		flagLogLevel(&c.loglevel),
	}

	return self
}
