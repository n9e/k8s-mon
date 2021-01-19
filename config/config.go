package config

import (
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	config_util "github.com/prometheus/common/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
)

type Config struct {
	CollectMode                    string                 `yaml:"collect_mode"`
	ServerSideNid                  string                 `yaml:"server_side_nid"`
	Step                           int64                  `yaml:"collect_step" binding:"required"`
	PushServerAddr                 string                 `yaml:"push_server_addr"`
	N9eNidLabelName                string                 `yaml:"n9e_nid_label_name"`
	MultiServerInstanceUniqueLabel string                 `yaml:"multi_server_instance_unique_label"`
	MultiFuncUniqueLabel           string                 `yaml:"multi_func_unique_label"`
	TimeOutSeconds                 int64                  `yaml:"time_out_second"`
	AppendTags                     map[string]string      `yaml:"append_tags"`
	CadvisorC                      *CadvisorConfig        `yaml:"cadvisor"`
	KubeletC                       *CommonApiServerConfig `yaml:"kubelet"`
	CoreDnsC                       *CommonApiServerConfig `yaml:"coredns"`
	KubeStatsC                     *CommonApiServerConfig `yaml:"kube_stats_metrics"`
	ApiServerC                     *CommonApiServerConfig `yaml:"apiserver"`
	KubeSchedulerC                 *CommonApiServerConfig `yaml:"kube_scheduler"`
	KubeControllerC                *CommonApiServerConfig `yaml:"kube_controller_manager"`
	KubeProxyC                     *CommonApiServerConfig `yaml:"kube_proxy"`
	KubeletNodeC                   *CommonApiServerConfig `yaml:"kubelet_node"`
}

type CommonApiServerConfig struct {
	HashModNum       uint64                       `yaml:"hash_mod_num"`
	HashModShard     uint64                       `yaml:"hash_mod_shard"`
	ConcurrencyLimit int64                        `yaml:"concurrency_limit"`
	IdentifyMode     string                       `yaml:"identify_mode"`
	Addr             string                       `yaml:"-"`
	UserSpecified    bool                         `yaml:"user_specified"`
	UserSpecifyAddrs []string                     `yaml:"addrs"`
	Scheme           string                       `yaml:"scheme"`
	MetricsPath      string                       `yaml:"metrics_path"`
	Port             int64                        `yaml:"port"`
	MetricsWhiteList []string                     `yaml:"metrics_white_list"`
	TagsWhiteList    []string                     `yaml:"tags_white_list"`
	HTTPClientConfig config_util.HTTPClientConfig `yaml:",inline"`
}

type CadvisorConfig struct {
	Addr           string `yaml:"addr"`
	IdentifyMode   string `yaml:"identify_mode"`
	DockerSockFile string `yaml:"docker_sock_file"`
}

func Load(s string) (*Config, error) {
	cfg := &Config{}

	//err := yaml.UnmarshalStrict([]byte(s), cfg)
	err := yaml.Unmarshal([]byte(s), cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadFile(filename string, logger log.Logger) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg, err := Load(string(content))
	if err != nil {
		level.Error(logger).Log("msg", "parsing YAML file errr...", "error", err)
		return nil, err
	}
	err = setDefaultVarAndValidate(cfg)
	return cfg, err
}

func setDefaultVarAndValidate(sc *Config) error {
	defaultHttpC := config_util.HTTPClientConfig{
		BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		TLSConfig: config_util.TLSConfig{
			CAFile:             "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
			InsecureSkipVerify: true,
		},
	}

	emptyHc := config_util.HTTPClientConfig{}

	switch sc.CollectMode {
	case COLLECT_MODE_CADVISOR_PLUGIN:
		if sc.CadvisorC == nil {
			return errors.New("[collect_mode=cadvisor_plugin]  cadvisor config missed")
		}
		if sc.CadvisorC != nil {
			if sc.CadvisorC.DockerSockFile != "" {
				viper.SetDefault("g_docker_sock_file", sc.CadvisorC.DockerSockFile)
			} else {
				viper.SetDefault("g_docker_sock_file", "/var/run/docker.sock")
			}
			if sc.CadvisorC.Addr != "" {
				viper.SetDefault("g_cadvisor_addr", sc.CadvisorC.Addr)
			} else {
				viper.SetDefault("g_cadvisor_addr", "http://127.0.0.1:4194/")
			}

		}
	case COLLECT_MODE_KUBELET_AGENT:
		if sc.PushServerAddr == "" {
			return errors.New("[n9e push addr empty]")
		}
		if sc.KubeletC == nil {
			return errors.New("[collect_mode=kubelet_agent] kubelet config missed")

		}
		m := map[string]struct{}{
			"nid":      {},
			"endpoint": {},
		}
		if _, loaded := m[sc.KubeletC.IdentifyMode]; !loaded {
			return errors.New("[collect_mode=kubelet_agent] identify_mode must be [nid|endpoint]")
		}
		if sc.KubeletC.HTTPClientConfig == emptyHc {
			sc.KubeletC.HTTPClientConfig = defaultHttpC
		}

	case COLLECT_MODE_SERVER_SIDE:
		if sc.PushServerAddr == "" {
			return errors.New("[n9e push addr empty]")
		}
		if sc.ServerSideNid == "" {
			return errors.New("[collect_mode=server_side] nid must set,exit....")
		}
		if err := checkValidateForServerSide(sc.ApiServerC, FUNCNAME_APISERVER); err != nil {
			return err
		}
		if err := checkValidateForServerSide(sc.KubeSchedulerC, FUNCNAME_KUBESCHEDULER); err != nil {
			return err
		}
		if err := checkValidateForServerSide(sc.KubeControllerC, FUNCNAME_KUBECONTROLLER); err != nil {
			return err
		}
		if err := checkValidateForServerSide(sc.CoreDnsC, FUNCNAME_COREDNS); err != nil {
			return err
		}
		if err := checkValidateForServerSide(sc.KubeProxyC, FUNCNAME_KUBEPROXY); err != nil {
			return err
		}
		if err := checkValidateForServerSide(sc.KubeStatsC, FUNCNAME_KUBESTATSMETRICS); err != nil {
			return err
		}

	default:
		return errors.New("collect_mode must be one of [kubelet_agent|cadvisor_plugin|server_side]")
	}

	// set default

	setGlobalDefaultVars(sc)
	return nil
}

func checkValidateForServerSide(sc *CommonApiServerConfig, funcName string) error {
	if sc == nil {
		return nil
	}

	defaultHttpC := config_util.HTTPClientConfig{
		BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		TLSConfig: config_util.TLSConfig{
			CAFile:             "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
			InsecureSkipVerify: true,
		},
	}

	emptyHc := config_util.HTTPClientConfig{}

	if sc.UserSpecified {
		if len(sc.UserSpecifyAddrs) == 0 {
			return errors.New(fmt.Sprintf("[collect_mode=server_side][%s.user_specified=true][user_specify_addrs must given ]", funcName))
		}
		for _, x := range sc.UserSpecifyAddrs {
			_, err := url.Parse(x)
			if err != nil {
				return errors.New(fmt.Sprintf("[collect_mode=server_side][%s.user_specified=true][UserSpecifyAddrs url parse error][url:%s,err:%s]",
					funcName,
					x,
					err.Error(),
				))
			}
		}

	} else {
		if sc.MetricsPath == "" || sc.Scheme == "" || sc.Port == 0 {
			return errors.New(fmt.Sprintf("[collect_mode=server_side][%s.user_specified=false][scheme=%s|metrics_path=%s|port=%d must set together]",
				sc.Scheme,
				sc.MetricsPath,
				sc.Port,
				funcName))
		}
	}

	if sc.HTTPClientConfig == emptyHc {
		sc.HTTPClientConfig = defaultHttpC
	}

	return nil
}

func setGlobalDefaultVars(sc *Config) {
	if sc.Step != 0 {
		viper.SetDefault("g_step", sc.Step)
	} else {
		viper.SetDefault("g_step", 30)
	}
	if sc.N9eNidLabelName == "" {
		sc.N9eNidLabelName = DEFAULT_N9ENIDLABELNAME
	}
	if sc.MultiServerInstanceUniqueLabel == "" {
		sc.MultiServerInstanceUniqueLabel = APPENDTAG_SERVER_ADDR
	}

	if sc.MultiFuncUniqueLabel == "" {
		sc.MultiFuncUniqueLabel = APPENDTAG_FUNC_NAME
	}

	if sc.TimeOutSeconds == 0 {
		sc.TimeOutSeconds = 10
	}

}
