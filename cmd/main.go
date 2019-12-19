package main

import (
  "fmt"
  "github.com/768bit/promethium/cmd/daemon"
  "github.com/768bit/promethium/cmd/vmm"
  "gopkg.in/urfave/cli.v1"
  "log"
  "os"
  "strings"
)

var (
  Version   string
  Build     string
  BuildDate string
  GitCommit string
)

func main() {
  cli.VersionFlag = cli.BoolFlag{
    Name:  "version",
    Usage: "Show the version of vcli",
  }

  cli.HelpFlag = cli.BoolFlag{Name: "help"}

  app := cli.NewApp()

  app.Version = fmt.Sprintf("%s  Git Commit: %s  Build Date: %s", Version, GitCommit, strings.Replace(BuildDate, "_", " ", -1))

  app.Commands = []cli.Command{
    daemon.RunDaemonCommand,
    vmm.VmmSubCommand,
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }

}
