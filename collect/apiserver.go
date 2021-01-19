package collect

import (
	"github.com/go-kit/kit/log"
	"github.com/n9e/k8s-mon/config"
)

func DoApiServerCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {

	ConcurrencyCurlMetricsByIpsSetNid(cg.ApiServerC, logger, dataMap, funcName, cg.AppendTags, cg.Step, cg.TimeOutSeconds, cg.MultiServerInstanceUniqueLabel, cg.MultiFuncUniqueLabel, cg.ServerSideNid, cg.PushServerAddr)

}
