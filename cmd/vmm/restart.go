package vmm

import (
	"github.com/768bit/promethium/api/client/vms"
	"github.com/urfave/cli/v2"
)

var RestartInstanceCommand = cli.Command{
	Name:  "restart",
	Usage: "Gracefully Restart instance.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "all, a",
		},
	},
	Action: func(c *cli.Context) error {
		params := vms.NewRestartVMParams()
		params.SetVMID(c.Args().Get(0))
		_, err := ApiCli.Vms.RestartVM(params)
		if err != nil {
			return err
		}
		println("possibly restarted")
		return nil
	},
}
