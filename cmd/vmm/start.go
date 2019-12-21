package vmm

import (
	"github.com/768bit/promethium/api/client/vms"
	"github.com/urfave/cli/v2"
)

var StartInstanceCommand = cli.Command{
	Name:  "start",
	Usage: "Start instance.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "all, a",
		},
	},
	Action: func(c *cli.Context) error {
		params := vms.NewStartVMParams()
		params.SetVMID(c.Args().Get(0))
		_, err := ApiCli.Vms.StartVM(params)
		if err != nil {
			return err
		}
		println("possibly started")
		return nil
	},
}
