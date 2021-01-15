package collect

import (
	"encoding/json"
	"github.com/prometheus/common/promlog"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestParseInfValue(t *testing.T) {

	file, err := os.Open("m")
	//file, err := os.Open("kubelet.metrics.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	m := map[string]struct{}{
		"rest_client_exec_plugin_ttl_seconds": {},
	}

	promlogConfig := &promlog.Config{}
	logger := promlog.New(promlogConfig)
	tagM := make(map[string]string)
	metrics, err := ParseCommon(content, m, tagM, 10, logger)
	log.Println(metrics, err)
	_, err = json.Marshal(metrics)
	if err != nil {
		log.Println("marshalError", err)
		//level.Error(logger).Log("msg", "HttpPostPushDataJsonMarshalError", "funcName", funcName, "url", url, "err", err)
		//return nil
	}
}
