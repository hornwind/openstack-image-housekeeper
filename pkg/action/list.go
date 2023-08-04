package action

import (
	"context"
	"os"
	"text/template"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	log "github.com/hornwind/openstack-image-keeper/pkg/logging"
	"github.com/urfave/cli/v2"
)

var _ Action = (*List)(nil)

// List is a struct for running 'list' command.
type List struct {
	loglevel string
}

var listOutputTpl string = `Name: {{ .Name }}
ID: {{ $.ID }}
CreatedAt: {{ .CreatedAt }}
Protected {{ .Protected }}
Hidden {{ .Hidden }}
{{ with .Tags -}}
Tags:
{{- range $key, $val := . }}
  {{ $val }}
{{- end }}
{{- end }}
{{ with .Properties -}}
Properties:
{{- range $key, $val := . }}
  {{ $key }}:{{ $val }}
{{- end }}
{{- end }}
`

// Run is the main function for 'list' command.
func (l *List) Run(ctx context.Context) error {
	log := log.GetLogger()
	if err := log.SetLogLevel(l.loglevel); err != nil {
		return err
	}
	ao, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return err
	}
	provider, err := openstack.AuthenticatedClient(ao)
	if err != nil {
		return err
	}
	listOpts := &images.ListOpts{
		Owner: os.Getenv("OS_PROJECT_ID"),
	}
	eo := &gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	}

	client, _ := openstack.NewImageServiceV2(provider, *eo)
	allPages, err := images.List(client, listOpts).AllPages()
	imgs, _ := images.ExtractImages(allPages)

	if err != nil {
		log.Error(err)
		return err
	}

	err = l.ListImages(ctx, imgs)
	return err
}

func (l *List) ListImages(ctx context.Context, imgs []images.Image) error {
	t := template.Must(template.New("Image").Parse(listOutputTpl))

	for _, i := range imgs {
		val := make(map[string]interface{}, 6)
		val["Name"] = i.Name
		val["ID"] = i.ID
		val["CreatedAt"] = i.CreatedAt
		val["Protected"] = i.Protected
		val["Hidden"] = i.Hidden
		val["Tags"] = i.Tags

		err := t.ExecuteTemplate(os.Stdout, "Image", val)
		if err != nil {
			return err
		}
	}
	return nil
}

// Cmd returns 'list' *cli.Command.
func (l *List) Cmd() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List of available images",
		Flags:   l.flags(),
		Action:  toCtx(l.Run),
	}
}

// flags return flag set of CLI urfave.
func (l *List) flags() []cli.Flag {
	self := []cli.Flag{
		flagLogLevel(&l.loglevel),
	}

	return self
}
