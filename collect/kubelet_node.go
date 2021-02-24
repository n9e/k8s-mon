package collect

import (
	"fmt"
	"time"

	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/n9e/k8s-mon/config"
)

func DoKubeletNodeOnNodeCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {

	start := time.Now()
	kubeletNodeAddr, err := GetPortListenAddr(cg.KubeletNodeC.Port)

	if kubeletNodeAddr == "" {
		level.Warn(logger).Log("msg", "getPortListenAddrEmptykubeletNodeAddr", "funcName", funcName, "err:", err, "port", cg.KubeletNodeC.Port)

	} else {

		cg.KubeletNodeC.Addr = fmt.Sprintf("%s://%s:%d/%s", cg.KubeletNodeC.Scheme, kubeletNodeAddr, cg.KubeletNodeC.Port, cg.KubeletNodeC.MetricsPath)
		level.Info(logger).Log("msg", "getPortListenAddrForKubeletAddr", "funcName", funcName, "port", cg.KubeletNodeC.Port, "ipaddr", kubeletNodeAddr, "kubeletPath", cg.KubeletNodeC.Addr)
	}
	if cg.KubeletNodeC.Addr == "" && len(cg.KubeletNodeC.UserSpecifyAddrs) > 0 {
		cg.KubeletNodeC.Addr = cg.KubeletNodeC.UserSpecifyAddrs[0]
	}
	// 添加server_addr tag
	m := map[string]string{
		cg.MultiServerInstanceUniqueLabel: kubeletNodeAddr,
		cg.MultiFuncUniqueLabel:           funcName,
	}
	for k, v := range cg.AppendTags {
		m[k] = v
	}

	metricList, err := CurlTlsMetricsApi(logger, funcName, cg.KubeletNodeC, m, cg.Step, cg.TimeOutSeconds, true)
	if err != nil {
		level.Error(logger).Log("msg", "CurlTlsMetricsApiResError", "err:", err, "funcName", funcName)
		return
	}

	if len(metricList) == 0 {
		level.Error(logger).Log("msg", "DoCollectEmptyMetricsResult", "funcName", funcName)
		return
	}
	ml := make([]dataobj.MetricValue, 0)
	for _, m := range metricList {

		m.Nid = cg.ServerSideNid
		ml = append(ml, m)

	}
	level.Debug(logger).Log("msg", "DoCollectSuccessfullyReadyToPush", "funcName", funcName, "metrics_num", len(metricList), "time_took_seconds", time.Since(start).Seconds(), "metric_addr", cg.KubeletNodeC.Addr)

	go PushWork(cg.PushServerAddr, cg.TimeOutSeconds, ml, logger, funcName)

}

func DoKubeletNodeCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {
	ConcurrencyCurlMetricsByIpsSetNid(cg.KubeletNodeC, logger, dataMap, funcName, cg.AppendTags, cg.Step, cg.TimeOutSeconds, cg.MultiServerInstanceUniqueLabel, cg.MultiFuncUniqueLabel, cg.ServerSideNid, cg.PushServerAddr, true)

}
