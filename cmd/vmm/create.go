package vmm

import (
	"fmt"
	"github.com/768bit/promethium/api/client/vms"
	"github.com/768bit/promethium/api/models"
	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"
)

var CreateInstanceCommand = cli.Command{
	Name:  "create",
	Usage: "Create instance.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "auto-start",
		},
		&cli.StringFlag{
			Name:     "name",
			Aliases:  []string{"n"},
			Required: true,
		},
		&cli.Int64Flag{
			Name:     "cpu",
			Aliases:  []string{"c"},
			Required: true,
		},
		&cli.Int64Flag{
			Name:     "mem",
			Aliases:  []string{"m"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "disk-size, d",
			Aliases:  []string{"d"},
			Required: true,
		},
		&cli.StringFlag{
			Name: "net",
		},
		&cli.StringFlag{
			Name:     "image",
			Aliases:  []string{"i"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "kernel-image",
			Aliases: []string{"k"},
		},
		&cli.StringFlag{
			Name:  "storage",
			Value: "default-local",
		},
	},
	Action: func(c *cli.Context) error {
		ds, err := units.FromHumanSize(c.String("disk-size"))
		if err != nil {
			return err
		}
		fmt.Printf("Parsed Size: %s -> %d\n", c.String("disk-size"), ds)
		params := vms.NewCreateVMParams()
		params.SetVMConfig(&models.NewVM{
			Cpus:             c.Int64("cpu"),
			Name:             c.String("name"),
			Memory:           c.Int64("mem"),
			AutoStart:        c.Bool("auto-start"),
			PrimaryNetworkID: "",
			RootDiskSize:     ds,
			FromImage:        c.String("image"),
			KernelImage:      c.String("kernel-image"),
			StorageName:      c.String("storage"),
		})
		resp, err := ApiCli.Vms.CreateVM(params)
		if err != nil {
			return err
		}
		print(resp.Payload.ID.String())
		return nil
	},
}
