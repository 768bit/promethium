package vmm

import (
  "errors"
  "github.com/768bit/promethium/api/client"
  "github.com/768bit/promethium/cmd/common"
  "gopkg.in/urfave/cli.v1"
)

var ApiCli *client.APIClient

var VmmSubCommand = cli.Command{
  Name:  "vmm",
  Aliases: []string{"vm"},
  Usage: "Commands for managing Virtual Machines.",
  Flags: []cli.Flag{
    cli.StringFlag{
      Name:        "host, h",
    },
    cli.IntFlag{
      Name:        "port, p",
    },
    cli.BoolFlag{
      Name:        "tcp, t",
    },
  },
  Before: func(context *cli.Context) error {
    if !context.Bool("tcp") && context.String("host") == "" && context.Int("port") == 0 {
      ApiCli = common.MakeClientUnix()
    } else {
      if !context.Bool("tcp") {
        return errors.New("Must use the --tcp, -t flag if connecting to tcp socket")
      }
      host := context.String("host")
      if host == "" {
        host = "http://127.0.0.1"
      }
      port := context.Int("port")
      if port == 0 {
        port = 8921
      }
      ApiCli = common.MakeClient(host, port, "")
    }
    return nil
  },
  Subcommands: []cli.Command{
    ListInstancesCommand,
  },
}