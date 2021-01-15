package config

import (
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	config_util "github.com/prometheus/common/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	CollectMode   string `yaml:"collect_mode"`
	ServerSideNid string `yaml:"server_side_nid"`
	Step          int64  `yaml:"collect_step" binding:"required"`

	AppendTags                     map[string]string      `yaml:"append_tags"`
	CadvisorC                      *CadvisorConfig        `yaml:"cadvisor"`
	KubeletC                       *CommonApiServerConfig `yaml:"kubelet"`
	CoreDnsC                       *CommonApiServerConfig `yaml:"coredns"`
	KubeStatsC                     *CommonApiServerConfig `yaml:"kube_stats_metrics"`
	ApiServerC                     *CommonApiServerConfig `yaml:"apiserver"`
	KubeSchedulerC                 *CommonApiServerConfig `yaml:"kube_scheduler"`
	KubeControllerC                *CommonApiServerConfig `yaml:"kube_controller_manager"`
	PushServerAddr                 string                 `yaml:"push_server_addr"`
	N9eNidLabelName                string                 `yaml:"n9e_nid_label_name"`
	MultiServerInstanceUniqueLabel string                 `yaml:"multi_server_instance_unique_label"`
	TimeOutSeconds                 int64                  `yaml:"time_out_second"`
}

type CommonApiServerConfig struct {
	IdentifyMode     string                       `yaml:"identify_mode"`
	Addr             string                       `yaml:"addr"`
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
	if sc.CollectMode == "" {
		return errors.New("collect_mode must given choice are [kubelet_agent|cadvisor_plugin|server_side]")
	} else {
		m := map[string]struct{}{
			"kubelet_agent":   {},
			"cadvisor_plugin": {},
			"server_side":     {},
		}
		if _, loaded := m[sc.CollectMode]; !loaded {
			return errors.New("collect_mode must be [kubelet_agent|cadvisor_plugin|server_side]")
		}
	}

	if sc.CollectMode == "kubelet_agent" {
		if sc.KubeletC == nil {
			return errors.New("[collect_mode=kubelet_agent] kubelet config missed")

		}
		if sc.KubeletC.IdentifyMode == "" {
			sc.KubeletC.IdentifyMode = "nid"
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

	}

	if sc.CollectMode == "server_side" {
		if sc.ServerSideNid == "" {
			return errors.New("[collect_mode=server_side] nid must set,exit....")
		}
		if sc.ApiServerC == nil {
			return errors.New("[collect_mode=server_side] apiserver config missed")
		}
		if sc.KubeStatsC == nil {
			return errors.New("[collect_mode=server_side] KubeStatsC config missed")
		}

		if sc.KubeStatsC.HTTPClientConfig == emptyHc {
			sc.KubeStatsC.HTTPClientConfig = defaultHttpC
		}
		if sc.ApiServerC.HTTPClientConfig == emptyHc {
			sc.ApiServerC.HTTPClientConfig = defaultHttpC
		}
		//if sc.KubeSchedulerC.HTTPClientConfig == emptyHc {
		//	sc.KubeSchedulerC.HTTPClientConfig = defaultHttpC
		//}
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

	if sc.Step != 0 {
		viper.SetDefault("g_step", sc.Step)
	} else {
		viper.SetDefault("g_step", 15)
	}
	if sc.N9eNidLabelName == "" {
		sc.N9eNidLabelName = "nid"
	}
	if sc.MultiServerInstanceUniqueLabel == "" {
		sc.MultiServerInstanceUniqueLabel = "server_addr"
	}
	return nil
}
