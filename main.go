package main

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/n9e/k8s-mon/collect"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/n9e/k8s-mon/config"
	"github.com/oklog/run"
)

func main() {

	var (
		app        = kingpin.New(filepath.Base(os.Args[0]), "The k8s-mon")
		configFile = app.Flag("config.file", "k8s-mon configuration file path.").Default("k8s-mon.yml").String()
	)
	promlogConfig := promlog.Config{}

	app.Version(version.Print("k8s-mon"))
	app.HelpFlag.Short('h')
	promlogflag.AddFlags(app, &promlogConfig)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	var logger log.Logger
	logger = func(config *promlog.Config) log.Logger {
		var (
			l  log.Logger
			le level.Option
		)
		if config.Format.String() == "logfmt" {
			l = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		} else {
			l = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		}

		switch config.Level.String() {
		case "debug":
			le = level.AllowDebug()
		case "info":
			le = level.AllowInfo()
		case "warn":
			le = level.AllowWarn()
		case "error":
			le = level.AllowError()
		}
		l = level.NewFilter(l, le)
		l = log.With(l, "ts", log.TimestampFormat(
			func() time.Time { return time.Now().Local() },
			"2006-01-02T15:04:05.000Z07:00",
		), "caller", log.DefaultCaller)
		return l
	}(&promlogConfig)

	sConfig, err := config.LoadFile(*configFile, logger)
	if err != nil {
		level.Error(logger).Log("msg", "config.LoadFile Error,Exiting ...", "error", err)
		return
	}
	// cadvisor 模式沿用之前的插件运行一次
	if sConfig.CollectMode == "cadvisor_plugin" {
		level.Info(logger).Log("collect_mode", "cadvisor_plugin", "msg", "start collecting..")
		collect.CollectCadvisorM(sConfig.CadvisorC, logger)
		return
	}

	level.Info(logger).Log("collect_mode", sConfig.CollectMode, "msg", "start collecting..")
	var g run.Group

	ctxAll, cancelAll := context.WithCancel(context.Background())

	{
		// Termination handler.
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		cancel := make(chan struct{})
		g.Add(

			func() error {
				select {
				case <-term:
					level.Warn(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
					cancelAll()
					return nil
					//TODO clean work here
				case <-cancel:
					level.Warn(logger).Log("msg", "agent finally exit...")
					return nil
				}
			},
			func(err error) {
				close(cancel)

			},
		)
	}

	if sConfig.CollectMode == "kubelet_agent" {
		// collector .
		g.Add(func() error {

			dataM := collect.NewHistoryMap()

			//err := collect.CommonCollectTicker(sConfig, ctxAll, logger, collect.DoKubeletCollect, "kubelet")
			err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, dataM, collect.DoKubeletCollect, collect.FUNCNAME_KUBELET)
			if err != nil {
				level.Error(logger).Log("msg", "kubelet collect-manager stopped")
			}

			return err
		}, func(err error) {
			cancelAll()

		})
	}

	if sConfig.CollectMode == "server_side" {

		serviceIsM := collect.NewHistoryMap()
		collect.GetServerAddrByGetPod(logger, serviceIsM)

		// get pod
		g.Add(func() error {
			err := collect.GetServerAddrTicker(ctxAll, logger, sConfig.Step, serviceIsM)
			if err != nil {
				level.Error(logger).Log("msg", "get_pod stopped")
			}

			return err
		}, func(err error) {
			cancelAll()

		})
		// kube-scheduler
		g.Add(func() error {
			err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, serviceIsM, collect.DoKubeSchedulerCollect, collect.FUNCNAME_KUBESCHEDULER)
			if err != nil {
				level.Error(logger).Log("msg", "kube-scheduler  collect-manager stopped")
			}

			return err
		}, func(err error) {
			cancelAll()

		})

		// kube-controller-manager
		g.Add(func() error {
			err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, serviceIsM, collect.DoKubeControllerCollect, collect.FUNCNAME_KUBECONTROLLER)
			if err != nil {
				level.Error(logger).Log("msg", "kube-controller-manager  collect-manager stopped")
			}

			return err
		}, func(err error) {
			cancelAll()

		})
		// coredns
		g.Add(func() error {
			err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, serviceIsM, collect.DoKubeCoreDnsCollect, collect.FUNCNAME_COREDNS)
			if err != nil {
				level.Error(logger).Log("msg", "coredns  collect-manager stopped")
			}

			return err
		}, func(err error) {
			cancelAll()

		})

		// api-server
		g.Add(func() error {

			//err := collect.CommonCollectTicker(sConfig, ctxAll, logger, collect.DoApiServerCollect, "api-server")
			err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, serviceIsM, collect.DoApiServerCollect, collect.FUNCNAME_APISERVER)
			if err != nil {
				level.Error(logger).Log("msg", "api-server  collect-manager stopped")
			}

			return err
		}, func(err error) {
			cancelAll()

		})
		// kube-stats-metrics
		g.Add(func() error {
			// ksm指标多延迟启动
			time.Sleep(2)
			err := collect.CommonCollectTicker(sConfig, ctxAll, logger, collect.DoKubeStatsMetricsCollect, collect.FUNCNAME_KUBESTATSMETRICS)
			if err != nil {
				level.Error(logger).Log("msg", "kube-stats-metrics collect-manager stopped")
			}

			return err
		}, func(err error) {
			cancelAll()

		})

	}

	g.Run()

}
