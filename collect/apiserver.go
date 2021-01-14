package collect

import (
	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/n9e/k8s-mon/config"
	"time"
)

func DoApiServerCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {
	start := time.Now()
	// get dataM

	metricList := CurlMetricsByIps(cg.ApiServerC, logger, dataMap, funcName, cg.AppendTags, cg.Step, cg.TimeOutSeconds, cg.MultiServerInstanceUniqueLabel)

	if len(metricList) == 0 {
		level.Error(logger).Log("msg", "FinallyCurlTlsMetricsResEmpty", "func_name", funcName)
		return
	}

	//  设置nid
	ml := make([]dataobj.MetricValue, 0)
	for _, m := range metricList {
		m.Nid = cg.ServerSideNid
		ml = append(ml, m)

	}
	level.Info(logger).Log("msg", "DoCollectSuccessfullyReadyToPush", "funcName", funcName, "metrics_num", len(ml), "time_took_seconds", time.Since(start).Seconds())
	go PushWork(cg.PushServerAddr, cg.TimeOutSeconds, ml, logger, funcName)

}
