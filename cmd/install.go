package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/768bit/promethium/lib/config"
	"github.com/768bit/vutils"
	"github.com/urfave/cli/v2"
)

var InstallCommand = cli.Command{
	Name:    "install",
	Aliases: []string{"ls"},
	Usage:   "Install promethium to your system",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "binary-only",
		},
		&cli.StringFlag{
			Name:    "install-directory",
			Aliases: []string{"d"},
			Value:   "/opt/promethium",
		},
		&cli.StringFlag{
			Name:    "config-directory",
			Aliases: []string{"c"},
			Value:   config.PROMETHIUM_CONFIG_DIR,
		},
		&cli.StringFlag{
			Name:    "socket-path",
			Aliases: []string{"s"},
			Value:   config.PROMETHIUM_SOCKET_PATH,
		},
		&cli.StringFlag{
			Name:    "daemon-user",
			Aliases: []string{"u"},
			Value:   "promethium",
		},
		&cli.StringFlag{
			Name:    "daemon-group",
			Aliases: []string{"g"},
			Value:   "promethium",
		},
		&cli.StringFlag{
			Name:    "jail-user",
			Aliases: []string{"j"},
			Value:   "promethium_jail",
		},
		&cli.StringFlag{
			Name:    "jail-group",
			Aliases: []string{"r"},
			Value:   "promethium_jail",
		},
		&cli.BoolFlag{
			Name:  "enable-http",
			Value: false,
			Usage: "Set this flag to enable setting --http-bind-address, -b and --http-bind-port, -p",
		},
		&cli.StringFlag{
			Name:    "http-bind-address, b",
			Aliases: []string{"b"},
			Value:   "0.0.0.0",
			Usage:   "Can only be set when --enable-http flag is set",
		},
		&cli.IntFlag{
			Name:    "http-bind-port, p",
			Aliases: []string{"p"},
			Value:   8921,
			Usage:   "Can only be set when --enable-http flag is set",
		},
	},
	Action: func(c *cli.Context) error {

		if os.Getuid() != 0 {
			return errors.New("Unable to run installation - this needs to be run as root using sudo or from a root shell.")
		}

		configPath := c.String("config-directory")
		installPath := c.String("install-directory")
		socketPath := c.String("socket-path")
		daemonUser := c.String("daemon-user")
		daemonGroup := c.String("daemon-user")
		jailUser := c.String("jail-group")
		jailGroup := c.String("jail-group")
		enableHttp := c.Bool("enable-http")
		httpAddr := c.String("http-bind-address")
		httpPort := c.Int("http-bind-port")

		fullDaemonConfigPath := filepath.Join(configPath, "daemon.json")

		if !c.Bool("binary-only") {

			//install everything needed for the service to run...

			//this will need to be run with appropriate privs and will install the service, the config etc...

			//will also check group membership for certain elements...

			//first of all gen the config...

			if vutils.Files.CheckPathExists(fullDaemonConfigPath) {
				fmt.Println("Unable to generate Promethium daemon config as it already exists")
			} else {

				//gen and save config...

				err := config.NewPromethiumDaemonConfig(configPath, installPath, socketPath, daemonUser, daemonGroup, jailUser, jailGroup, enableHttp, uint(httpPort), httpAddr, false, 0, "", "", "", "")
				if err != nil {
					return err
				}

			}

		}

		//now load the config...

		conf, err := config.LoadPromethiumDaemonConfigAtPath(configPath)
		if err != nil {
			return err
		}

		DEST_EXEC_PATH := "/usr/bin/promethium"

		//install *this* binary on the path /usr/bin/promethium

		ex, err := os.Executable()
		if err != nil {
			return err
		}

		vutils.Files.Copy(ex, DEST_EXEC_PATH)

		os.Chmod(DEST_EXEC_PATH, 0755)

		//now we have valid config and binary in path install service...

		return config.InstallServiceUnit("/etc/systemd/system", "promethium.service", DEST_EXEC_PATH, configPath, conf.User, conf.Group, false, false)

		//return nil
	},
}
