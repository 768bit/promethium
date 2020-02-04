package img

import (
	"github.com/urfave/cli/v2"
	"os"

	"github.com/768bit/promethium/api/client/images"
	"github.com/go-openapi/runtime"
)

var PullImageCommand = cli.Command{
	Name:  "pull",
	Usage: "Pull images from remote to local",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "target-storage",
			Aliases: []string{"t"},
			Value:   "default-local",
		},
	},
	Action: func(c *cli.Context) error {
		//get args (which is path)
		storageTarget := c.String("target-storage")
		alen := c.Args().Len()
		if alen > 0 {
			//ok lets pull these...
			for i := 0; i < alen; i++ {
				isLocal, path := EstablishPathToSource(c.Args().Get(i))
				if isLocal && path != "" {
					brdr, err := os.Open(path)
					if err != nil {
						return err
					}
					params := images.NewPushImageParams()
					params.SetInFileBlob(runtime.NamedReader("inFileBlob", brdr))
					params.SetTargetStorage(&storageTarget)
					resp, err := ApiCli.Images.PushImage(params)
					if err != nil {
						return err
					}
					println(resp.Error())
				} else {
					println("Remote Path", path)
				}
			}
		}

		return nil
	},
}
