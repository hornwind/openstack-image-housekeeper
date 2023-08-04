package action

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	log "github.com/hornwind/openstack-image-keeper/pkg/logging"
	"github.com/urfave/cli/v2"
)

type Publication struct {
	authOpts  *gophercloud.AuthOptions
	provider  *gophercloud.ProviderClient
	client    *gophercloud.ServiceClient
	eo        *gophercloud.EndpointOpts
	loglevel  string
	dryRun    bool
	protected bool
	hidden    bool
}

var (
	tplPublishOutput = `These images will be published:
{{- range $key, $val := .imgForPublication }}
  {{ $val.ID }}
{{- end }}

These images will be private:
	{{- range $key, $val := .imagesForUnpublish }}
  {{ $val.ID }}
{{- end }}
{{- print "\n" }}
`
)

func (p *Publication) Run(ctx context.Context) error {
	log := log.GetLogger()
	if err := log.SetLogLevel(p.loglevel); err != nil {
		return err
	}
	if err := p.configureClient(); err != nil {
		return err
	}

	imgUUID, ok := ctx.Value("firstArg").(string)
	if !ok {
		msg := "Image id assertion failed"
		err := fmt.Errorf("%s", msg)
		return err
	}

	imagesWithSameName, err := p.getImagesWithSameName(imgUUID)
	if err != nil {
		return err
	}

	if p.dryRun {
		return p.dryRunAnnounce(imgUUID, imagesWithSameName)
	}

	if err := p.updateImagesWithSameName(imagesWithSameName, images.ImageVisibilityPrivate, false, false); err != nil {
		return err
	}
	if err := p.setProtected(imgUUID, p.protected); err != nil {
		return err
	}
	if err := p.setVisibility(imgUUID, images.ImageVisibilityPublic); err != nil {
		return err
	}
	if err := p.setHidden(imgUUID, p.hidden); err != nil {
		return err
	}

	return nil
}

func (p *Publication) dryRunAnnounce(uuid string, imagesWithSameName []images.Image) error {
	imgForUnpublish := []images.Image{}
	for _, img := range imagesWithSameName {
		i := img
		if i.ID == uuid {
			continue
		}
		if i.Protected || !i.Hidden || i.Visibility == images.ImageVisibilityPublic {
			imgForUnpublish = append(imgForUnpublish, i)
		}
	}

	imgForPublication, err := p.getImagesByUUID(uuid)
	if err != nil {
		return err
	}

	val := make(map[string]interface{}, 6)
	val["imgForPublication"] = imgForPublication
	val["imagesForUnpublish"] = imgForUnpublish
	template.Must(template.New("Output").Parse(tplPublishOutput)).Execute(os.Stdout, val) //nolint:errcheck

	return nil
}

func (p *Publication) getImagesByUUID(uuid string) ([]images.Image, error) {
	targetImageListOpts := &images.ListOpts{
		Owner: os.Getenv("OS_PROJECT_ID"),
		ID:    uuid,
	}

	allPages, err := images.List(p.client, targetImageListOpts).AllPages()
	if err != nil {
		return nil, err
	}
	imgs, err := images.ExtractImages(allPages)
	if err != nil {
		return nil, err
	}
	return imgs, nil
}

func (p *Publication) getImagesWithSameName(uuid string) ([]images.Image, error) {
	imgs, err := p.getImagesByUUID(uuid)
	if err != nil {
		return nil, err
	}

	imagesByNameOpts := &images.ListOpts{
		Owner: os.Getenv("OS_PROJECT_ID"),
		Name:  imgs[0].Name,
	}

	allImagesWithSameName, err := images.List(p.client, imagesByNameOpts).AllPages()
	if err != nil {
		return nil, err
	}
	imagesWithSameName, err := images.ExtractImages(allImagesWithSameName)
	if err != nil {
		return nil, err
	}

	return imagesWithSameName, nil
}

func (p *Publication) setVisibility(id string, visibility images.ImageVisibility) error {
	image := images.Update(p.client, id, images.UpdateOpts{
		images.UpdateVisibility{Visibility: visibility},
	})
	return image.Err
}

func (p *Publication) setHidden(id string, hidden bool) error {
	image := images.Update(p.client, id, images.UpdateOpts{
		images.ReplaceImageHidden{NewHidden: hidden},
	})
	return image.Err
}

func (p *Publication) setProtected(id string, protected bool) error {
	image := images.Update(p.client, id, images.UpdateOpts{
		images.ReplaceImageProtected{NewProtected: protected},
	})
	return image.Err
}

func (p *Publication) updateImagesWithSameName(imageList []images.Image, visibility images.ImageVisibility, protected, hidden bool) error {
	for _, image := range imageList {
		i := image

		if err := p.setProtected(i.ID, protected); err != nil {
			return err
		}
		if err := p.setVisibility(i.ID, visibility); err != nil {
			return err
		}
		if err := p.setHidden(i.ID, hidden); err != nil {
			return err
		}
	}
	return nil
}

func (p *Publication) configureClient() error {
	ao, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return err
	}
	p.authOpts = &ao
	p.provider, err = openstack.AuthenticatedClient(*p.authOpts)
	if err != nil {
		return err
	}
	p.eo = &gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	}
	p.client, err = openstack.NewImageServiceV2(p.provider, *p.eo)
	if err != nil {
		return err
	}
	return nil
}

// function Cmd
func (p *Publication) Cmd() *cli.Command {
	return &cli.Command{
		Name:   "publish",
		Usage:  "Publication image by id",
		Flags:  p.flags(),
		Action: toCtx(p.Run),
	}
}

// flags return flag set of CLI urfave.
func (p *Publication) flags() []cli.Flag {
	self := []cli.Flag{
		flagDryRun(&p.dryRun),
		flagProtected(&p.protected),
		flagHidden(&p.hidden),
		flagLogLevel(&p.loglevel),
	}

	return self
}
