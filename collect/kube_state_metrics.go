package collect

import (
	"fmt"
	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/n9e/k8s-mon/config"
	"strings"
	"time"
)

func DoKubeStatsMetricsCollect(cg *config.Config, logger log.Logger, funcName string) {
	start := time.Now()
	metrics, err := CurlTlsMetricsApi(logger, funcName, cg.KubeStatsC, cg.AppendTags, cg.Step, cg.TimeOutSeconds)

	if err != nil {
		level.Error(logger).Log("msg", "DoKubeStatsMetricsCollectCurlTlsMetricsApiResError", "err", err)
		return
	}
	if len(metrics) == 0 {
		level.Error(logger).Log("msg", "DoKubeStatsMetricsCollectEmptyMetricsResult")
		return
	}
	level.Info(logger).Log("msg", "DoKubeStatsMetricsCollectCurlTlsMetricsApiRes", "resNum", len(metrics))

	// kube_deployment_labels{deployment="test-server01-deployment", env="prod", instance="10.100.30.139:8080", job="business", label_app="test-server01",
	//label_dtree_name="bive", label_dtree_nid="200", namespace="default", region="ap-south-1", service="k8s-mon"}

	rm := map[string]struct{}{
		"kube_daemonset_labels":             {},
		"kube_deployment_labels":            {},
		"kube_endpoint_labels":              {},
		"kube_namespace_labels":             {},
		"kube_node_labels":                  {},
		"kube_persistentvolumeclaim_labels": {},
		"kube_persistentvolume_labels":      {},
		"kube_pod_labels":                   {},
		"kube_replicaset_labels":            {},
		"kube_secret_labels":                {},
		"kube_service_labels":               {},
		"kube_statefulset_labels":           {},
		"kube_storageclass_labels":          {},
	}
	//
	//rsm := map[string]struct{}{
	//	"daemonset":             {},
	//	"deployment":            {},
	//	"endpoint":              {},
	//	"namespace":             {},
	//	"node":                  {},
	//	"persistentvolume":      {},
	//	"persistentvolumeclaim": {},
	//	"pod":                   {},
	//	"replicaset":            {},
	//	"secret":                {},
	//	"service":               {},
	//	"statefulset":           {},
	//	"storageclass":          {},
	//}

	// tag 白名单

	tagWhiteM := make(map[string]struct{})
	if len(cg.KubeStatsC.TagsWhiteList) == 0 {
		tagWhiteM = map[string]struct{}{
			"namespace":             {},
			"pod":                   {},
			"pod_name":              {},
			"container":             {},
			"container_name":        {},
			"reason":                {},
			"secret":                {},
			"condition":             {},
			"node":                  {},
			"deployment":            {},
			"configmap":             {},
			"daemonset":             {},
			"endpoint":              {},
			"replicaset":            {},
			"statefulset":           {},
			"persistentvolumeclaim": {},
			"persistentvolume":      {},
			"service":               {},
			"phase":                 {},
			"type":                  {},
			"status":                {},
			"resource":              {},
			"unit":                  {},
			"host_ip":               {},
			"pod_ip":                {},
			"cluster_ip":            {},
			"external_name":         {},
			"load_balancer_ip":      {},
			"storageclass":          {},
			"kernel_version":        {},
			"os_image":              {},
			"role":                  {},
			"volume":                {},
			"access_mode":           {},
			"volumename":            {},
			"provisioner":           {},
			"reclaimPolicy":         {},
			"volumeBindingMode":     {},
			"poddisruptionbudget":   {},
		}
	} else {
		for _, i := range cg.KubeStatsC.TagsWhiteList {
			tagWhiteM[i] = struct{}{}
		}
	}
	nidLabelName := fmt.Sprintf("label_%s", cg.N9eNidLabelName)
	//insM := make(map[string]map[string]string)
	nidM := make(map[string]string)
	sepStr := "|||||"
	// 为了防止 deployment  daemonset重名问题
	//deployment="test-server01" daemonset="test-server01"
	for _, metric := range metrics {
		if _, loaded := rm[metric.Metric]; loaded {
			ss := strings.Split(metric.Metric, "_")
			resouceName := ss[1]
			resouceId := ""
			nid := ""
			for k, v := range metric.TagsMap {

				if k == nidLabelName {
					nid = v
				}
				if k == resouceName {
					resouceId = resouceName + sepStr + v

				}
			}
			if resouceId != "" && nid != "" {
				nidM[resouceId] = nid
			}

		}

	}

	// 整理label:

	var metricList []dataobj.MetricValue
	for _, metric := range metrics {
		// 去掉kube_<>_labels
		//if _, loaded := rm[metric.Metric]; loaded {
		//	continue
		//}

		for k, v := range metric.TagsMap {

			kk := k + sepStr + v

			if nid, loaded := nidM[kk]; loaded {
				metric.Nid = nid
			}
			if _, loaded := tagWhiteM[k]; !loaded {
				delete(metric.TagsMap, k)
			}
		}

		// tags string
		metric.Tags = makeAppendTags(metric.TagsMap, cg.AppendTags)
		metric.Step = cg.Step
		metricList = append(metricList, metric)

	}
	level.Info(logger).Log("msg", "DoCollectSuccessfullyReadyToPush", "funcName", funcName, "metrics_num", len(metricList), "time_took_seconds", time.Since(start).Seconds(), "metric_addr", cg.KubeStatsC.Addr)

	go PushWork(cg.PushServerAddr, cg.TimeOutSeconds, metricList, logger, funcName)
}
