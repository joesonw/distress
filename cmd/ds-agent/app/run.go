package app

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	pStats := cmd.Flags().String("stats", "", "stats server")
	pName := cmd.Flags().String("name", "", "job name")

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
				logger.Fatal("unable to parse --out/-o", zap.Error(err))
			}
			switch u.Scheme {
			case "influx+http", "influx+https":
				prot := strings.Split(u.Scheme, "+")[1]
				q := u.Query()
				influxClient := influxdb2.NewClient(fmt.Sprintf("%s://%s", prot, u.Host), q.Get("token"))
				reporter = metrics.Influx(
					influxClient.WriteAPI(q.Get("org"), q.Get("bucket")),
					time.Second,
					*pName,
				)
			default:
				logger.Fatal(fmt.Sprintf("output \"%s\" is not supoprted", u.Scheme), zap.Error(err))
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
				logger.Error("unable to open file", zap.Error(err))
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
		}, reporter)
		if err != nil {
			logger.Fatal("unable to create job", zap.Error(err))
		}

		if s := *pStats; s != "" {
			if err := startStatsServer(s, job); err != nil {
				logger.Fatal("unable to start stats server", zap.Error(err))
			}
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
		if err := reporter.Finish(); err != nil {
			logger.Error("unable to report stast", zap.Error(err))
		}
	}

	return cmd
}
