package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/768bit/promethium/cmd/daemon"
	"github.com/768bit/promethium/cmd/img"
	"github.com/768bit/promethium/cmd/vmm"
	"github.com/urfave/cli/v2"
)

var (
	Version   string
	Build     string
	BuildDate string
	GitCommit string
)

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version",
		Usage: "Show the version of vcli",
	}

	cli.HelpFlag = &cli.BoolFlag{Name: "help"}

	app := cli.NewApp()

	app.Version = fmt.Sprintf("%s  Git Commit: %s  Build Date: %s", Version, GitCommit, strings.Replace(BuildDate, "_", " ", -1))

	app.Commands = []*cli.Command{
		&InstallCommand,
		&daemon.RunDaemonCommand,
		&vmm.VmmSubCommand,
		&img.ImagesSubCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
