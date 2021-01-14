package collect

import (
	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/n9e/k8s-mon/config"
	"strings"
	"time"
)

func DoKubeletCollect(cg *config.Config, logger log.Logger, dataMap *HistoryMap, funcName string) {
	// 通过kubelet prometheus 接口拿到数据后做ETL
	// 根据docker inspect 接口拿到所有容器的数据，根据podName一致找到pause 容器的label 给对应的pod数据
	start := time.Now()
	metrics, err := CurlTlsMetricsApi(logger, funcName, cg.KubeletC, cg.AppendTags, cg.Step, cg.TimeOutSeconds)
	if err != nil {
		level.Error(logger).Log("msg", "CurlTlsMetricsApiResError", "err:", err)
		return
	}

	if len(metrics) == 0 {
		level.Error(logger).Log("msg", "DoKubeletCollectEmptyMetricsResult")
		return
	}

	insM, err := getLabelMapByDockerSdk(cg.N9eNidLabelName)
	if err != nil {
		level.Error(logger).Log("msg", "DoKubeletCollect getInspectAll error", "err", err)
		return
	}

	// tag 白名单

	tagWhiteM := make(map[string]struct{})
	if len(cg.KubeletC.TagsWhiteList) == 0 {
		tagWhiteM = map[string]struct{}{
			"pod":            {},
			"pod_name":       {},
			"container":      {},
			"container_name": {},
			"namespace":      {},
			"device":         {},
			"state":          {},
			"interface":      {},
			"cpu":            {},
		}
	} else {
		for _, i := range cg.KubeletC.TagsWhiteList {
			tagWhiteM[i] = struct{}{}
		}
	}

	var metricList []dataobj.MetricValue
	// sum (rate (container_cpu_usage_seconds_total{container="prometheus"}[1m])) by( container) /( sum (container_spec_cpu_quota{container="prometheus"}) by(container) /100000) * 100

	// 遍历中的结果map，用来做两个metrics 之间的关联关系: 如 container_cpu_usage_seconds_total/container_spec_cpu_quota
	// container_spec_cpu_quota 如果没设置request 或者limit则没有

	// container_cpu_usage_seconds_total
	dataMapContainerCpuUsage := make(map[string]float64)
	metricMapContainerCpuUsage := make(map[string]dataobj.MetricValue)

	// container_cpu_user_seconds_total
	dataMapContainerCpuUser := make(map[string]float64)
	metricMapContainerCpuUser := make(map[string]dataobj.MetricValue)

	// container_cpu_sys_seconds_total
	dataMapContainerCpuSys := make(map[string]float64)
	metricMapContainerCpuSys := make(map[string]dataobj.MetricValue)

	// container_cpu_cfs_periods_total
	dataMapContainerCfsPeriods := make(map[string]float64)
	metricMapContainerCfsPeriods := make(map[string]dataobj.MetricValue)

	// container_cpu_cfs_throttled_periods_total
	dataMapContainerCfsThrottledPeriods := make(map[string]float64)
	metricMapContainerCfsThrottledPeriods := make(map[string]dataobj.MetricValue)

	dataMapContainerSpecCpuQuota := make(map[string]float64)

	// container_spec_memory_limit_bytes
	dataMapContainerMemLimit := make(map[string]float64)

	// container_memory_working_set_bytes
	dataMapContainerMemWorkingSet := make(map[string]float64)
	metricMapContainerMemWorkingSet := make(map[string]dataobj.MetricValue)

	// container_memory_usage_bytes
	dataMapContainerMemUsage := make(map[string]float64)
	metricMapContainerMemUsage := make(map[string]dataobj.MetricValue)

	// container_fs_usage_bytes
	dataMapContainerFsUsage := make(map[string]float64)
	metricMapContainerFsUsage := make(map[string]dataobj.MetricValue)

	// container_fs_limit_bytes
	dataMapContainerFsLimite := make(map[string]float64)

	for _, metric := range metrics {
		// TODO 指标中存在id=/system.slicexxx 和 id=/kubepods-besteffort的两种值,貌似后者与 qos有关，可以忽略
		//container_memory_working_set_bytes{container="test-server01",id="/kubepods-besteffort-poda3896315_7194_43b3_a7e3_e393a7ac9fb5.slice/596251fa9a561ab703532ef65963aba62816b086c28a6edaaef4f425d33f98de",image="sha256:3b66c5fd9fdfff983dc1a9aedc4d1a466b33c53ca4e779df17d334667a3e0202",name="k8s_test-server01_test-server01-deployment-7b77c479c-tznsb_default_a3896315-7194-43b3-a7e3-e393a7ac9fb5_0",namespace="default",pod="test-server01-deployment-7b77c479c-tznsb"} 0 1607933517735
		//container_memory_working_set_bytes{container="test-server01",id="/system.slice/containerd.service/kubepods-besteffort-poda3896315_7194_43b3_a7e3_e393a7ac9fb5.slice/596251fa9a561ab703532ef65963aba62816b086c28a6edaaef4f425d33f98de",image="sha256:3b66c5fd9fdfff983dc1a9aedc4d1a466b33c53ca4e779df17d334667a3e0202",name="k8s_test-server01_test-server01-deployment-7b77c479c-tznsb_default_a3896315-7194-43b3-a7e3-e393a7ac9fb5_0",namespace="default",pod="test-server01-deployment-7b77c479c-tznsb"} 3.403776e+06 1607933507684
		//labelId, loaded := metric.TagsMap["id"]
		//
		//if loaded {
		//	if !strings.HasPrefix(labelId, "/system.slice") {
		//		continue
		//	}
		//
		//}

		// container_network_transmit_packets_total{container="POD",id="/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-podfb98a1a6_f862_4383_a723_6d51bf9dc5ae.slice/docker-311295f4f5ae13ac3569ff0d8a144bbf072fdd1b98fa95b47f9a67f80eab9bcc.scope",image="registry.aliyuncs.com/k8sxio/pause:3.2",interface="tunl0",name="k8s_POD_test-server01-deployment-56869b48f7-8splz_default_fb98a1a6-f862-4383-a723-6d51bf9dc5ae_0",namespace="default",pod="test-server01-deployment-56869b48f7-8splz"} 0 1609756906487
		//if metric.TagsMap["interface"] == "tunl0" {
		//	continue
		//}

		podName := metric.TagsMap["pod"]
		if podName == "" {
			// 适配低版本k8s pod_name和pod问题
			podName = metric.TagsMap["pod_name"]

		}
		if podName == "" {
			continue
		}

		// 不采集 container="" 和 container="POD"
		containerName := metric.TagsMap["container"]
		if containerName == "" {
			// 适配低版本k8s container_name和container问题
			containerName = metric.TagsMap["container_name"]

		}
		if containerName == "" || (containerName == "POD" && !strings.HasPrefix(metric.Metric, "container_network_")) {
			continue
		}
		// 赋值nid
		labelNid := ""
		if labelM, loaded := insM[podName]; loaded {
			labelNid = labelM[cg.N9eNidLabelName]
			for k, v := range labelM {
				if k == "" || v == "" {
					continue
				}
				metric.TagsMap[k] = v
			}
		}

		// 整理label:

		for k, _ := range metric.TagsMap {

			if _, loaded := tagWhiteM[k]; !loaded {
				delete(metric.TagsMap, k)
			}
		}

		// tags string
		metric.Tags = makeAppendTags(metric.TagsMap, cg.AppendTags)

		// 决定nid还是endpoint ,nid可能为空，交给服务端处理
		if cg.KubeletC.IdentifyMode == "nid" {
			metric.Nid = labelNid
		} else if cg.KubeletC.IdentifyMode == "endpoint" {
			metric.Endpoint = podName
		}
		// TODO 防止配置了nid 模式，服务组件pod 没有 nid
		if metric.Nid == "" {
			metric.Endpoint = podName
		}
		metric.Step = cg.Step

		// agent侧聚合
		// cpu
		thisCounterStats := CounterStats{Value: metric.Value, Ts: metric.Timestamp}
		// 容器unique-key 应该为podName+containerName
		sepStr := "||"
		cUniqueKey := podName + sepStr + containerName

		switch metric.Metric {
		case "container_cpu_usage_seconds_total":
			mapKey := "container_cpu_usage_seconds_total" + sepStr + cUniqueKey
			s := NewCommonCounterHis()

			s.UpdateCounterStat(thisCounterStats)

			obj, loaded := dataMap.Map.LoadOrStore(mapKey, s)
			if !loaded {
				// 说明第一次
				//level.Info(logger).Log("msg", "MapDataNotFound", "metric_name", "container_cpu_usage_seconds_total", "mapKey", mapKey)
				continue
			} else {
				//level.Info(logger).Log("msg", "MapDataGet", "metric_name", "container_cpu_usage_seconds_total", "mapKey", mapKey)
				dataHis := obj.(*CommonCounterHis)
				dataHis.UpdateCounterStat(thisCounterStats)
				dataRate := dataHis.DeltaCounter()
				dataMapContainerCpuUsage[cUniqueKey] = dataRate

				metric.Metric = "cpu.util"

				metricMapContainerCpuUsage[cUniqueKey] = metric
				dataMap.Map.Store(mapKey, dataHis)
				newM := metric
				newM.Metric = "cpu.cores.occupy"
				metricList = append(metricList, newM)
			}

		case "container_cpu_user_seconds_total":
			mapKey := "container_cpu_user_seconds_total" + sepStr + cUniqueKey
			s := NewCommonCounterHis()

			s.UpdateCounterStat(thisCounterStats)

			obj, loaded := dataMap.Map.LoadOrStore(mapKey, s)
			if !loaded {
				continue
			} else {
				dataHis := obj.(*CommonCounterHis)
				dataHis.UpdateCounterStat(thisCounterStats)
				dataRate := dataHis.DeltaCounter()
				dataMapContainerCpuUser[cUniqueKey] = dataRate
				metric.Metric = "cpu.user"

				metricMapContainerCpuUser[cUniqueKey] = metric
				dataMap.Map.Store(mapKey, dataHis)
				// 这里continue的目的是，metirc已经被改名为cpu.user做计算了
				// 而且原始点也没有必要上报了，不需要下面在append 到list中
				continue
			}
		case "container_cpu_system_seconds_total":
			mapKey := "container_cpu_system_seconds_total" + sepStr + cUniqueKey
			s := NewCommonCounterHis()

			s.UpdateCounterStat(thisCounterStats)

			obj, loaded := dataMap.Map.LoadOrStore(mapKey, s)
			if !loaded {
				// 说明第一次
				//level.Info(logger).Log("msg", "MapDataNotFound", "metric_name", "container_cpu_system_seconds_total", "mapKey", mapKey)
				continue
			} else {
				//level.Info(logger).Log("msg", "MapDataGet", "metric_name", "container_cpu_system_seconds_total", "mapKey", mapKey)
				dataHis := obj.(*CommonCounterHis)
				dataHis.UpdateCounterStat(thisCounterStats)
				dataRate := dataHis.DeltaCounter()
				dataMapContainerCpuSys[cUniqueKey] = dataRate
				metric.Metric = "cpu.sys"
				metricMapContainerCpuSys[cUniqueKey] = metric
				dataMap.Map.Store(mapKey, dataHis)
				continue
			}

		case "container_spec_cpu_quota":
			dataMapContainerSpecCpuQuota[cUniqueKey] = metric.Value
			metric.Metric = "cpu.spec.quota"

		case "container_cpu_cfs_periods_total":
			mapKey := "container_cpu_cfs_periods_total" + sepStr + cUniqueKey
			s := NewCommonCounterHis()

			s.UpdateCounterStat(thisCounterStats)

			obj, loaded := dataMap.Map.LoadOrStore(mapKey, s)
			if !loaded {
				continue
			} else {
				//level.Info(logger).Log("msg", "MapDataGet", "metric_name", "container_cpu_system_seconds_total", "mapKey", mapKey)
				dataHis := obj.(*CommonCounterHis)
				dataHis.UpdateCounterStat(thisCounterStats)
				dataRate := dataHis.DeltaCounter()
				dataMapContainerCfsPeriods[cUniqueKey] = dataRate

				metricMapContainerCfsPeriods[cUniqueKey] = metric
				dataMap.Map.Store(mapKey, dataHis)
				// rename
				metric.Metric = "cpu.periods"

			}

		case "container_cpu_cfs_throttled_periods_total":
			mapKey := "container_cpu_cfs_throttled_periods_total" + sepStr + cUniqueKey
			s := NewCommonCounterHis()

			s.UpdateCounterStat(thisCounterStats)

			obj, loaded := dataMap.Map.LoadOrStore(mapKey, s)
			if !loaded {
				// 说明第一次
				//level.Info(logger).Log("msg", "MapDataNotFound", "metric_name", "container_cpu_system_seconds_total", "mapKey", mapKey)
				continue
			} else {
				//level.Info(logger).Log("msg", "MapDataGet", "metric_name", "container_cpu_system_seconds_total", "mapKey", mapKey)
				dataHis := obj.(*CommonCounterHis)
				dataHis.UpdateCounterStat(thisCounterStats)
				dataRate := dataHis.DeltaCounter()

				dataMapContainerCfsThrottledPeriods[cUniqueKey] = dataRate
				metricMapContainerCfsThrottledPeriods[cUniqueKey] = metric
				dataMap.Map.Store(mapKey, dataHis)
				// rename
				metric.Metric = "cpu.throttled.periods"
			}
		case "container_cpu_cfs_throttled_seconds_total":
			metric.Metric = "cpu.throttled.time"

		// memory
		case "container_spec_memory_limit_bytes":
			dataMapContainerMemLimit[cUniqueKey] = metric.Value
			metric.Metric = "mem.bytes.total"

		case "container_memory_working_set_bytes":
			dataMapContainerMemWorkingSet[cUniqueKey] = metric.Value
			metricMapContainerMemWorkingSet[cUniqueKey] = metric

			metric.Metric = "mem.bytes.workingset"

		case "container_memory_usage_bytes":
			dataMapContainerMemUsage[cUniqueKey] = metric.Value
			metricMapContainerMemUsage[cUniqueKey] = metric
			metric.Metric = "mem.bytes.used"

		case "container_memory_cache":
			metric.Metric = "mem.bytes.cached"
		case "container_memory_rss":
			metric.Metric = "mem.bytes.rss"
		case "container_memory_swap":
			metric.Metric = "mem.bytes.swap"

		//	filesystem
		case "container_fs_limit_bytes":
			metric.Metric = "disk.bytes.total"
			dataMapContainerFsLimite[cUniqueKey] = metric.Value

		case "container_fs_usage_bytes":
			dataMapContainerFsUsage[cUniqueKey] = metric.Value
			metricMapContainerFsUsage[cUniqueKey] = metric
			metric.Metric = "disk.bytes.used"

		case "container_fs_reads_bytes_total":
			metric.Metric = "disk.io.read.bytes"
		case "container_fs_writes_bytes_total":
			metric.Metric = "disk.io.write.bytes"

		// network
		// bytes
		case "container_network_receive_bytes_total":
			metric.Metric = "net.in.bytes"
		case "container_network_transmit_bytes_total":
			metric.Metric = "net.out.bytes"
		// packet
		case "container_network_receive_packets_total":
			metric.Metric = "net.in.pps"
		case "container_network_transmit_packets_total":
			metric.Metric = "net.out.pps"
		// error
		case "container_network_receive_errors_total":
			metric.Metric = "net.in.errs"
		case "container_network_transmit_errors_total":
			metric.Metric = "net.out.errs"
		// drop
		case "container_network_receive_packets_dropped_total":
			metric.Metric = "net.in.dropped"
		case "container_network_transmit_packets_dropped_total":
			metric.Metric = "net.out.dropped"

		// system
		case "container_processes":
			metric.Metric = "sys.ps.process.count"
		case "container_threads":
			metric.Metric = "sys.ps.thread.count"
		case "container_file_descriptors":
			metric.Metric = "sys.fd.count.used"
		case "container_ulimits_soft":
			metric.Metric = "sys.fd.soft.ulimits"
		case "container_sockets":
			metric.Metric = "sys.socket.count.used"
		case "container_tasks_state":
			metric.Metric = "sys.task.state"
		}

		metricList = append(metricList, metric)

	}

	// agent侧计算
	// cpu.util= container_cpu_usage_seconds_total/container_spec_cpu_quota
	metricList = CpuComputeFunc(dataMapContainerCpuUsage, metricMapContainerCpuUsage, dataMapContainerSpecCpuQuota, metricList)
	// cpu.user= container_cpu_user_seconds_total/container_spec_cpu_quota
	metricList = CpuComputeFunc(dataMapContainerCpuUser, metricMapContainerCpuUser, dataMapContainerSpecCpuQuota, metricList)
	// cpu.sys= container_cpu_user_seconds_total/container_spec_cpu_quota
	metricList = CpuComputeFunc(dataMapContainerCpuSys, metricMapContainerCpuSys, dataMapContainerSpecCpuQuota, metricList)

	// cpu.throttled.util = delta(container_cpu_cfs_throttled_periods_total)/delta(container_cpu_cfs_periods_total)
	metricList = CommonComputeFunc(dataMapContainerCfsThrottledPeriods, dataMapContainerCfsPeriods, metricMapContainerCfsPeriods, metricList, "cpu.throttled.util")

	// mem.bytes.used.percent = container_memory_usage_bytes/container_spec_memory_limit_bytes
	metricList = CommonComputeFunc(dataMapContainerMemUsage, dataMapContainerMemLimit, metricMapContainerMemUsage, metricList, "mem.bytes.used.percent")

	// mem.bytes.workingset.percent = container_memory_working_set_bytes/container_spec_memory_limit_bytes
	metricList = CommonComputeFunc(dataMapContainerMemWorkingSet, dataMapContainerMemLimit, metricMapContainerMemWorkingSet, metricList, "mem.bytes.workingset.percent")

	// disk.bytes.used.percent = container_fs_usage_bytes / container_fs_limit_bytes
	metricList = CommonComputeFunc(dataMapContainerFsUsage, dataMapContainerFsLimite, metricMapContainerFsUsage, metricList, "disk.bytes.used.percent")

	level.Info(logger).Log("msg", "DoCollectSuccessfullyReadyToPush", "funcName", funcName, "metrics_num", len(metricList), "time_took_seconds", time.Since(start).Seconds(), "metric_addr", cg.KubeletC.Addr)

	go PushWork(cg.PushServerAddr, cg.TimeOutSeconds, metricList, logger, funcName)
}

func CommonComputeFunc(dataMapMolecular map[string]float64, dataMapDenominator map[string]float64, metricMap map[string]dataobj.MetricValue, metricList []dataobj.MetricValue, newMetric string) []dataobj.MetricValue {

	for cUniqueKey, molecularValue := range dataMapMolecular {
		denominatorValue, loaded := dataMapDenominator[cUniqueKey]
		if !loaded {
			continue
		}
		m, loaded := metricMap[cUniqueKey]
		if !loaded {
			continue
		}
		var utilValue float64
		if denominatorValue == 0 {
			utilValue = 0
		} else {
			utilValue = molecularValue / denominatorValue * 100
		}

		m.CounterType = "GAUGE"
		m.ValueUntyped = utilValue
		m.Value = utilValue
		m.Metric = newMetric

		metricList = append(metricList, m)

	}
	return metricList
}

func CpuComputeFunc(dataMapContainerCpuUsage map[string]float64, metricMapContainerCpu map[string]dataobj.MetricValue, dataMapContainerSpecCpuQuota map[string]float64, metricList []dataobj.MetricValue) []dataobj.MetricValue {
	for cUniqueKey, rateContainerCpuUsage := range dataMapContainerCpuUsage {

		m, loaded := metricMapContainerCpu[cUniqueKey]
		if !loaded {
			continue
		}

		quo, loaded := dataMapContainerSpecCpuQuota[cUniqueKey]
		if !loaded {
			continue
		}
		var utilValue float64
		if quo == 0 {
			utilValue = 0
		} else {
			utilValue = rateContainerCpuUsage / quo * 100000 * 100
		}

		m.ValueUntyped = utilValue
		m.Value = utilValue
		m.CounterType = "GAUGE"
		metricList = append(metricList, m)
		if m.Metric == "cpu.util" && m.Value <= 100 {
			newM := m
			newM.Metric = "cpu.idle"
			newM.ValueUntyped = 100 - newM.Value
			metricList = append(metricList, newM)
		}

	}
	return metricList
}
