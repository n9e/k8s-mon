package collect

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/didi/nightingale/src/common/dataobj"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func PushWork(url string, tt int64, metricList []dataobj.MetricValue, logger log.Logger, funcName string) {
	retry := 0
	start := time.Now()
	for {

		resp := HttpPostPushData(url, tt, metricList, logger, funcName)
		if resp == nil {
			return
		}
		if resp.StatusCode == 200 {
			level.Debug(logger).Log("msg", "PushWorkSuccess", "funcName", funcName, "url", url, "metricsNum", len(metricList), "time_took_seconds", time.Since(start).Seconds())
			return
		}
		defer resp.Body.Close()
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			level.Error(logger).Log("msg", "PushWorkReadResp.BodyFailed", "funcName", funcName, "url", url, "metricsNum", len(metricList), "err", err)
			break
		}
		var resV map[string]string

		err = json.Unmarshal(respBytes, &resV)
		if err != nil {
			level.Error(logger).Log("msg", "PushWorkJson.UnmarshalFailed", "funcName", funcName, "url", url, "metricsNum", len(metricList), "err", err)
			break
		}
		resS := ""
		for k, v := range resV {
			resS = k + ":" + v
		}

		level.Warn(logger).Log("msg", "PushWorkFailed", "funcName", funcName, "url", url, "rc", resp.StatusCode, "respErrorStr", resS)
		time.Sleep(time.Millisecond * 500)

		retry += 1
		if retry == 3 {
			break
		}
	}

}

func HttpPostPushData(url string, tt int64, data interface{}, logger log.Logger, funcName string) *http.Response {

	bytesData, err := json.Marshal(data)
	if err != nil {
		level.Error(logger).Log("msg", "HttpPostPushDataJsonMarshalError", "funcName", funcName, "url", url, "err", err)
		return nil
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	// http短连接
	//request.Close = true
	if err != nil {

		level.Error(logger).Log("msg", "HttpPostPushDataBuildNewHttpPostReqError1", "funcName", funcName, "url", url, "err", err)
		return nil
	}
	header := http.Header{}
	request.Header = header
	client := http.Client{}
	client.Timeout = time.Second * time.Duration(tt)
	resp, err := client.Do(request)
	if err != nil {
		level.Error(logger).Log("msg", "HttpPostPushDataBuildNewHttpPostReqError2", "funcName", funcName, "url", url, "err", err)
		return nil
	}
	return resp

}
