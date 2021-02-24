package collect

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	config_util "github.com/prometheus/common/config"

	"github.com/n9e/k8s-mon/config"
)

type bucket struct {
	upperBound float64
	count      float64
}

type buckets []bucket

func (b buckets) Len() int           { return len(b) }
func (b buckets) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b buckets) Less(i, j int) bool { return b[i].upperBound < b[j].upperBound }

func ensureMonotonic(buckets buckets) {
	max := buckets[0].count
	for i := 1; i < len(buckets); i++ {
		switch {
		case buckets[i].count > max:
			max = buckets[i].count
		case buckets[i].count < max:
			buckets[i].count = max
		}
	}
}

func coalesceBuckets(buckets buckets) buckets {
	last := buckets[0]
	i := 0
	for _, b := range buckets[1:] {
		if b.upperBound == last.upperBound {
			last.count += b.count
		} else {
			buckets[i] = last
			last = b
			i++
		}
	}
	buckets[i] = last
	return buckets[:i+1]
}

func bucketQuantile(q float64, buckets buckets) float64 {
	if q < 0 {
		return math.Inf(-1)
	}
	if q > 1 {
		return math.Inf(+1)
	}
	if len(buckets) < 1 {
		return math.NaN()
	}
	sort.Sort(buckets)

	if !math.IsInf(buckets[len(buckets)-1].upperBound, +1) {
		return math.NaN()
	}

	buckets = coalesceBuckets(buckets)
	ensureMonotonic(buckets)

	if len(buckets) < 2 {
		return math.NaN()
	}
	observations := buckets[len(buckets)-1].count
	if observations == 0 {
		return math.NaN()
	}
	rank := q * observations
	b := sort.Search(len(buckets)-1, func(i int) bool { return buckets[i].count >= rank })

	if b == len(buckets)-1 {
		return buckets[len(buckets)-2].upperBound
	}
	if b == 0 && buckets[0].upperBound <= 0 {
		return buckets[0].upperBound
	}
	var (
		bucketStart float64
		bucketEnd   = buckets[b].upperBound
		count       = buckets[b].count
	)
	if b > 0 {
		bucketStart = buckets[b-1].upperBound
		count -= buckets[b-1].count
		rank -= buckets[b-1].count
	}
	return bucketStart + (bucketEnd-bucketStart)*(rank/count)
}

func checkFloatValidate(qu float64) (isValidate bool) {
	return math.IsNaN(qu) || math.IsInf(qu, 1) || math.IsInf(qu, 1)
}

func NewMetricFunc(nid string, newMetricName string, value float64, step int64, appendTags map[string]string, metricList []dataobj.MetricValue) []dataobj.MetricValue {

	metric := dataobj.MetricValue{}
	metric.Nid = nid
	metric.Metric = newMetricName
	metric.Timestamp = time.Now().Unix()
	metric.Step = step
	metric.CounterType = config.METRIC_TYPE_GAUGE
	metric.ValueUntyped = value
	metric.Value = value
	metric.TagsMap = appendTags
	metricList = append(metricList, metric)
	return metricList
}

func PercentComputeForKsm(mfenzi map[string]float64, mfenmu map[string]float64, nid string, newMetricName string, sameKeyName string, step int64, appendTags map[string]string, metricList []dataobj.MetricValue) []dataobj.MetricValue {

	for sameKey, fenzi := range mfenzi {
		fenmu, loaded := mfenmu[sameKey]
		if !loaded {
			continue
		}
		var percent float64
		if fenmu == 0 {
			percent = 0
		} else {
			percent = (fenzi / fenmu) * 100
		}

		metricPercent := dataobj.MetricValue{}
		metricPercent.Nid = nid
		metricPercent.Metric = newMetricName + "_percent"
		metricPercent.Timestamp = time.Now().Unix()
		metricPercent.Step = step
		metricPercent.CounterType = config.METRIC_TYPE_GAUGE
		metricPercent.ValueUntyped = percent
		metricPercent.Value = percent
		metricPercent.TagsMap = appendTags
		metricPercent.TagsMap[sameKeyName] = sameKey
		metricList = append(metricList, metricPercent)

		metricValue := dataobj.MetricValue{}
		metricValue.Nid = nid
		metricValue.Metric = newMetricName + "_value"
		metricValue.Timestamp = time.Now().Unix()
		metricValue.Step = step
		metricValue.CounterType = config.METRIC_TYPE_GAUGE
		metricValue.ValueUntyped = fenzi
		metricValue.Value = fenzi
		metricValue.TagsMap = appendTags
		metricValue.TagsMap[sameKeyName] = sameKey

		metricList = append(metricList, metricValue)
	}

	return metricList
}

func avgCompute(m map[string]float64, nid string, metricName string, step int64, appendTags map[string]string) (metricList []dataobj.MetricValue) {
	sum := m["sum"]
	count := m["count"]
	var avg float64
	if count == 0 {
		avg = 0
	} else {
		avg = sum / count
	}
	metric := dataobj.MetricValue{}
	metric.Nid = nid
	metric.Metric = metricName + "_avg"
	metric.Timestamp = time.Now().Unix()
	metric.Step = step
	metric.CounterType = config.METRIC_TYPE_GAUGE
	metric.ValueUntyped = avg
	metric.Value = avg
	metric.TagsMap = appendTags
	metricList = append(metricList, metric)
	return
}

func histogramDeltaWork(dataMap *HistoryMap, bucketM map[float64]float64, newtagsm map[string]string, funcName string, newMetricName string, serverSideNid string, step int64, metricList []dataobj.MetricValue) []dataobj.MetricValue {
	now := time.Now().Unix()
	sepStr := "||"
	newM := make(map[float64]float64)
	for up, count := range bucketM {
		thisCounterStats := CounterStats{Value: count, Ts: now}
		s := NewCommonCounterHis()

		s.UpdateCounterStat(thisCounterStats)
		mapKey := funcName + sepStr + newMetricName + sepStr + fmt.Sprintf("%f", up)
		obj, loaded := dataMap.Map.LoadOrStore(mapKey, s)
		if !loaded {
			continue
		}
		dataHis := obj.(*CommonCounterHis)
		dataHis.UpdateCounterStat(thisCounterStats)
		dataRate := dataHis.DeltaCounter()

		dataMap.Map.Store(mapKey, dataHis)
		newM[up] = dataRate

	}
	if len(newM) > 0 {

		mm := bucketMetrics(newM, serverSideNid, newMetricName, step, newtagsm)
		metricList = append(metricList, mm...)
	}
	return metricList
}

func successfulRate(m map[string]float64, nid string, metricName string, step int64, appendTags map[string]string) (metricList []dataobj.MetricValue) {
	var (
		suSum  float64 = 0
		allSum float64 = 0
		value  float64 = 0
	)
	for label, sum := range m {
		if strings.HasPrefix(label, "2") || strings.HasPrefix(label, "3") {
			suSum += sum
		}
		allSum += sum

	}
	if allSum == 0 {
		value = 0
	} else {
		value = (suSum / allSum) * 100
	}

	metric := dataobj.MetricValue{}
	metric.Nid = nid
	metric.Metric = metricName
	metric.Timestamp = time.Now().Unix()
	metric.Step = step
	metric.CounterType = config.METRIC_TYPE_GAUGE
	metric.ValueUntyped = value
	metric.Value = value
	metric.TagsMap = appendTags
	metricList = append(metricList, metric)
	return
}

func bucketMetrics(m map[float64]float64, nid string, metricName string, step int64, appendTags map[string]string) (metricList []dataobj.MetricValue) {
	//fmt.Println("bucketMetrics", metricName, m)
	if !strings.HasSuffix(metricName, "_bucket") {
		return
	}
	var bks buckets
	for upperBound, count := range m {
		bk := bucket{upperBound: upperBound, count: count}
		bks = append(bks, bk)
	}
	quantile50 := bucketQuantile(0.5, bks)
	quantile90 := bucketQuantile(0.9, bks)
	quantile95 := bucketQuantile(0.95, bks)
	quantile99 := bucketQuantile(0.99, bks)

	qm := make(map[string]float64)
	if !checkFloatValidate(quantile50) {
		qm["50"] = quantile50
	}
	if !checkFloatValidate(quantile90) {
		qm["90"] = quantile90
	}
	if !checkFloatValidate(quantile95) {
		qm["95"] = quantile95
	}
	if !checkFloatValidate(quantile99) {
		qm["99"] = quantile99
	}

	ss := strings.Split(metricName, "_bucket")
	newName := fmt.Sprintf("%s_%s", ss[0], "quantile")
	for qu, value := range qm {
		newTagsM := make(map[string]string)
		newTagsM["quantile"] = qu
		for k, v := range appendTags {
			newTagsM[k] = v
		}

		metric := dataobj.MetricValue{}
		metric.Nid = nid
		metric.Metric = newName
		metric.Timestamp = time.Now().Unix()
		metric.Step = step
		metric.CounterType = config.METRIC_TYPE_GAUGE
		metric.ValueUntyped = value
		metric.Value = value

		metric.TagsMap = newTagsM
		metricList = append(metricList, metric)
	}
	return
}

func GetHostName(logger log.Logger) string {
	name, err := os.Hostname()
	if err != nil {
		level.Error(logger).Log("msg", "GetHostNameError", "err", err)
	}
	return name

}

func GetPortListenAddr(port int64) (portListenAddr string, err error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return "", err
	}
	for _, address := range addrs {

		ipnet, ok := address.(*net.IPNet)
		if !ok {
			continue
		}

		addr := ipnet.IP.To4()
		if addr == nil {
			continue
		}
		// 检查ip地址判断是否回环地址
		if ipnet.IP.IsLoopback() {
			continue
		}
		adds := addr.String()
		addrAndPort := fmt.Sprintf("%s:%d", adds, port)
		conn, err := net.DialTimeout("tcp", addrAndPort, time.Second*1)

		if err != nil {
			continue
		}
		conn.Close()
		portListenAddr = adds
		break

	}
	return
}

func CommonCollectTickerForWithDataM(cg *config.Config, ctx context.Context, logger log.Logger, dataMap *HistoryMap, collectFunc func(*config.Config, log.Logger, *HistoryMap, string), funcName string) error {
	ticker := time.NewTicker(time.Second * (time.Duration(cg.Step)))
	level.Info(logger).Log("msg", "CommonCollectTickerForWithDataM start....", "funcName", funcName)

	collectFunc(cg, logger, dataMap, funcName)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			collectFunc(cg, logger, dataMap, funcName)
		case <-ctx.Done():
			level.Info(logger).Log("msg", "CommonCollectTicker exit....", "funcName", funcName)
			return nil
		}
	}

}

func CommonCollectTicker(cg *config.Config, ctx context.Context, logger log.Logger, collectFunc func(*config.Config, log.Logger, string), funcName string) error {
	ticker := time.NewTicker(time.Second * (time.Duration(cg.Step)))
	level.Info(logger).Log("msg", "CommonCollectTicker start....", "funcName", funcName)

	collectFunc(cg, logger, funcName)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			collectFunc(cg, logger, funcName)
		case <-ctx.Done():
			level.Info(logger).Log("msg", "CommonCollectTicker exit....", "funcName", funcName)
			return nil
		}
	}

}

func GetServerAddrTicker(ctx context.Context, logger log.Logger, step int64, dataMap *HistoryMap) error {
	ticker := time.NewTicker(time.Second * (time.Duration(step)))
	level.Info(logger).Log("msg", "GetServerAddrTicker start....")

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			GetServerAddrAll(logger, dataMap)
		case <-ctx.Done():
			level.Info(logger).Log("msg", "GetServerAddrTicker exit....")
			return nil
		}
	}

}

func MapWhiteMetricsMap(metricsWhiteList []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, mn := range metricsWhiteList {
		m[mn] = struct{}{}
	}
	return m
}

func GetServerSideAddr(cg *config.CommonApiServerConfig, logger log.Logger, dataMap *HistoryMap, funcName string) (metricUrlMap map[string]string) {
	metricUrlMap = make(map[string]string)
	if cg.UserSpecified && len(cg.UserSpecifyAddrs) > 0 {
		for _, murl := range cg.UserSpecifyAddrs {
			u, err := url.Parse(murl)
			if err != nil {
				level.Error(logger).Log("msg", "GetServerSideAddrParseUrlErrorByUserSpecifyAddrs",
					"funcName", funcName,
					"metricUrl", murl,
					"err", err)
				continue
			}
			metricUrlMap[u.Host] = murl
		}

		return
	}
	obj, loaded := dataMap.Map.Load(funcName)
	if !loaded {
		level.Error(logger).Log("msg", "GetServerAddrByGetPodErrorNoValue", "funcName:", funcName)
		return
	}

	ips := obj.([]string)
	if len(ips) == 0 {
		level.Error(logger).Log("msg", "GetServerAddrByGetPodErrorEmptyAddrs", "funcName:", funcName)
		return
	}
	level.Debug(logger).Log("msg", "GetServerAddrByGetPodorNode", "funcName", funcName, "num", len(ips))
	for _, ip := range ips {
		murl := fmt.Sprintf("%s://%s:%d/%s", cg.Scheme, ip, cg.Port, cg.MetricsPath)
		metricUrlMap[ip] = murl
	}

	return
}

func AsyncCurlMetricsAndPush(controlChan chan int, c *config.CommonApiServerConfig, logger log.Logger, funcName string, m map[string]string, step int64, tw int64, index int, allNum int, serverSideNid string, pushServerAddr string, dropBucket bool) {
	start := time.Now()
	defer func() {
		<-controlChan
	}()
	metricList, err := CurlTlsMetricsApi(logger, funcName, c, m, step, tw, dropBucket)

	if err != nil {
		level.Error(logger).Log("msg", "CurlTlsMetricsResError", "func_name", funcName, "err:", err, "seq", fmt.Sprintf("%d/%d", index, allNum), "addr", c.Addr)
		return
	}
	if len(metricList) == 0 {
		level.Error(logger).Log("msg", "CurlTlsMetricsResEmpty", "func_name", funcName, "seq", fmt.Sprintf("%d/%d", index, allNum), "addr", c.Addr)
		return
	}
	ml := make([]dataobj.MetricValue, 0)
	for _, m := range metricList {
		if m.CounterType == config.METRIC_TYPE_COUNTER {
			m.Metric = m.Metric + config.COUNTER_TO_GAUGE_METRIC_NAME_SUFFIX
		}

		m.Nid = serverSideNid
		ml = append(ml, m)

	}
	level.Info(logger).Log("msg", "DoCollectSuccessfullyReadyToPush", "funcName", funcName, "seq", fmt.Sprintf("%d/%d", index, allNum), "metrics_num", len(ml), "time_took_seconds", time.Since(start).Seconds())
	go PushWork(pushServerAddr, tw, ml, logger, funcName)

}

//func CurlMetricsByIps(cg *config.CommonApiServerConfig, logger log.Logger, dataMap *HistoryMap, funcName string, appendTags map[string]string, step int64, tw int64, multiServerInstanceUniqueLabel string) (metricList []dataobj.MetricValue) {
//	metricUrlMap := GetServerSideAddr(cg, logger, dataMap, funcName)
//	if len(metricUrlMap) == 0 {
//		level.Error(logger).Log("msg", "GetServerSideAddrEmpty", "funcName:", funcName)
//		return
//	}
//	seq := 0
//	for uniqueHost, murl := range metricUrlMap {
//		tmp := *cg
//		c := &tmp
//		c.Addr = murl
//		// 添加service_addr tag
//		m := map[string]string{
//			multiServerInstanceUniqueLabel: uniqueHost,
//		}
//		for k, v := range appendTags {
//			m[k] = v
//		}
//
//		ms, err := CurlTlsMetricsApi(logger, funcName, c, m, step, tw)
//
//		if err != nil {
//			level.Error(logger).Log("msg", "CurlTlsMetricsResError", "func_name", funcName, "err:", err, "seq", fmt.Sprintf("%d/%d", seq+1, len(metricUrlMap)), "addr", c.Addr)
//			//return
//		}
//		if len(ms) == 0 {
//			level.Error(logger).Log("msg", "CurlTlsMetricsResEmpty", "func_name", funcName, "seq", fmt.Sprintf("%d/%d", seq+1, len(metricUrlMap)), "addr", c.Addr)
//			//return
//		}
//		metricList = append(metricList, ms...)
//		seq += 1
//	}
//	return
//}

func sum64(hash [md5.Size]byte) uint64 {
	var s uint64

	for i, b := range hash {
		shift := uint64((md5.Size - i - 1) * 8)

		s |= uint64(b) << shift
	}
	return s
}

func ConcurrencyCurlMetricsByIpsSetNid(cg *config.CommonApiServerConfig, logger log.Logger, dataMap *HistoryMap, funcName string, appendTags map[string]string, step int64, tw int64, multiServerInstanceUniqueLabel string, multiFuncUniqueLabel string, serverSideNid string, pushServerAddr string, dropBucket bool) {
	metricUrlMap := GetServerSideAddr(cg, logger, dataMap, funcName)
	if len(metricUrlMap) == 0 {
		level.Error(logger).Log("msg", "GetServerSideAddrEmpty", "funcName:", funcName)
		return
	}
	seq := 0
	if cg.ConcurrencyLimit == 0 {
		cg.ConcurrencyLimit = 10
	}
	controlChan := make(chan int, cg.ConcurrencyLimit)

	for uniqueHost, murl := range metricUrlMap {
		if cg.HashModNum == 0 {
			goto collect
		} else {
			mod := sum64(md5.Sum([]byte(murl))) % cg.HashModNum
			level.Debug(logger).Log("msg", "metricsHashModEnabled",
				"funcName:", funcName,
				"cg.HashModNum:", cg.HashModNum,
				"cg.HashModShard:", cg.HashModShard,
				"mod:", mod,
			)

			if mod != cg.HashModShard {
				continue
			}
			goto collect
		}

	collect:
		controlChan <- 1

		tmp := *cg
		c := &tmp
		c.Addr = murl
		// 添加service_addr tag
		m := map[string]string{
			multiServerInstanceUniqueLabel: uniqueHost,
			multiFuncUniqueLabel:           funcName,
		}
		for k, v := range appendTags {
			m[k] = v
		}
		go AsyncCurlMetricsAndPush(controlChan, c, logger, funcName, m, step, tw, seq+1, len(metricUrlMap), serverSideNid, pushServerAddr, dropBucket)
		seq += 1
	}
	return
}

func CurlTlsMetricsApi(logger log.Logger, funcName string, cg *config.CommonApiServerConfig, appendTagsM map[string]string, step int64, timeout int64, dropBucket bool) ([]dataobj.MetricValue, error) {
	//start := time.Now()
	client, err := config_util.NewClientFromConfig(cg.HTTPClientConfig, funcName, false, false)
	if err != nil {
		level.Error(logger).Log("msg", "NewClientFromConfig error", "funcName", funcName, "addr", cg.Addr, "err", err)
		return nil, err
	}

	client.Timeout = time.Duration(timeout) * time.Second
	req, err := http.NewRequest("GET", cg.Addr, nil)
	resp, err := client.Do(req)
	if err != nil {
		level.Error(logger).Log("msg", "client request error", "funcName", funcName, "err", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		level.Error(logger).Log("msg", "target_scrape_status_code_not_200_maybe_unauthorized", "funcName", funcName, "StatusCode", resp.StatusCode, "Status", resp.Status)
		return nil, errors.New(fmt.Sprintf("server returned HTTP status %s", resp.Status))
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		level.Error(logger).Log("msg", "readbody error", "funcName", funcName, "err", err)
		return nil, err
	}

	metrics, err := ParseCommon(bodyBytes, MapWhiteMetricsMap(cg.MetricsWhiteList), appendTagsM, step, logger, dropBucket)
	return metrics, err
}
