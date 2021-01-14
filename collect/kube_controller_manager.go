package collect

import (
	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/n9e/k8s-mon/config"
	"time"
)

func DoKubeControllerCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {
	start := time.Now()

	metricList := CurlMetricsByIps(cg.KubeControllerC, logger, dataMap, funcName, cg.AppendTags, cg.Step, cg.TimeOutSeconds, cg.MultiServerInstanceUniqueLabel)

	if len(metricList) == 0 {
		level.Error(logger).Log("msg", "FinallyCurlTlsMetricsResEmpty", "func_name", funcName)
		return
	}

	ml := make([]dataobj.MetricValue, 0)
	for _, m := range metricList {
		//level.Info(logger).Log("msg", "curl_result_check", "funcName", funcName, "seq", fmt.Sprintf("%d/%d", index, len(metricList)-1), m.Metric, m.Tags)

		m.Nid = cg.ServerSideNid
		ml = append(ml, m)

	}
	level.Info(logger).Log("msg", "DoCollctSuccessfullyReadyToPush", "funcName", funcName, "metrics_num", len(ml), "time_took_seconds", time.Since(start).Seconds())

	go PushWork(cg.PushServerAddr, cg.TimeOutSeconds, ml, logger, funcName)

}
