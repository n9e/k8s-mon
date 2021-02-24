package collect

import (
	"fmt"
	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/n9e/k8s-mon/config"
	"strconv"
	"strings"
	"time"
)

func DoKubeControllerCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {

	start := time.Now()
	metricUrlMap := GetServerSideAddr(cg.KubeControllerC, logger, dataMap, funcName)
	if len(metricUrlMap) == 0 {
		level.Error(logger).Log("msg", "GetServerSideAddrEmpty", "funcName:", funcName)
		return
	}

	rest_client_request_duration_seconds_bucket := "rest_client_request_duration_seconds_bucket"
	rest_client_request_duration_seconds_bucket_m := make(map[float64]float64)

	workqueue_queue_duration_seconds_bucket := "workqueue_queue_duration_seconds_bucket"
	workqueue_queue_duration_seconds_bucket_m := make(map[float64]float64)

	workqueue_work_duration_seconds_bucket := "workqueue_work_duration_seconds_bucket"
	workqueue_work_duration_seconds_bucket_m := make(map[float64]float64)

	rest_client_request_duration_seconds_sum := "rest_client_request_duration_seconds_sum"
	rest_client_request_duration_seconds_count := "rest_client_request_duration_seconds_count"

	workqueue_queue_duration_seconds_sum := "workqueue_queue_duration_seconds_sum"
	workqueue_queue_duration_seconds_count := "workqueue_queue_duration_seconds_count"

	workqueue_work_duration_seconds_sum := "workqueue_work_duration_seconds_sum"
	workqueue_work_duration_seconds_count := "workqueue_work_duration_seconds_count"

	avg_m := make(map[string]map[string]float64)

	var metricList []dataobj.MetricValue
	index := 0
	allNum := len(metricUrlMap)
	for uniqueHost, murl := range metricUrlMap {
		tmp := *cg.KubeControllerC
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
		metrics, err := CurlTlsMetricsApi(logger, funcName, c, newtagsm, cg.Step, cg.TimeOutSeconds, false)

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
			case rest_client_request_duration_seconds_bucket:

				upperBound := metric.TagsMap["le"]
				upperBoundV, _ := strconv.ParseFloat(upperBound, 64)
				rest_client_request_duration_seconds_bucket_m[upperBoundV] += metric.Value
				continue
			case workqueue_queue_duration_seconds_bucket:

				upperBound := metric.TagsMap["le"]
				upperBoundV, _ := strconv.ParseFloat(upperBound, 64)
				workqueue_queue_duration_seconds_bucket_m[upperBoundV] += metric.Value
				continue
			case workqueue_work_duration_seconds_bucket:

				upperBound := metric.TagsMap["le"]
				upperBoundV, _ := strconv.ParseFloat(upperBound, 64)
				workqueue_work_duration_seconds_bucket_m[upperBoundV] += metric.Value
				continue
				//	共同指标
			case rest_client_request_duration_seconds_sum:
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im

			case rest_client_request_duration_seconds_count:
				newName := strings.Split(metric.Metric, "_count")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["count"] += metric.Value
				avg_m[newName] = im
			case workqueue_queue_duration_seconds_sum:
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im

			case workqueue_queue_duration_seconds_count:
				newName := strings.Split(metric.Metric, "_count")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["count"] += metric.Value
				avg_m[newName] = im

			case workqueue_work_duration_seconds_sum:
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im
			case workqueue_work_duration_seconds_count:
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
	if len(metricList) == 0 {
		level.Error(logger).Log("msg", "CurlTlsMetricsResFinallyEmptyNotPush", "func_name", funcName)
		return
	}

	newtagsm := map[string]string{
		cg.MultiFuncUniqueLabel: funcName,
	}
	for k, v := range cg.AppendTags {
		newtagsm[k] = v
	}

	// 开始算quantile
	metricList = histogramDeltaWork(dataMap, rest_client_request_duration_seconds_bucket_m, newtagsm, funcName, "controller_manager_"+rest_client_request_duration_seconds_bucket, cg.ServerSideNid, cg.Step, metricList)
	metricList = histogramDeltaWork(dataMap, workqueue_queue_duration_seconds_bucket_m, newtagsm, funcName, "controller_manager_"+workqueue_queue_duration_seconds_bucket, cg.ServerSideNid, cg.Step, metricList)
	metricList = histogramDeltaWork(dataMap, workqueue_work_duration_seconds_bucket_m, newtagsm, funcName, "controller_manager_"+workqueue_work_duration_seconds_bucket, cg.ServerSideNid, cg.Step, metricList)

	// 开始算平均值
	for mName, avgm := range avg_m {
		mm := avgCompute(avgm, cg.ServerSideNid, mName, cg.Step, newtagsm)
		metricList = append(metricList, mm...)

	}
	level.Debug(logger).Log("msg", "DoCollectSuccessfullyReadyToPush", "funcName", funcName, "metrics_num", len(metricList), "time_took_seconds", time.Since(start).Seconds())
	go PushWork(cg.PushServerAddr, cg.TimeOutSeconds, metricList, logger, funcName)

}
