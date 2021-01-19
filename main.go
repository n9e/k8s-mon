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
	if sConfig.CollectMode == config.COLLECT_MODE_CADVISOR_PLUGIN {
		level.Info(logger).Log("collect_mode", config.COLLECT_MODE_CADVISOR_PLUGIN, "msg", "start collecting..")
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

	if sConfig.CollectMode == config.COLLECT_MODE_KUBELET_AGENT {
		// kubelet_agent .
		dataM := collect.NewHistoryMap()
		g.Add(func() error {

			err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, dataM, collect.DoKubeletCollect, config.FUNCNAME_KUBELET)
			if err != nil {
				level.Error(logger).Log("msg", "kubelet-collect-manager stopped")
			}

			return err
		}, func(err error) {
			cancelAll()

		})

	}

	if sConfig.CollectMode == config.COLLECT_MODE_SERVER_SIDE {

		serviceIsM := collect.NewHistoryMap()
		collect.GetServerAddrAll(logger, serviceIsM)

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
		if sConfig.KubeSchedulerC != nil {
			g.Add(func() error {
				err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, serviceIsM, collect.DoKubeSchedulerCollect, config.FUNCNAME_KUBESCHEDULER)
				if err != nil {
					level.Error(logger).Log("msg", "kube-scheduler  collect-manager stopped")
				}

				return err
			}, func(err error) {
				cancelAll()

			})
		}

		// kube-controller-manager
		if sConfig.KubeControllerC != nil {
			g.Add(func() error {
				err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, serviceIsM, collect.DoKubeControllerCollect, config.FUNCNAME_KUBECONTROLLER)
				if err != nil {
					level.Error(logger).Log("msg", "kube-controller-manager  collect-manager stopped")
				}

				return err
			}, func(err error) {
				cancelAll()

			})
		}

		// coredns
		if sConfig.CoreDnsC != nil {
			g.Add(func() error {
				err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, serviceIsM, collect.DoKubeCoreDnsCollect, config.FUNCNAME_COREDNS)
				if err != nil {
					level.Error(logger).Log("msg", "coredns  collect-manager stopped")
				}

				return err
			}, func(err error) {
				cancelAll()

			})
		}

		if sConfig.ApiServerC != nil {
			// api-server
			g.Add(func() error {

				//err := collect.CommonCollectTicker(sConfig, ctxAll, logger, collect.DoApiServerCollect, "api-server")
				err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, serviceIsM, collect.DoApiServerCollect, config.FUNCNAME_APISERVER)
				if err != nil {
					level.Error(logger).Log("msg", "api-server  collect-manager stopped")
				}

				return err
			}, func(err error) {
				cancelAll()

			})
		}

		if sConfig.KubeletNodeC != nil {
			// kubelet-node
			g.Add(func() error {

				err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, serviceIsM, collect.DoKubeletNodeCollect, config.FUNCNAME_KUBELET_NODE)
				if err != nil {
					level.Error(logger).Log("msg", "kubelet-node  collect-manager stopped")
				}

				return err
			}, func(err error) {
				cancelAll()

			})
		}

		if sConfig.KubeProxyC != nil {
			// kube-proxy
			g.Add(func() error {

				err := collect.CommonCollectTickerForWithDataM(sConfig, ctxAll, logger, serviceIsM, collect.DoKubeProxyCollect, config.FUNCNAME_KUBEPROXY)
				if err != nil {
					level.Error(logger).Log("msg", "kube-proxy  collect-manager stopped")
				}

				return err
			}, func(err error) {
				cancelAll()

			})
		}

		if sConfig.KubeStatsC != nil {
			// kube-stats-metrics
			g.Add(func() error {
				// ksm指标多延迟启动
				time.Sleep(2)
				err := collect.CommonCollectTicker(sConfig, ctxAll, logger, collect.DoKubeStatsMetricsCollect, config.FUNCNAME_KUBESTATSMETRICS)
				if err != nil {
					level.Error(logger).Log("msg", "kube-stats-metrics collect-manager stopped")
				}

				return err
			}, func(err error) {
				cancelAll()

			})
		}

	}

	g.Run()

}
