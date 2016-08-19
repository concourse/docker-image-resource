package main

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Sirupsen/logrus"
	engineapi "github.com/docker/engine-api/client"
	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/tonistiigi/buildcache"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"
)

var saveCmd = &cobra.Command{
	Use:   "save [image]",
	Short: "Save buildcache",
	RunE:  runSave,
}

func init() {
	saveCmd.Flags().StringP("output", "o", "-", "output file")
	saveCmd.Flags().StringP("graph", "g", "", "graph directory")
}

func callOnSignal(ctx context.Context, fn func(), s ...os.Signal) {
	ctx, c := context.WithCancel(ctx)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, s...)
	go func() {
		for range sigchan {
			logrus.Debugf("received signal")
			fn()
			c()
		}
	}()
	go func() {
		<-ctx.Done()
		signal.Stop(sigchan)
	}()
}

func runSave(cmd *cobra.Command, args []string) (reterr error) {
	if len(args) == 0 {
		return errors.New("image reference missing")
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	if output == "-" && terminal.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("refusing to output to terminal, specify output file")
	}

	client, err := engineapi.NewEnvClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	callOnSignal(ctx, cancel, syscall.SIGINT)
	defer cancel()

	graphdir, err := cmd.Flags().GetString("graph")
	if err != nil {
		return err
	}

	c, err := buildcache.New(client).Get(ctx, graphdir, args[0])
	if err != nil {
		return err
	}

	if output == "-" {
		_, err := io.Copy(os.Stdout, c)
		return err
	}

	f, err := ioutil.TempFile(filepath.Dir(output), ".buildcache-")
	if err != nil {
		return err
	}
	defer func() {
		if reterr != nil {
			os.RemoveAll(f.Name())
		}
	}()
	if n, err := io.Copy(f, c); err != nil {
		return err
	} else {
		logrus.Debugf("saving: %v", humanize.Bytes(uint64(n)))
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), output)
}
