package img

import (
	"os"
	"strings"
	"time"
	"unsafe"

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
			Name:    "all",
			Aliases: []string{"a"},
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
			var stringTypes []string
			stringTypes = *(*[]string)(unsafe.Pointer(&item.Contains))
			rows[i] = []string{item.ID, item.Name, item.Version, item.Architecture, strings.Join(stringTypes, ","), odstr}
		}
		printer.Render([]string{"ID", "OS", "Version", "Arch", "Contains", "Created"}, rows, nil, false)
		return nil
	},
}
