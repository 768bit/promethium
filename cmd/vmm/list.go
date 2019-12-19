package vmm

import (
  "context"
  "github.com/landoop/tableprinter"
  "gopkg.in/urfave/cli.v1"
  "os"
)

var ListInstancesCommand = cli.Command{
  Name:  "list",
  Aliases: []string{"ls"},
  Usage: "List instances.",
  Flags: []cli.Flag{
    cli.BoolFlag{
      Name:        "all, a",
    },
  },
  Action: func(c *cli.Context) error {
    printer := tableprinter.New(os.Stdout)
    printer.Render([]string{"ID"}, nil, nil, false)
    list, _, err := ApiCli.VmsApi.GetVMList(context.Background(), nil)
    if err != nil {
      return err
    }
    for _, item := range list {
      printer.RenderRow([]string{item.Id}, nil)
    }
    return nil
  },
}