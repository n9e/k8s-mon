package collect

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	config_util "github.com/prometheus/common/config"

	"github.com/n9e/k8s-mon/config"
)

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

func AsyncCurlMetricsAndPush(controlChan chan int, c *config.CommonApiServerConfig, logger log.Logger, funcName string, m map[string]string, step int64, tw int64, index int, allNum int, serverSideNid string, pushServerAddr string) {
	start := time.Now()
	defer func() {
		<-controlChan
	}()
	metricList, err := CurlTlsMetricsApi(logger, funcName, c, m, step, tw)

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

func ConcurrencyCurlMetricsByIpsSetNid(cg *config.CommonApiServerConfig, logger log.Logger, dataMap *HistoryMap, funcName string, appendTags map[string]string, step int64, tw int64, multiServerInstanceUniqueLabel string, multiFuncUniqueLabel string, serverSideNid string, pushServerAddr string) {
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
		if cg.HashModNum == 0 && cg.HashModShard == 0 {
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
		go AsyncCurlMetricsAndPush(controlChan, c, logger, funcName, m, step, tw, seq+1, len(metricUrlMap), serverSideNid, pushServerAddr)
		seq += 1
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
