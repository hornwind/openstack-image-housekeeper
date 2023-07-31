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
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

var _ Action = (*CleanupByName)(nil)

// CleanupByName is a struct for running 'cleanup' command.
type CleanupByName struct {
	savedImages       map[string]images.Image
	imagesForDeletion map[string]images.Image
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
func (d *CleanupByName) Run(ctx context.Context) error {
	d.savedImages = make(map[string]images.Image, 0)
	d.imagesForDeletion = make(map[string]images.Image, 0)
	ao, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return err
	}
	authOpts := &ao
	provider, err := openstack.AuthenticatedClient(*authOpts)
	if err != nil {
		return err
	}
	eo := &gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	}

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
	client, err := openstack.NewImageServiceV2(provider, *eo)
	if err != nil {
		return err
	}

	log.Printf("Dry-run %t", d.dryRun)

	d.buildLists(imageName, client, listOpts)

	if !d.dryRun {
		log.Printf("Running cleanup for %s", imageName)
		d.cleanupImages(ctx, client)
	}

	return err
}

func (d *CleanupByName) buildLists(name string, client *gophercloud.ServiceClient, listOpts *images.ListOpts) error {
	allPages, err := images.List(client, *listOpts).AllPages()
	if err != nil {
		log.Error(err)
		return err
	}

	imgs, _ := images.ExtractImages(allPages)
	if len(imgs) < 1 {
		return nil
	}

	commits, err := gh.GetNCommitsFromHead(d.scanDepth)
	if err != nil {
		return err
	}

	return d.filterImagesByCommitAndTime(imgs, commits)
}

func (d *CleanupByName) filterImagesByCommitAndTime(imgs []images.Image, commits []string) error {
	latestImg := images.Image{}
	currentCommitImg := images.Image{}

	for step, i := range imgs {
		if i.Visibility == "public" {
			d.savedImages[i.ID] = i
			continue
		}
		if latestImg.ID == "" {
			log.Debugf("latest image on step %d is %s", step, i.ID)
			d.savedImages[i.ID] = i
			latestImg = i
		}

		// latest for current commit
		for _, tag := range i.Tags {

			if slices.Contains(commits, tag) {
				if currentCommitImg.ID == "" {
					log.Debugf("latest image on step %d is %s", step, i.ID)
					currentCommitImg = i
					d.savedImages[i.ID] = i
					continue
				}

				if i.CreatedAt.After(currentCommitImg.CreatedAt) {
					log.Debugln(i.ID, "after", currentCommitImg.ID)
					delete(d.savedImages, currentCommitImg.ID)
					d.imagesForDeletion[currentCommitImg.ID] = currentCommitImg
					d.savedImages[i.ID] = i
					currentCommitImg = i
					log.Debugf("latest image on step %d is %s", step, i.ID)
					continue
				}
				if i.CreatedAt.Before(currentCommitImg.CreatedAt) {
					log.Debugln(i.ID, "before", currentCommitImg.ID)
					delete(d.savedImages, i.ID)
					d.imagesForDeletion[i.ID] = i
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
							delete(d.savedImages, currentCommitImg.ID)
							d.imagesForDeletion[currentCommitImg.ID] = currentCommitImg
							d.savedImages[i.ID] = i
							currentCommitImg = i
							log.Debugf("latest image on step %d is %s", step, i.ID)
							continue
						}
						if tagCommitIdx > currentCommitImgCommitTagIdx {
							log.Debugln(i.ID, "before", currentCommitImg.ID)
							delete(d.savedImages, i.ID)
							d.imagesForDeletion[i.ID] = i
							log.Debugf("latest image on step %d is %s", step, currentCommitImg.ID)
							continue
						}
						if tagCommitIdx == currentCommitImgCommitTagIdx {
							d.savedImages[i.ID] = i
							continue
						}
					}
				}
			}
		}

		if i.CreatedAt.After(latestImg.CreatedAt) {
			log.Debugln(i.ID, "after", latestImg.ID)
			delete(d.savedImages, latestImg.ID)
			d.imagesForDeletion[latestImg.ID] = latestImg
			d.savedImages[i.ID] = i
			latestImg = i
			log.Debugf("latest image on step %d is %s", step, i.ID)
			continue
		}
		if i.CreatedAt.Before(latestImg.CreatedAt) {
			log.Debugln(i.ID, "before", latestImg.ID)
			delete(d.savedImages, i.ID)
			d.imagesForDeletion[i.ID] = i
			log.Debugf("latest image on step %d is %s", step, latestImg.ID)
			continue
		}
		if i.CreatedAt == latestImg.CreatedAt {
			latestImg = i
			d.savedImages[i.ID] = i
		}
	}

	val := make(map[string]interface{}, 6)
	val["savedImages"] = d.savedImages
	val["imagesForDeletion"] = d.imagesForDeletion
	template.Must(template.New("Output").Parse(tplOutput)).Execute(os.Stdout, val)

	return nil
}

func (d *CleanupByName) cleanupImages(ctx context.Context, client *gophercloud.ServiceClient) error {
	for _, img := range d.imagesForDeletion {
		result := images.Delete(client, img.ID)
		log.Printf("Delete image %s", img.ID)
		if result.Err != nil {
			return result.Err
		}
	}

	return nil
}

// Cmd returns 'cleanup' *cli.Command.
func (d *CleanupByName) Cmd() *cli.Command {
	return &cli.Command{
		Name:   "cleanup",
		Usage:  "Cleanup images by name",
		Flags:  d.flags(),
		Action: toCtx(d.Run),
	}
}

// flags return flag set of CLI urfave.
func (d *CleanupByName) flags() []cli.Flag {
	self := []cli.Flag{
		flagScanDepth(&d.scanDepth),
		flagDryRun(&d.dryRun),
	}

	return self
}
