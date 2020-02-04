package vmm

import (
	"github.com/768bit/promethium/api/client/vms"
	"github.com/urfave/cli/v2"
)

var ShutdownInstanceCommand = cli.Command{
	Name:  "shutdown",
	Usage: "Gracefully Shutdown instance.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "all, a",
		},
	},
	Action: func(c *cli.Context) error {
		params := vms.NewShutdownVMParams()
		params.SetVMID(c.Args().Get(0))
		_, err := ApiCli.Vms.ShutdownVM(params)
		if err != nil {
			return err
		}
		println("possibly stopped")
		return nil
	},
}
