package collect

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/cadvisor/client"
	v1 "github.com/google/cadvisor/info/v1"
	"github.com/spf13/viper"
	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"

	"github.com/n9e/k8s-mon/config"
)

type Metric struct {
	Nid          string            `json:"nid"`
	Metric       string            `json:"metric"`
	Timestamp    int64             `json:"timestamp"`
	Step         int64             `json:"step"`
	ValueUntyped interface{}       `json:"value"`
	Value        float64           `json:"-"`
	CounterType  string            `json:"counterType"`
	Tags         string            `json:"tags"`
	TagsMap      map[string]string `json:"tagsMap"` //保留2种格式，方便后端组件使用
}

type ContainerInspect struct {
	ID       string
	Host     string
	Nid      string
	Type     string
	CpuQuota float64
	MemQuota float64
}

type DockerStats struct {
	PidsStats Pids `json:"pids_stats"`
}

type Pids struct {
	Current int `json:"current"`
}

type HostIpMap struct {
	sync.RWMutex
	M map[string]string
}

func (h *HostIpMap) Load(f string) error {
	h.Lock()
	defer h.Unlock()

	b, err := file.ToBytes(f)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &h.M)
}

func (h *HostIpMap) WriteFile(f string) error {
	h.Lock()
	defer h.Unlock()

	b, err := json.Marshal(h.M)
	if err != nil {
		return err
	}

	_, err = file.WriteBytes(f, b)
	return err
}

func (h *HostIpMap) Set(host, ip string) {
	h.Lock()
	defer h.Unlock()
	h.M[host] = ip
}

func (h *HostIpMap) Get(host string) (string, bool) {
	h.RLock()
	defer h.RUnlock()
	ip, exists := h.M[host]
	return ip, exists
}

type ContainerData struct {
	sync.RWMutex
	Res []Metric
}

func (c *ContainerData) ContainerDataAppend(dat ...Metric) {
	c.Lock()
	defer c.Unlock()
	for _, d := range dat {
		if d.Nid != "" {
			c.Res = append(c.Res, d)
		}
	}
}

var (
	Data          map[string]*ContainerInspect
	containerData = &ContainerData{Res: make([]Metric, 0)}
	now           int64
)

func CollectCadvisorM(cg *config.CadvisorConfig, logger log.Logger) {
	datas, err := getContainersData(cg.Addr)
	if err != nil {
		level.Error(logger).Log("msg", "get containers data from cadvisor raw api", "error", err)
		return
	}
	now = time.Now().Unix()

	var wg sync.WaitGroup

	for i := 0; i < len(datas); i++ {
		item := datas[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			collectDockerStats(item, cg.IdentifyMode)
		}()
	}

	wg.Wait()
	dat, err := json.Marshal(containerData.Res)
	if err != nil {
		level.Error(logger).Log("msg", "  marshal containers datafailed", "error", err)
		return
	}
	fmt.Print(string(dat))
}

func collectDockerStats(container v1.ContainerInfo, identifyMode string) {
	stats := container.Stats
	if len(container.Stats) < 2 {
		logger.Error("container.Stats num < 2")
		return
	}
	inspect, err := getInspect(container.Id, identifyMode)
	if err != nil {
		logger.Errorf("cannot inspect container[%s], error: %v", container.Id, err)
		return
	}

	endpoint := inspect.Host
	nid := inspect.Nid
	duration := float64(stats[1].Timestamp.Unix() - stats[0].Timestamp.Unix())

	restartCount := 0
	restartCountStr := container.Spec.Labels["io.kubernetes.container.restartCount"] // 获取容器的名字
	if restartCountStr != "" {
		restartCount, err = strconv.Atoi(restartCountStr)
		if err != nil {
			logger.Warningf("get %s restartCount err %v", container.Id, err)
		}
	}
	containerData.ContainerDataAppend(toDockerMetric("sys.restart.count", nid, endpoint, restartCount, nil))

	containerName := container.Spec.Labels["io.kubernetes.container.name"] // 获取容器的名字
	if containerName == "POD" {
		hostname := container.Spec.Labels["io.kubernetes.pod.name"]
		// 如果是pause容器就忽略,只上报net指标
		containerData.ContainerDataAppend(netMetric(stats, nid, hostname)...)
		return
	} else if containerName == "" {
		// 如果是其他类型容器，继续采集net指标
		containerData.ContainerDataAppend(netMetric(stats, nid, endpoint)...)
	}

	containerData.ContainerDataAppend(cpuMetric(stats, nid, endpoint, inspect.CpuQuota, duration)...)

	containerData.ContainerDataAppend(memMetric(stats, nid, endpoint, inspect.MemQuota)...)

	containerData.ContainerDataAppend(ioMetric(stats, nid, endpoint, duration)...)

	containerData.ContainerDataAppend(fsMetric(stats, nid, endpoint)...)

	containerData.ContainerDataAppend(sysMetric(stats, nid, endpoint)...)
}

func getInspect(containerId, identifyMode string) (containerInspect *ContainerInspect, err error) {
	containerInspect = new(ContainerInspect)

	client := &http.Client{
		Transport: &http.Transport{
			Dial: getDial,
		},
		Timeout: time.Duration(2 * time.Second),
	}

	url := "http://localhost/containers/" + containerId + "/json"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("Http.NewRequest failed, url: %s, Err: %v", url, err)
		return
	}
	req.Close = true // If we don't do this, it will exhaust all the socket

	resp, err := client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		err = fmt.Errorf("Request from Unix socket failed! Err: %v", err)
		return
	}

	var inspect types.ContainerJSON
	err = json.NewDecoder(resp.Body).Decode(&inspect)
	if err != nil {
		err = fmt.Errorf("json Decode failed, err: %v", err)
		return
	}

	if inspect.Config == nil {
		err = fmt.Errorf("inspect %s config is nil", containerId)
		return
	}

	envs := inspect.Config.Env
	if envs != nil {
		for i := 0; i < len(envs); i++ {
			if strings.Contains(envs[i], "N9E_NID") {
				containerInspect.Nid = strings.TrimSpace(strings.Split(envs[i], "=")[1])
			}
		}
	}

	if containerInspect.Nid == "" {
		logger.Debugf("N9E_NID not found:%+v", inspect.Config)
	}

	containerInspect.ID = inspect.ID
	if identifyMode == "hostname" {
		containerInspect.Host = inspect.Config.Hostname
	} else {
		if envs != nil {
			for i := 0; i < len(envs); i++ {
				if strings.Contains(envs[i], "CCP_POD_IP") {
					containerInspect.Host = strings.TrimSpace(strings.Split(envs[i], "=")[1])
				}
			}
		}

	}

	containerInspect.MemQuota = float64(inspect.HostConfig.Memory)
	cpuPeriod := inspect.HostConfig.CPUPeriod
	if cpuPeriod != 0 {
		containerInspect.CpuQuota = float64(inspect.HostConfig.CPUQuota / cpuPeriod)
	} else {
		containerInspect.CpuQuota = 0
	}

	return
}

func getDial(network, addr string) (net.Conn, error) {

	sockFile := viper.GetString("g_docker_sock_file")
	if !file.IsExist(sockFile) {
		return nil, fmt.Errorf("sockFile is not exists")
	}
	return net.Dial("unix", sockFile)
}

type rwStats struct {
	r uint64
	w uint64
}

func sysMetric(stats []*v1.ContainerStats, nid, endpoint string) (res []Metric) {
	ps := stats[1].Processes
	res = append(res, toDockerMetric("sys.ps.process.used", nid, endpoint, ps.ProcessCount, nil))
	res = append(res, toDockerMetric("sys.ps.thread.used", nid, endpoint, ps.ProcessCount, nil))
	res = append(res, toDockerMetric("sys.fd.count.used", nid, endpoint, ps.FdCount, nil))
	res = append(res, toDockerMetric("sys.socket.count.used", nid, endpoint, ps.SocketCount, nil))

	return res
}

func fsMetric(stats []*v1.ContainerStats, nid, endpoint string) (res []Metric) {
	devices := stats[1].Filesystem
	for _, device := range devices {
		tag := map[string]string{"device": device.Device}

		res = append(res, toDockerMetric("disk.bytes.total", nid, endpoint, device.Limit, tag))
		res = append(res, toDockerMetric("disk.bytes.used", nid, endpoint, device.Usage, tag))
		res = append(res, toDockerMetric("disk.bytes.used.percent", nid, endpoint, float64(device.Usage)/float64(device.Limit), tag))
	}

	return res
}

func ioMetric(stats []*v1.ContainerStats, nid, endpoint string, duration float64) (res []Metric) {
	l := len(stats)
	if l < 2 {
		logger.Warning("len(stats) < 2")
		return
	}

	preDm := make(map[string]rwStats) // device : r, w stat
	curDm := make(map[string]rwStats)

	// get pre value
	for _, pd := range stats[l-2].DiskIo.IoServiceBytes {
		preDm[pd.Device] = rwStats{r: pd.Stats["Read"], w: pd.Stats["Write"]}
	}

	// get cur value
	for _, pd := range stats[l-1].DiskIo.IoServiceBytes {
		curDm[pd.Device] = rwStats{r: pd.Stats["Read"], w: pd.Stats["Write"]}
	}

	for device, rw := range curDm {
		tag := map[string]string{"device": path.Base(device)}
		var ioRead, ioWrite float64
		if pre, exist := preDm[device]; exist {
			ioRead = float64(rw.r-pre.r) / duration
			ioWrite = float64(rw.w-pre.w) / duration
		} else {
			ioRead = float64(rw.r-0) / duration
			ioWrite = float64(rw.w-0) / duration
		}
		res = append(res, toDockerMetric("disk.io.read.bytes", nid, endpoint, ioRead, tag))
		res = append(res, toDockerMetric("disk.io.write.bytes", nid, endpoint, ioWrite, tag))
	}

	return res
}

func memMetric(stats []*v1.ContainerStats, nid, endpoint string, memQuota float64) (res []Metric) {
	cache := stats[1].Memory.Cache
	rss := stats[1].Memory.RSS
	swap := stats[1].Memory.Swap
	memoryUsage := stats[1].Memory.Usage - cache
	memoryUsedPercent := 0.0

	if memQuota > 0 {
		memoryUsedPercent = float64(stats[1].Memory.Usage-stats[1].Memory.Cache) / memQuota * 100
	}

	res = append(res, toDockerMetric("mem.bytes.total", nid, endpoint, memQuota, nil))
	res = append(res, toDockerMetric("mem.bytes.used", nid, endpoint, memoryUsage, nil))
	res = append(res, toDockerMetric("mem.bytes.used.percent", nid, endpoint, memoryUsedPercent, nil))

	res = append(res, toDockerMetric("mem.bytes.cached", nid, endpoint, cache, nil))
	res = append(res, toDockerMetric("mem.bytes.rss", nid, endpoint, rss, nil))
	res = append(res, toDockerMetric("mem.bytes.swap", nid, endpoint, swap, nil))

	return res

}

func cpuMetric(stats []*v1.ContainerStats, nid, endpoint string, cpuQuota, duration float64) (res []Metric) {

	var cpuUser, cpuSys, cpuUsage float64
	cpuIdle := 100.0
	newStats := stats[1].Cpu.Usage
	oldStats := stats[0].Cpu.Usage

	if cpuQuota > 0 {
		cpuUser = delta(newStats.User, oldStats.User) / 1000000000 / duration / cpuQuota * 100 // unit:nanoseconds,10-9s.用户态CPU占用的时间,除以监控间隔等于平均每秒cpuuser使用时间,如2s(4核),最后除以quota变为每秒占用cpu百分比:2/4=50%
		cpuSys = delta(newStats.System, oldStats.System) / 1000000000 / duration / cpuQuota * 100
		cpuUsage = delta(newStats.Total, oldStats.Total) / 1000000000 / duration / cpuQuota * 100 //时间间隔内平均每秒钟CPU使用总量(unit:s)
		if cpuUsage > 100 {
			cpuIdle = 0.0
			logger.Error("cpuUsage > 100", nid, endpoint, cpuUsage)
		} else {
			cpuIdle = 100.0 - cpuUsage
			if cpuIdle < 0 {
				cpuIdle = 0
			}
		}

	}

	periods := stats[1].Cpu.CFS.Periods
	throttledPeriods := stats[1].Cpu.CFS.ThrottledPeriods
	throttledTime := stats[1].Cpu.CFS.ThrottledTime / 1000000000

	res = append(res, toDockerMetric("cpu.user", nid, endpoint, cpuUser, nil))
	res = append(res, toDockerMetric("cpu.sys", nid, endpoint, cpuSys, nil))
	res = append(res, toDockerMetric("cpu.idle", nid, endpoint, cpuIdle, nil))
	res = append(res, toDockerMetric("cpu.util", nid, endpoint, cpuUsage, nil))
	res = append(res, toDockerMetric("cpu.periods", nid, endpoint, periods, nil))
	res = append(res, toDockerMetric("cpu.throttled_periods", nid, endpoint, throttledPeriods, nil))
	res = append(res, toDockerMetric("cpu.throttled_time", nid, endpoint, throttledTime, nil))

	return res
}

func netMetric(stats []*v1.ContainerStats, nid, endpoint string) (res []Metric) {
	duration := float64(stats[1].Timestamp.Unix() - stats[0].Timestamp.Unix())
	newStats := stats[1].Network.InterfaceStats
	oldStats := stats[0].Network.InterfaceStats

	netIn := delta(newStats.RxBytes, oldStats.RxBytes) * 8 / duration
	netInPackets := delta(newStats.RxPackets, oldStats.RxPackets) / duration
	netInErrors := delta(newStats.RxErrors, oldStats.RxErrors) / duration
	netInDropped := delta(newStats.RxDropped, oldStats.RxDropped) / duration
	netOut := delta(newStats.TxBytes, oldStats.TxBytes) * 8 / duration
	netOutPackets := delta(newStats.TxPackets, oldStats.TxPackets) / duration
	netOutErrors := delta(newStats.TxErrors, oldStats.TxErrors) / duration
	netOutDropped := delta(newStats.TxDropped, oldStats.TxDropped) / duration

	res = append(res, toDockerMetric("net.sockets.tcp.timewait", nid, endpoint, stats[1].Network.Tcp.TimeWait, nil))
	res = append(res, toDockerMetric("net.in.bits", nid, endpoint, netIn, nil))
	res = append(res, toDockerMetric("net.in.pps", nid, endpoint, netInPackets, nil))
	res = append(res, toDockerMetric("net.in.errs", nid, endpoint, netInErrors, nil))
	res = append(res, toDockerMetric("net.in.dropped", nid, endpoint, netInDropped, nil))
	res = append(res, toDockerMetric("net.out.bits", nid, endpoint, netOut, nil))
	res = append(res, toDockerMetric("net.out.pps", nid, endpoint, netOutPackets, nil))
	res = append(res, toDockerMetric("net.out.errs", nid, endpoint, netOutErrors, nil))
	res = append(res, toDockerMetric("net.out.dropped", nid, endpoint, netOutDropped, nil))

	res = append(res, toDockerMetric("net.tcp.established", nid, endpoint, stats[1].Network.Tcp.Established, nil))

	return
}

func delta(now, last uint64) float64 {
	if now < last {
		return 0
	}

	return float64(now - last)
}

func toDockerMetric(name, nid, endpoint string, value interface{}, tags map[string]string) Metric {
	ret := Metric{Nid: nid, Metric: name, ValueUntyped: value, TagsMap: map[string]string{}, Step: int64(viper.GetInt("g_step")), Timestamp: now, CounterType: config.GAUGE}
	for k, v := range tags {
		ret.TagsMap[k] = v
	}
	ret.TagsMap["podname"] = endpoint
	return ret
}

func getContainersData(cadUrl string) (containerData []v1.ContainerInfo, err error) {
	clientCon, err := client.NewClient(cadUrl)
	if err != nil {
		return
	}

	query := &v1.ContainerInfoRequest{NumStats: 2} // 取cadvisor最近的两个值,后一个值是最新值,用后值减前值
	containerData, err = clientCon.AllDockerContainers(query)
	return
}
