package img

import (
	"os"
	"time"

	"github.com/docker/go-units"
	"github.com/landoop/tableprinter"
	"github.com/urfave/cli/v2"
)

var ListImagesCommand = cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "List images.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "all, a",
		},
	},
	Action: func(c *cli.Context) error {
		printer := tableprinter.New(os.Stdout)
		list, err := ApiCli.Images.GetImagesList(nil)
		if err != nil {
			return err
		}
		rows := make([][]string, len(list.Payload))
		for i, item := range list.Payload {
			odstr := item.CreatedAt.String()
			t, err := time.Parse("2006-01-02T15:04:05Z", odstr)
			if err == nil {
				d := time.Now().Sub(t)
				odstr = units.HumanDuration(d) + " ago"
			} else {
				println(err.Error())
			}
			rows[i] = []string{item.ID, item.Name, item.Version, item.Architecture, odstr}
		}
		printer.Render([]string{"ID", "OS", "Version", "Arch", "Created"}, rows, nil, false)
		return nil
	},
}
