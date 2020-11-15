package app

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/joesonw/distress/pkg/metrics"
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
	pOut := cmd.Flags().StringP("out", "o", "console", "stats output target")

	cmd.Args = cobra.ExactValidArgs(1)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		logger := *pLogger

		var reporter metrics.Reporter
		switch *pOut {
		case "console":
			reporter = metrics.Console()
		default:
			u, err := url.Parse(*pOut)
			if err != nil {
				logger.With(zap.Error(err)).Fatal("unable to parse --out/-o")
			}
			switch u.Scheme {
			case "influx+http", "influx+https":
				prot := strings.Split(u.Scheme, "+")[1]
				q := u.Query()
				reporter = metrics.Influx(
					influxdb2.NewClient(fmt.Sprintf("%s://%s", prot, u.Host), q.Get("token")),
					q.Get("org"),
					q.Get("bucket"),
				)
			default:
				logger.With(zap.Error(err)).Fatal(fmt.Sprintf("output \"%s\" is not supoprted", u.Scheme))
			}
		}

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
			dir := *pDirectory
			if !strings.HasPrefix(dir, "/") {
				cwd, _ := os.Getwd()
				dir = filepath.Join(cwd, dir)
			}
			fs = afero.NewBasePathFs(afero.NewOsFs(), dir)
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
		})

		if err != nil {
			logger.With(zap.Error(err)).Fatal("unable to create job")
		}

		if *pDuration > 0 {
			job.RunDuration(*pDuration)
		} else {
			amount := int64(1)
			if *pAmount > 0 {
				amount = int64(*pAmount)
			}
			job.RunAmount(amount)
		}

		job.Report(reporter)
	}

	return cmd
}
