package collect

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/n9e/k8s-mon/config"
)

func DoKubeSchedulerCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {

	//ConcurrencyCurlMetricsByIpsSetNid(cg.KubeSchedulerC, logger, dataMap, funcName, cg.AppendTags, cg.Step, cg.TimeOutSeconds, cg.MultiServerInstanceUniqueLabel, cg.MultiFuncUniqueLabel, cg.ServerSideNid, cg.PushServerAddr)
	start := time.Now()
	metricUrlMap := GetServerSideAddr(cg.KubeSchedulerC, logger, dataMap, funcName)
	if len(metricUrlMap) == 0 {
		level.Error(logger).Log("msg", "GetServerSideAddrEmpty", "funcName:", funcName)
		return
	}
	var metricList []dataobj.MetricValue
	index := 0
	allNum := len(metricUrlMap)

	scheduler_e2e_scheduling_duration_seconds_bucket := "scheduler_e2e_scheduling_duration_seconds_bucket"
	scheduler_e2e_scheduling_duration_seconds_bucket_m := make(map[float64]float64)

	scheduler_pod_scheduling_duration_seconds_bucket := "scheduler_pod_scheduling_duration_seconds_bucket"
	scheduler_pod_scheduling_duration_seconds_bucket_m := make(map[float64]float64)

	scheduler_scheduling_algorithm_duration_seconds_bucket := "scheduler_scheduling_algorithm_duration_seconds_bucket"
	scheduler_scheduling_algorithm_duration_seconds_bucket_m := make(map[float64]float64)

	scheduler_pod_scheduling_attempts_bucket := "scheduler_pod_scheduling_attempts_bucket"
	scheduler_pod_scheduling_attempts_bucket_m := make(map[float64]float64)

	avg_m := make(map[string]map[string]float64)

	scheduler_e2e_scheduling_duration_seconds_sum := "scheduler_e2e_scheduling_duration_seconds_sum"
	scheduler_e2e_scheduling_duration_seconds_count := "scheduler_e2e_scheduling_duration_seconds_count"

	scheduler_pod_scheduling_duration_seconds_sum := "scheduler_pod_scheduling_duration_seconds_sum"
	scheduler_pod_scheduling_duration_seconds_count := "scheduler_pod_scheduling_duration_seconds_count"

	scheduler_scheduling_algorithm_duration_seconds_sum := "scheduler_scheduling_algorithm_duration_seconds_sum"
	scheduler_scheduling_algorithm_duration_seconds_count := "scheduler_scheduling_algorithm_duration_seconds_count"

	scheduler_pod_scheduling_attempts_sum := "scheduler_pod_scheduling_attempts_sum"
	scheduler_pod_scheduling_attempts_count := "scheduler_pod_scheduling_attempts_count"

	for uniqueHost, murl := range metricUrlMap {
		tmp := *cg.KubeSchedulerC
		c := &tmp
		c.Addr = murl
		// 添加service_addr tag
		newtagsm := map[string]string{
			cg.MultiServerInstanceUniqueLabel: uniqueHost,
			cg.MultiFuncUniqueLabel:           funcName,
		}
		for k, v := range cg.AppendTags {
			newtagsm[k] = v
		}
		metrics, err := CurlTlsMetricsApi(logger, funcName, c, newtagsm, cg.Step, cg.TimeOutSeconds)

		if err != nil {
			level.Error(logger).Log("msg", "CurlTlsMetricsResError", "func_name", funcName, "err:", err, "seq", fmt.Sprintf("%d/%d", index, allNum), "addr", c.Addr)
			continue
		}
		if len(metrics) == 0 {
			level.Error(logger).Log("msg", "CurlTlsMetricsResEmpty", "func_name", funcName, "seq", fmt.Sprintf("%d/%d", index, allNum), "addr", c.Addr)
			continue
		}

		for _, metric := range metrics {

			switch metric.Metric {
			case scheduler_e2e_scheduling_duration_seconds_bucket:

				upperBound := metric.TagsMap["le"]
				upperBoundV, _ := strconv.ParseFloat(upperBound, 64)
				scheduler_e2e_scheduling_duration_seconds_bucket_m[upperBoundV] += metric.Value
				continue
			case scheduler_pod_scheduling_duration_seconds_bucket:

				upperBound := metric.TagsMap["le"]
				upperBoundV, _ := strconv.ParseFloat(upperBound, 64)
				scheduler_pod_scheduling_duration_seconds_bucket_m[upperBoundV] += metric.Value
				continue
			case scheduler_scheduling_algorithm_duration_seconds_bucket:

				upperBound := metric.TagsMap["le"]
				upperBoundV, _ := strconv.ParseFloat(upperBound, 64)
				scheduler_scheduling_algorithm_duration_seconds_bucket_m[upperBoundV] += metric.Value
				continue
			case scheduler_pod_scheduling_attempts_bucket:

				upperBound := metric.TagsMap["le"]
				upperBoundV, _ := strconv.ParseFloat(upperBound, 64)
				scheduler_pod_scheduling_attempts_bucket_m[upperBoundV] += metric.Value
				continue

			case scheduler_e2e_scheduling_duration_seconds_sum:
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im

			case scheduler_e2e_scheduling_duration_seconds_count:
				newName := strings.Split(metric.Metric, "_count")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["count"] += metric.Value
				avg_m[newName] = im
			case scheduler_pod_scheduling_duration_seconds_sum:
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im

			case scheduler_pod_scheduling_duration_seconds_count:
				newName := strings.Split(metric.Metric, "_count")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["count"] += metric.Value
				avg_m[newName] = im
			case scheduler_scheduling_algorithm_duration_seconds_sum:
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im

			case scheduler_scheduling_algorithm_duration_seconds_count:
				newName := strings.Split(metric.Metric, "_count")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["count"] += metric.Value
				avg_m[newName] = im
			case scheduler_pod_scheduling_attempts_sum:
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im

			case scheduler_pod_scheduling_attempts_count:
				newName := strings.Split(metric.Metric, "_count")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["count"] += metric.Value
				avg_m[newName] = im
			default:
				if strings.HasSuffix(metric.Metric, "_bucket") {
					continue
				}

			}
			//	common
			if metric.CounterType == config.METRIC_TYPE_COUNTER {
				metric.Metric = metric.Metric + config.COUNTER_TO_GAUGE_METRIC_NAME_SUFFIX
			}

			metric.Nid = cg.ServerSideNid
			metricList = append(metricList, metric)

		}

	}

	newtagsm := map[string]string{
		cg.MultiFuncUniqueLabel: funcName,
	}
	for k, v := range cg.AppendTags {
		newtagsm[k] = v
	}

	// 开始算quantile
	metricList = histogramDeltaWork(dataMap, scheduler_e2e_scheduling_duration_seconds_bucket_m, newtagsm, funcName, scheduler_e2e_scheduling_duration_seconds_bucket, cg.ServerSideNid, cg.Step, metricList)
	metricList = histogramDeltaWork(dataMap, scheduler_pod_scheduling_duration_seconds_bucket_m, newtagsm, funcName, scheduler_pod_scheduling_duration_seconds_bucket, cg.ServerSideNid, cg.Step, metricList)
	metricList = histogramDeltaWork(dataMap, scheduler_scheduling_algorithm_duration_seconds_bucket_m, newtagsm, funcName, scheduler_scheduling_algorithm_duration_seconds_bucket, cg.ServerSideNid, cg.Step, metricList)
	metricList = histogramDeltaWork(dataMap, scheduler_pod_scheduling_attempts_bucket_m, newtagsm, funcName, scheduler_pod_scheduling_attempts_bucket, cg.ServerSideNid, cg.Step, metricList)

	// 开始算平均值
	for mName, avgm := range avg_m {
		mm := avgCompute(avgm, cg.ServerSideNid, mName, cg.Step, newtagsm)
		metricList = append(metricList, mm...)

	}
	level.Info(logger).Log("msg", "DoCollectSuccessfullyReadyToPush", "funcName", funcName, "metrics_num", len(metricList), "time_took_seconds", time.Since(start).Seconds())
	go PushWork(cg.PushServerAddr, cg.TimeOutSeconds, metricList, logger, funcName)

}
