package collect

import (
	"context"
	"fmt"
	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/n9e/k8s-mon/config"
	config_util "github.com/prometheus/common/config"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

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
			GetServerAddrByGetPod(logger, dataMap)
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
func CurlMetricsByIps(cg *config.CommonApiServerConfig, logger log.Logger, dataMap *HistoryMap, funcName string, appendTags map[string]string, step int64, tw int64, multiServerInstanceUniqueLabel string) (metricList []dataobj.MetricValue) {
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

	level.Info(logger).Log("msg", "GetServerAddrByGetPod", "funcName", funcName, "num", len(ips))

	for index, addr := range ips {
		tmp := *cg
		c := &tmp
		c.Addr = fmt.Sprintf("%s://%s:%d/%s", c.Scheme, addr, c.Port, c.MetricsPath)
		// 添加service_addr tag
		m := map[string]string{
			multiServerInstanceUniqueLabel: addr,
		}
		for k, v := range appendTags {
			m[k] = v
		}

		ms, err := CurlTlsMetricsApi(logger, funcName, c, m, step, tw)

		if err != nil {
			level.Error(logger).Log("msg", "CurlTlsMetricsResError", "func_name", funcName, "err:", err, "seq", fmt.Sprintf("%d/%d", index+1, len(ips)), "addr", c.Addr)
			//return
		}
		if len(ms) == 0 {
			level.Error(logger).Log("msg", "CurlTlsMetricsResEmpty", "func_name", funcName, "seq", fmt.Sprintf("%d/%d", index+1, len(ips)), "addr", c.Addr)
			//return
		}
		metricList = append(metricList, ms...)
	}
	return
}

func CurlTlsMetricsApi(logger log.Logger, funcName string, cg *config.CommonApiServerConfig, appendTagsM map[string]string, step int64, timeout int64) ([]dataobj.MetricValue, error) {
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
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		level.Error(logger).Log("msg", "readbody error", "funcName", funcName, "err", err)
		return nil, err
	}

	metrics, err := ParseCommon(bodyBytes, MapWhiteMetricsMap(cg.MetricsWhiteList), appendTagsM, step, logger)
	//fmt.Println("cg.MetricsWhiteList", cg.MetricsWhiteList)
	//level.Info(logger).Log("msg", "CurlTlsMetricsApi and ParseCommon time_took", "funcName", funcName, "addr", cg.Addr, "time_took_seconds", time.Since(start).Seconds())
	return metrics, err
}
