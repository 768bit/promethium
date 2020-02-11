package daemon

import (
	"log"

	"github.com/768bit/promethium/lib"
	"github.com/768bit/promethium/lib/config"
	"github.com/urfave/cli/v2"
)

func init() {
}

var RunDaemonCommand = cli.Command{
	Name:  "daemon",
	Usage: "Run the Promethium Daemon.",
	//SkipFlagParsing: true,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "foreground",
			Aliases: []string{"f"},
		},
		&cli.StringFlag{
			Name:    "config-path",
			Aliases: []string{"c"},
			Value:   config.PROMETHIUM_CONFIG_DIR,
		},
	},
	Action: func(c *cli.Context) error {
		foreground := c.Bool("foreground")
		log.Printf("Running Promethium Daemon...")
		if !foreground {
			//adaptflag.Adapt()
			//flag.Parse()
		}
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
