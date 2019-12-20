package daemon

import (
	"log"

	"github.com/768bit/promethium/lib"
	"github.com/urfave/cli/v2"
)

var RunDaemonCommand = cli.Command{
	Name:  "daemon",
	Usage: "Run the Promethium Daemon.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "foreground",
			Aliases: []string{"f"},
		},
	},
	Action: func(c *cli.Context) error {
		foreground := c.Bool("foreground")
		log.Printf("Running Promethium Daemon...")
		pd, err := lib.NewPromethiumDaemon(foreground)
		if err != nil {
			return err
		}
		if foreground {
			log.Printf("Waiting on foreground...")
			pd.Wait()
		}
		return nil
	},
}
