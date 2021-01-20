package collect

import (
	"github.com/go-kit/kit/log"

	"github.com/n9e/k8s-mon/config"
)

func DoKubeSchedulerCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {

	ConcurrencyCurlMetricsByIpsSetNid(cg.KubeSchedulerC, logger, dataMap, funcName, cg.AppendTags, cg.Step, cg.TimeOutSeconds, cg.MultiServerInstanceUniqueLabel, cg.MultiFuncUniqueLabel, cg.ServerSideNid, cg.PushServerAddr)

}
