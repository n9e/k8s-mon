package collect

import (
	"fmt"
	"time"

	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/n9e/k8s-mon/config"
)

func DoKubeProxyOnNodeCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {

	start := time.Now()
	kubeproxyAddr, err := GetPortListenAddr(cg.KubeProxyC.Port)

	if kubeproxyAddr == "" {
		level.Warn(logger).Log("msg", "getPortListenAddrEmptykubeProxyAddr", "funcName", funcName, "err:", err, "port", cg.KubeProxyC.Port)

	} else {

		cg.KubeProxyC.Addr = fmt.Sprintf("%s://%s:%d/%s", cg.KubeProxyC.Scheme, kubeproxyAddr, cg.KubeProxyC.Port, cg.KubeProxyC.MetricsPath)
		level.Info(logger).Log("msg", "getPortListenAddrForkubeProxyAddr", "funcName", funcName, "port", cg.KubeProxyC.Port, "ipaddr", kubeproxyAddr, "kubeletPath", cg.KubeProxyC.Addr)
	}
	if cg.KubeProxyC.Addr == "" && len(cg.KubeProxyC.UserSpecifyAddrs) > 0 {
		cg.KubeProxyC.Addr = cg.KubeProxyC.UserSpecifyAddrs[0]
	}

	// 添加server_addr tag
	m := map[string]string{
		cg.MultiServerInstanceUniqueLabel: kubeproxyAddr,
		cg.MultiFuncUniqueLabel:           funcName,
	}
	for k, v := range cg.AppendTags {
		m[k] = v
	}
	metricList, err := CurlTlsMetricsApi(logger, funcName, cg.KubeProxyC, m, cg.Step, cg.TimeOutSeconds, true)
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
	level.Info(logger).Log("msg", "DoCollectSuccessfullyReadyToPush", "funcName", funcName, "metrics_num", len(metricList), "time_took_seconds", time.Since(start).Seconds(), "metric_addr", cg.KubeProxyC.Addr)

	go PushWork(cg.PushServerAddr, cg.TimeOutSeconds, ml, logger, funcName)
}

func DoKubeProxyCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {
	ConcurrencyCurlMetricsByIpsSetNid(cg.KubeProxyC, logger, dataMap, funcName, cg.AppendTags, cg.Step, cg.TimeOutSeconds, cg.MultiServerInstanceUniqueLabel, cg.MultiFuncUniqueLabel, cg.ServerSideNid, cg.PushServerAddr, true)

}
