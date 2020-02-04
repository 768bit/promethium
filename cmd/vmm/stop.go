package vmm

import (
	"github.com/768bit/promethium/api/client/vms"
	"github.com/urfave/cli/v2"
)

var StopInstanceCommand = cli.Command{
	Name:  "stop",
	Usage: "Stop instance.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "all, a",
		},
	},
	Action: func(c *cli.Context) error {
		params := vms.NewStopVMParams()
		params.SetVMID(c.Args().Get(0))
		_, err := ApiCli.Vms.StopVM(params)
		if err != nil {
			return err
		}
		println("possibly stopped")
		return nil
	},
}
