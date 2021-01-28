package collect

import (
	"fmt"
	"strings"
	"time"

	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/n9e/k8s-mon/config"
)

func DoKubeStatsMetricsCollect(cg *config.Config, logger log.Logger, funcName string) {
	start := time.Now()
	if len(cg.KubeStatsC.UserSpecifyAddrs) == 0 {
		level.Error(logger).Log("msg", "DoKubeStatsMetricsEmptyUserSpecifyAddrs")
		return
	}
	cg.KubeStatsC.Addr = cg.KubeStatsC.UserSpecifyAddrs[0]
	metrics, err := CurlTlsMetricsApi(logger, funcName, cg.KubeStatsC, cg.AppendTags, cg.Step, cg.TimeOutSeconds, true)

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
	kube_pod_container_resource_requests_cpu_cores_m := make(map[string]float64)
	kube_pod_container_resource_limits_cpu_cores_m := make(map[string]float64)
	kube_node_status_allocatable_cpu_cores_m := make(map[string]float64)

	kube_pod_container_resource_requests_memory_bytes_m := make(map[string]float64)
	kube_pod_container_resource_limits_memory_bytes_m := make(map[string]float64)
	kube_node_status_capacity_memory_bytes_m := make(map[string]float64)

	kube_pod_info_m := make(map[string]float64)
	kube_node_status_capacity_pods_m := make(map[string]float64)

	kube_pod_container_resource_requests_cpu_cores := "kube_pod_container_resource_requests_cpu_cores"
	kube_pod_container_resource_limits_cpu_cores := "kube_pod_container_resource_limits_cpu_cores"
	kube_node_status_allocatable_cpu_cores := "kube_node_status_allocatable_cpu_cores"

	kube_pod_container_resource_requests_memory_bytes := "kube_pod_container_resource_requests_memory_bytes"
	kube_pod_container_resource_limits_memory_bytes := "kube_pod_container_resource_limits_memory_bytes"
	kube_node_status_capacity_memory_bytes := "kube_node_status_capacity_memory_bytes"

	kube_pod_info := "kube_pod_info"
	kube_node_status_capacity_pods := "kube_node_status_capacity_pods"

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
			if metric.Nid == "" {
				metric.Nid = cg.ServerSideNid
			}

			if _, loaded := tagWhiteM[k]; !loaded {
				delete(metric.TagsMap, k)
			}
		}
		labelNode := metric.TagsMap["node"]
		switch metric.Metric {

		// cpu
		case kube_pod_container_resource_requests_cpu_cores:
			kube_pod_container_resource_requests_cpu_cores_m[labelNode] += metric.Value
		case kube_pod_container_resource_limits_cpu_cores:
			kube_pod_container_resource_limits_cpu_cores_m[labelNode] += metric.Value
		case kube_node_status_allocatable_cpu_cores:
			kube_node_status_allocatable_cpu_cores_m[labelNode] += metric.Value

		//	mem
		case kube_pod_container_resource_requests_memory_bytes:
			kube_pod_container_resource_requests_memory_bytes_m[labelNode] += metric.Value
		case kube_pod_container_resource_limits_memory_bytes:
			kube_pod_container_resource_limits_memory_bytes_m[labelNode] += metric.Value
		case kube_node_status_capacity_memory_bytes:
			kube_node_status_capacity_memory_bytes_m[labelNode] += metric.Value
		// pod num
		case kube_pod_info:
			kube_pod_info_m[labelNode] += metric.Value
		case kube_node_status_capacity_pods:
			kube_node_status_capacity_pods_m[labelNode] += metric.Value
		}

		if metric.CounterType == config.METRIC_TYPE_COUNTER {
			metric.Metric = metric.Metric + config.COUNTER_TO_GAUGE_METRIC_NAME_SUFFIX
		}
		// tags string

		metric.Tags = makeAppendTags(metric.TagsMap, cg.AppendTags)
		metric.Step = cg.Step
		metricList = append(metricList, metric)

	}

	newtagsm := map[string]string{
		cg.MultiFuncUniqueLabel: funcName,
	}
	// 计算百分比
	// cpu
	metricList = PercentComputeForKsm(kube_pod_container_resource_requests_cpu_cores_m, kube_node_status_allocatable_cpu_cores_m, cg.ServerSideNid, "kube_node_pod_container_cpu_requests", "node", cg.Step, newtagsm, metricList)
	metricList = PercentComputeForKsm(kube_pod_container_resource_limits_cpu_cores_m, kube_node_status_allocatable_cpu_cores_m, cg.ServerSideNid, "kube_node_pod_container_cpu_limits", "node", cg.Step, newtagsm, metricList)
	// mem
	metricList = PercentComputeForKsm(kube_pod_container_resource_requests_memory_bytes_m, kube_node_status_capacity_memory_bytes_m, cg.ServerSideNid, "kube_node_pod_container_memory_requests", "node", cg.Step, newtagsm, metricList)
	metricList = PercentComputeForKsm(kube_pod_container_resource_limits_memory_bytes_m, kube_node_status_capacity_memory_bytes_m, cg.ServerSideNid, "kube_node_pod_container_memory_limits", "node", cg.Step, newtagsm, metricList)
	// pod
	metricList = PercentComputeForKsm(kube_pod_info_m, kube_node_status_capacity_pods_m, cg.ServerSideNid, "kube_node_pod_num", "node", cg.Step, newtagsm, metricList)

	level.Info(logger).Log("msg", "DoCollectSuccessfullyReadyToPush", "funcName", funcName, "metrics_num", len(metricList), "time_took_seconds", time.Since(start).Seconds(), "metric_addr", cg.KubeStatsC.Addr)

	go PushWork(cg.PushServerAddr, cg.TimeOutSeconds, metricList, logger, funcName)
}
