package collect

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/didi/nightingale/src/common/dataobj"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	fmodel "github.com/open-falcon/falcon-plus/common/model"
)

func valueConv(valueUntyped interface{}) (error, float64) {
	var vv float64
	var err error
	switch cv := valueUntyped.(type) {
	case string:
		vv, err = strconv.ParseFloat(cv, 64)
		if err != nil {
			return err, 0
		}
	case float64:
		vv = cv
	case uint64:
		vv = float64(cv)
	case int64:
		vv = float64(cv)
	case int:
		vv = float64(cv)
	default:
		return errors.New("unknow value type not in  [int|uint64|int64|float64]"), 0
	}
	return nil, vv
}

func newMetricValue(metric string, val interface{}, dataType string, ts int64, tagsMap map[string]string) *dataobj.MetricValue {
	_, v := valueConv(val)

	mv := &dataobj.MetricValue{
		Metric:       metric,
		ValueUntyped: v,
		Value:        v,
		CounterType:  dataType,
		Timestamp:    ts,
		TagsMap:      tagsMap,
	}
	//if err!=nil{
	//
	//}
	return mv
}

func NewGaugeMetric(metric string, val interface{}, ts int64, tagsMap map[string]string) *dataobj.MetricValue {
	return newMetricValue(metric, val, "GAUGE", ts, tagsMap)
}

func NewCounterMetric(metric string, val interface{}, ts int64, tagsMap map[string]string) *dataobj.MetricValue {
	return newMetricValue(metric, val, "COUNTER", ts, tagsMap)
}

func NewSubtractMetric(metric string, val interface{}, ts int64, tagsMap map[string]string) *dataobj.MetricValue {
	return newMetricValue(metric, val, "SUBTRACT", ts, tagsMap)
}

func NewCumulativeMetric(metric string, val interface{}, ts int64, tagsMap map[string]string) *dataobj.MetricValue {
	return NewCounterMetric(metric, val, ts, tagsMap)

	//return NewSubtractMetric(metric, val, ts, tagsMap)
}

func FmtFalconMetricValue(vs []*dataobj.MetricValue, step int64) []*fmodel.MetricValue {
	rt := []*fmodel.MetricValue{}
	for _, v := range vs {
		item := &fmodel.MetricValue{
			Endpoint: v.Endpoint,
			Metric:   v.Metric,
			Value:    v.Value,
			Tags:     v.Tags,
			Step:     step,
		}
		if v.CounterType == "SUBTRACT" {
			item.Type = "COUNTER"
		} else {
			item.Type = v.CounterType
		}
		rt = append(rt, item)
	}
	return rt
}

func ParseCommon(buf []byte, whiteMetricsMap map[string]struct{}, appendTagsMap map[string]string, step int64, logger log.Logger) ([]dataobj.MetricValue, error) {
	var metricList []dataobj.MetricValue
	var parser expfmt.TextParser
	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// Read raw data
	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)

	// Prepare output
	metricFamilies := make(map[string]*dto.MetricFamily)
	metricFamilies, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, fmt.Errorf("reading text format failed: %s", err)
	}
	now = time.Now().Unix()
	// read metrics
	for basename, mf := range metricFamilies {
		// 这里的basename是 metrics一族的name ，在bucket中 是前缀
		if filterIgnoreMetric(basename, whiteMetricsMap) {
			continue
		}
		metrics := []*dataobj.MetricValue{}
		for _, m := range mf.Metric {
			// pass ignore metric

			switch mf.GetType() {
			case dto.MetricType_GAUGE:
				// gauge metric
				metrics = makeCommon(basename, m)
			case dto.MetricType_COUNTER:
				// counter metric
				metrics = makeCommon(basename, m)
			case dto.MetricType_SUMMARY:
				// summary metric
				metrics = makeQuantiles(basename, m)
			case dto.MetricType_HISTOGRAM:
				// histogram metric
				metrics = makeBuckets(basename, m)
			case dto.MetricType_UNTYPED:
				// untyped as gauge
				metrics = makeCommon(basename, m)
			}
			// render endpoint info
			for _, metric := range metrics {
				// drop 所有bucket ，不能处理histogram
				if strings.HasSuffix(metric.Metric, "_bucket") {
					continue
				}
				metric.Tags = makeAppendTags(metric.TagsMap, appendTagsMap)
				metric.Step = step
				// set provided Time, ms to s
				if m.GetTimestampMs() > 0 {
					metric.Timestamp = m.GetTimestampMs() / 1000
				}

				metricList = append(metricList, *metric)
			}
		}
	}

	return metricList, err
}
func filterIgnoreMetric(metricName string, whiteMetricsMap map[string]struct{}) bool {
	if len(whiteMetricsMap) == 0 {
		return false
	}

	if _, loaded := whiteMetricsMap[metricName]; !loaded {
		return true
	}
	return false
}

// Get Quantiles from summary metric
func makeQuantiles(basename string, m *dto.Metric) []*dataobj.MetricValue {
	metrics := []*dataobj.MetricValue{}
	tags := makeLabels(m)

	countName := fmt.Sprintf("%s_count", basename)
	metrics = append(metrics, NewCumulativeMetric(countName, m.GetSummary().SampleCount, now, tags))

	sumName := fmt.Sprintf("%s_sum", basename)
	metrics = append(metrics, NewCumulativeMetric(sumName, m.GetSummary().SampleSum, now, tags))

	for _, q := range m.GetSummary().Quantile {
		tagsNew := make(map[string]string)
		for tagKey, tagValue := range tags {
			tagsNew[tagKey] = tagValue
		}
		if !math.IsNaN(q.GetValue()) {
			tagsNew["quantile"] = fmt.Sprint(q.GetQuantile())

			metrics = append(metrics, NewGaugeMetric(basename, float64(q.GetValue()), now, tagsNew))
		}
	}

	return metrics
}

// Get Buckets from histogram metric
func makeBuckets(basename string, m *dto.Metric) []*dataobj.MetricValue {
	metrics := []*dataobj.MetricValue{}
	tags := makeLabels(m)

	countName := fmt.Sprintf("%s_count", basename)
	metrics = append(metrics, NewCumulativeMetric(countName, m.GetHistogram().SampleCount, now, tags))

	sumName := fmt.Sprintf("%s_sum", basename)
	metrics = append(metrics, NewCumulativeMetric(sumName, m.GetHistogram().SampleSum, now, tags))

	for _, b := range m.GetHistogram().Bucket {
		tagsNew := make(map[string]string)
		for tagKey, tagValue := range tags {
			tagsNew[tagKey] = tagValue
		}
		tagsNew["le"] = fmt.Sprint(b.GetUpperBound())

		bucketName := fmt.Sprintf("%s_bucket", basename)
		metrics = append(metrics, NewGaugeMetric(bucketName, float64(b.GetCumulativeCount()), now, tagsNew))
	}

	return metrics
}

// Get gauge and counter from metric
func makeCommon(metricName string, m *dto.Metric) []*dataobj.MetricValue {
	var val float64
	metrics := []*dataobj.MetricValue{}
	tags := makeLabels(m)
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) && !math.IsInf(m.GetGauge().GetValue(), 1) && !math.IsInf(m.GetGauge().GetValue(), -1) {
			val = float64(m.GetGauge().GetValue())
			metrics = append(metrics, NewGaugeMetric(metricName, val, now, tags))
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) && !math.IsInf(m.GetCounter().GetValue(), 1) && !math.IsInf(m.GetCounter().GetValue(), -1) {
			val = float64(m.GetCounter().GetValue())
			metrics = append(metrics, NewCumulativeMetric(metricName, val, now, tags))
		}
	} else if m.Untyped != nil {
		// untyped as gauge
		if !math.IsNaN(m.GetUntyped().GetValue()) && !math.IsInf(m.GetUntyped().GetValue(), 1) && !math.IsInf(m.GetUntyped().GetValue(), -1) {
			val = float64(m.GetUntyped().GetValue())
			metrics = append(metrics, NewGaugeMetric(metricName, val, now, tags))
		}
	}
	return metrics
}

// Get labels from metric
func makeLabels(m *dto.Metric) map[string]string {
	tags := map[string]string{}
	for _, lp := range m.Label {
		k := lp.GetName()
		v := lp.GetValue()
		if v == "" || k == "" {
			continue
		}
		tags[k] = v
	}
	return tags
}

// append tags
func makeAppendTags(tagsMap map[string]string, appendTagsMap map[string]string) string {
	if len(tagsMap) == 0 && len(appendTagsMap) == 0 {
		return ""
	}

	if len(tagsMap) == 0 {
		return dataobj.SortedTags(appendTagsMap)
	}

	if len(appendTagsMap) == 0 {
		return dataobj.SortedTags(tagsMap)
	}

	for k, v := range appendTagsMap {
		tagsMap[k] = v
	}

	return dataobj.SortedTags(tagsMap)
}
