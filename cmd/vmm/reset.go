package vmm

import (
	"github.com/768bit/promethium/api/client/vms"
	"github.com/urfave/cli/v2"
)

var ResetInstanceCommand = cli.Command{
	Name:  "reset",
	Usage: "Forcefully Reset instance.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "all, a",
		},
	},
	Action: func(c *cli.Context) error {
		params := vms.NewResetVMParams()
		params.SetVMID(c.Args().Get(0))
		_, err := ApiCli.Vms.ResetVM(params)
		if err != nil {
			return err
		}
		println("possibly reset")
		return nil
	},
}
