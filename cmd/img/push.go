package img

import (
	"github.com/urfave/cli/v2"
)

var PushImageCommand = cli.Command{
	Name:  "push",
	Usage: "Push images to remote",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "target-storage",
			Aliases: []string{"t"},
			Value:   "default-local",
		},
	},
	Action: func(c *cli.Context) error {

		return nil
	},
}
