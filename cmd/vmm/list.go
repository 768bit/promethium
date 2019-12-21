package vmm

import (
	"os"

	"github.com/landoop/tableprinter"
	"github.com/urfave/cli/v2"
)

var ListInstancesCommand = cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "List instances.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "all, a",
		},
	},
	Action: func(c *cli.Context) error {
		printer := tableprinter.New(os.Stdout)
		printer.Render([]string{"ID"}, nil, nil, false)
		list, err := ApiCli.Vms.GetVMList(nil)
		if err != nil {
			return err
		}
		for _, item := range list.Payload {
			printer.RenderRow([]string{item.ID.String()}, nil)
		}
		return nil
	},
}
