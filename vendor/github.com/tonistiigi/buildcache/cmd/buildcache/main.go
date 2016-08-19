package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	mainCmd = &cobra.Command{
		Use:   os.Args[0],
		Short: "Manage Docker build cache",
		// SilenceUsage:  true,
		// SilenceErrors: true,
	}
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)

	mainCmd.PersistentFlags().StringP("host", "H", "unix:///var/lib/docker.sock", "Docker socket to connect to")

	mainCmd.AddCommand(
		saveCmd,
		// loadCmd,
		// pushCmd,
		// pullCmd,
	)
}

func main() {
	if err := mainCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
