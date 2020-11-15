package app

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/uber-go/tally"
	"go.uber.org/zap"

	goutil "github.com/joesonw/distress/pkg/util"
)

func MakeCmdRun(
	pLogger **zap.Logger,
	pDebug *bool,
) *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
	}

	pEnvs := cmd.PersistentFlags().StringArrayP("env", "e", nil, "set lua script environment variables")
	pDuration := cmd.Flags().DurationP("duration", "t", 0, "run amount of take, takes precedence of --amount/-n")
	pAmount := cmd.Flags().IntP("amount", "n", 1, "amount of requests/runs to be made")
	pConcurrency := cmd.Flags().IntP("concurrency", "c", 1, "run concurrency")
	pFile := cmd.Flags().StringP("file", "f", "", "zip file of contents")
	pDirectory := cmd.Flags().StringP("directory", "d", "", "directory of contents")
	//pOut := cmd.Flags().StringP("out", "o", "", "metrics output target")
	pInf := cmd.Flags().Bool("inf", false, "run infinitely until stop")

	cmd.Args = cobra.ExactValidArgs(1)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		logger := *pLogger

		envs := map[string]string{}
		for _, env := range *pEnvs {
			kvs := strings.Split(env, "=")
			if len(kvs) >= 2 {
				envs[kvs[0]] = strings.Join(kvs[1:], "=")
			}
		}

		var fs afero.Fs
		var newFSPath string
		var err error

		if *pDirectory != "" {
			fs = afero.NewBasePathFs(afero.NewOsFs(), *pDirectory)
			newFSPath = *pDirectory
		} else if *pFile != "" {
			fs, err = goutil.NewAferoFsByPath(*pFile)
			if err != nil {
				logger.With(zap.Error(err)).Error("unable to open file")
			}
			newFSPath = filepath.Dir(*pFile)
		} else {
			logger.Fatal("either --file/-f or --directory/-d has to be specified")
		}

		concurrency := 1
		if *pConcurrency > 1 {
			concurrency = *pConcurrency
		}

		job, err := newJob(logger, fs, args[0], concurrency, envs, func() afero.Fs {
			return afero.NewBasePathFs(afero.NewOsFs(), newFSPath)
		}, tally.NoopScope)

		if err != nil {
			logger.With(zap.Error(err)).Fatal("unable to create job")
		}

		if *pInf {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, syscall.SIGTERM, os.Interrupt)
			job.RunInfinity(ch)
			job.Report()
		} else {
			if *pDuration > 0 {
				job.RunDuration(*pDuration)
			} else {
				amount := int64(1)
				if *pAmount > 0 {
					amount = int64(*pAmount)
				}
				job.RunAmount(amount)
			}
		}
	}

	return cmd
}
