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

//func DoApiServerCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {
//
//	ConcurrencyCurlMetricsByIpsSetNid(cg.ApiServerC, logger, dataMap, funcName, cg.AppendTags, cg.Step, cg.TimeOutSeconds, cg.MultiServerInstanceUniqueLabel, cg.MultiFuncUniqueLabel, cg.ServerSideNid, cg.PushServerAddr)
//
//}

func DoApiServerCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {

	start := time.Now()
	metricUrlMap := GetServerSideAddr(cg.ApiServerC, logger, dataMap, funcName)
	if len(metricUrlMap) == 0 {
		level.Error(logger).Log("msg", "GetServerSideAddrEmpty", "funcName:", funcName)
		return
	}

	apiserver_request_duration_seconds_bucket_m := make(map[string]map[float64]float64)
	apiserver_request_duration_seconds_bucket := "apiserver_request_duration_seconds_bucket"

	apiserver_response_sizes_bucket_m := make(map[float64]float64)
	apiserver_response_sizes_bucket := "apiserver_response_sizes_bucket"

	workqueue_queue_duration_seconds_bucket_m := make(map[float64]float64)
	workqueue_queue_duration_seconds_bucket := "workqueue_queue_duration_seconds_bucket"

	workqueue_work_duration_seconds_bucket_m := make(map[float64]float64)
	workqueue_work_duration_seconds_bucket := "workqueue_work_duration_seconds_bucket"

	etcd_request_duration_seconds_bucket_m := make(map[float64]float64)
	etcd_request_duration_seconds_bucket := "etcd_request_duration_seconds_bucket"

	apiserver_request_total := "apiserver_request_total"
	apiserver_request_total_m := make(map[string]float64)

	avg_m := make(map[string]map[string]float64)
	apiserver_request_duration_seconds_sum := "apiserver_request_duration_seconds_sum"
	apiserver_request_duration_seconds_count := "apiserver_request_duration_seconds_count"

	authentication_duration_seconds_sum := "authentication_duration_seconds_sum"
	authentication_duration_seconds_count := "authentication_duration_seconds_count"

	apiserver_response_sizes_sum := "apiserver_response_sizes_sum"
	apiserver_response_sizes_count := "apiserver_response_sizes_count"

	workqueue_queue_duration_seconds_sum := "workqueue_queue_duration_seconds_sum"
	workqueue_queue_duration_seconds_count := "workqueue_queue_duration_seconds_count"

	workqueue_work_duration_seconds_sum := "workqueue_work_duration_seconds_sum"
	workqueue_work_duration_seconds_count := "workqueue_work_duration_seconds_count"

	etcd_request_duration_seconds_sum := "etcd_request_duration_seconds_sum"
	etcd_request_duration_seconds_count := "etcd_request_duration_seconds_count"

	var metricList []dataobj.MetricValue
	index := 0
	allNum := len(metricUrlMap)
	for uniqueHost, murl := range metricUrlMap {
		tmp := *cg.ApiServerC
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
			case apiserver_request_duration_seconds_bucket:
				verb := metric.TagsMap["verb"]

				if verb == "" || verb == "CONNECT" || verb == "WATCH" {
					continue
				}

				upperBound := metric.TagsMap["le"]
				upperBoundV, _ := strconv.ParseFloat(upperBound, 64)

				allm, loaded := apiserver_request_duration_seconds_bucket_m["all"]
				if !loaded {
					allm = make(map[float64]float64)
				}
				allm[upperBoundV] += metric.Value
				apiserver_request_duration_seconds_bucket_m["all"] = allm
				verbm, loaded := apiserver_request_duration_seconds_bucket_m[verb]
				if !loaded {
					verbm = make(map[float64]float64)
				}
				verbm[upperBoundV] += metric.Value
				apiserver_request_duration_seconds_bucket_m[verb] = verbm

				continue
			case apiserver_response_sizes_bucket:
				verb := metric.TagsMap["verb"]

				if verb == "" || verb == "CONNECT" || verb == "WATCH" {
					continue
				}

				upperBound := metric.TagsMap["le"]
				upperBoundV, _ := strconv.ParseFloat(upperBound, 64)
				apiserver_response_sizes_bucket_m[upperBoundV] += metric.Value
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
			case etcd_request_duration_seconds_bucket:

				upperBound := metric.TagsMap["le"]
				upperBoundV, _ := strconv.ParseFloat(upperBound, 64)
				etcd_request_duration_seconds_bucket_m[upperBoundV] += metric.Value
				continue
			case apiserver_request_total:
				code := metric.TagsMap["code"]
				apiserver_request_total_m[code] += metric.Value
			case apiserver_request_duration_seconds_sum:
				verb := metric.TagsMap["verb"]

				if verb == "" || verb == "CONNECT" || verb == "WATCH" {
					continue
				}
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im

			case apiserver_request_duration_seconds_count:
				verb := metric.TagsMap["verb"]

				if verb == "" || verb == "CONNECT" || verb == "WATCH" {
					continue
				}
				newName := strings.Split(metric.Metric, "_count")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["count"] += metric.Value
				avg_m[newName] = im
			case apiserver_response_sizes_sum:
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im

			case apiserver_response_sizes_count:
				newName := strings.Split(metric.Metric, "_count")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["count"] += metric.Value
				avg_m[newName] = im
			case authentication_duration_seconds_sum:
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im

			case authentication_duration_seconds_count:
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
			case etcd_request_duration_seconds_sum:
				newName := strings.Split(metric.Metric, "_sum")[0]
				im, loaded := avg_m[newName]
				if !loaded {
					im = make(map[string]float64)
				}
				im["sum"] += metric.Value
				avg_m[newName] = im
			case etcd_request_duration_seconds_count:
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

	// 开始算quantile
	newtagsm := map[string]string{
		cg.MultiFuncUniqueLabel: funcName,
	}
	for k, v := range cg.AppendTags {
		newtagsm[k] = v
	}
	// apiserver_request_duration_seconds_all_bucket
	for label, m := range apiserver_request_duration_seconds_bucket_m {

		var newMetricName string
		if label == "all" {
			newMetricName = "apiserver_request_duration_seconds_all_bucket"
		} else {
			newMetricName = "apiserver_request_duration_seconds_verb_" + strings.ToLower(label) + "_bucket"
		}
		metricList = histogramDeltaWork(dataMap, m, newtagsm, funcName, newMetricName, cg.ServerSideNid, cg.Step, metricList)
	}

	metricList = histogramDeltaWork(dataMap, apiserver_response_sizes_bucket_m, newtagsm, funcName, apiserver_response_sizes_bucket, cg.ServerSideNid, cg.Step, metricList)

	metricList = histogramDeltaWork(dataMap, workqueue_queue_duration_seconds_bucket_m, newtagsm, funcName, workqueue_queue_duration_seconds_bucket, cg.ServerSideNid, cg.Step, metricList)

	metricList = histogramDeltaWork(dataMap, workqueue_work_duration_seconds_bucket_m, newtagsm, funcName, workqueue_work_duration_seconds_bucket, cg.ServerSideNid, cg.Step, metricList)

	metricList = histogramDeltaWork(dataMap, etcd_request_duration_seconds_bucket_m, newtagsm, funcName, etcd_request_duration_seconds_bucket, cg.ServerSideNid, cg.Step, metricList)

	// 开始算rate
	mm := successfulRate(apiserver_request_total_m, cg.ServerSideNid, "apiserver_request_successful_rate", cg.Step, newtagsm)
	metricList = append(metricList, mm...)

	// 开始算平均值
	for mName, avgm := range avg_m {
		mm = avgCompute(avgm, cg.ServerSideNid, mName, cg.Step, newtagsm)
		metricList = append(metricList, mm...)

	}

	level.Info(logger).Log("msg", "DoCollectSuccessfullyReadyToPush", "funcName", funcName, "metrics_num", len(metricList), "time_took_seconds", time.Since(start).Seconds())
	go PushWork(cg.PushServerAddr, cg.TimeOutSeconds, metricList, logger, funcName)

}
