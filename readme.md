# 问题排查
- 遇到问题请先看，[问题排查文档](./问题排查.md)  ，其中汇总了场景问题和排查思路。


# 模式说明
- 对应配置项为collect_mode `cadvisor_plugin | kubelet_agent | server_side` 三选一
- 代码为同一套代码

|  模式名称  |  部署运行方式 |  collect_mode配置   |  说明 | 
|  ----  | ----  | ---- |---| 
| 夜莺插件形式采集cadvisor raw api	| 可执行的插件由夜莺agent调用|	cadvisor_plugin | 文档在readme最下面 (原有cadvisor采集模式) |  
| 容器基础资源指标采集	|  k8s daemonset 部署在每一个node上|	kubelet_agent | 统称为新模式  (kubelet地址由对应metrics port listen地址决定) |  
| 集中采集k8s服务组件	|k8s deployment 部署 |	server_side | 统称为新模式  |  

# k8s-mon对比prometheus的优点
- Histogram指标分位值预计算，节约存储、降低服务端压力(高基数指标往往来自histogram的海量bucket，同时除了算分位值以外的应用如分布情况又较少)
- 基础资源指标预计算，counterTogauge 简化最终查看的表达式
- 容器基础资源/k8s资源指标和夜莺树绑定


# 新模式：部署在k8s中采集相关指标
## 原理说明

- 通过抓取各个组件的/metrics接口获得prometheus的数据，ETL后push到夜莺
- 各个采集项目有metirc和tag的白名单过滤
- `cadvisor`数据需要hold点做计算比率类指标，多用在百分比的情况，其余不需要
- `counter`类型将有夜莺agent转换为`gauge`型，即数值已经转为`rate`了 ,**所有`counter`类型metric_name 加`_rate`后缀**
- 指标说明在`metrics-detail`文件夹里
- k8s yaml配置在 `k8s-config`中
- 服务组件监控时多实例问题：用户无需关心，k8s-mon自动发现并采集 
- 采集每node上的`kube-proxy` `kubelet-node`指标时支持并发数配置和静态分片
- 服务组件采集时会添加`func_name`标签作为区分具体组件任务，类似`prometheus`的`job`标签 
- 基础指标添加`node_ip` ,`node_name`作为宿主机标识标签
- ksm指标没有nid的默认上报到服务节点`server_side_nid` ，例如`kube_node_status_allocatable_cpu_cores`这种共享指标
- **服务组件采集预聚合了一些指标，包括 分位值、平均值、成功率**，对应文档在 `metrics-detail/preaggregation.md` 
- 服务组件采集了对应golang 进程的指标 包括 内存、goroutine等 ，对应文档在 `metrics-detail/process-resource.md`

## 采集内容说明

- 一般来说在k8s集群汇总我们关注一下4类指标 

|  指标类型  |  采集源 |  应用举例  |  部署方式 | 
|  ----  | ----  | ---- |---| 
| 容器基础资源指标	| kubelet 内置cadvisor |	查看容器cpu、mem等 | k8s daemonset   |  
| k8s资源指标	| [kube-stats-metrics](https://github.com/kubernetes/kube-state-metrics) (简称ksm)|	查看pod状态、查看deployment信息等 | k8s deployment (需要提前部署ksm)   |  
| k8s服务组件指标	|各个服务组件的metrics接口(多实例自动发现)<br> apiserver <br> kube-controller-manager <br> kube-scheduler <br> etcd <br> coredns <br> kube-proxy <br> kubelet-node |	查看请求延迟/QPS等 | 和ksm同一套代码，部署在  k8s deployment   |  
| 业务指标(暂不支持)	| pod暴露的metrics接口|	- | -  |  

## 采集地址配置/发现说明
- 每种项目配置了相关配置段才会开启，如果不想采集某类指标可以去掉其配置
- 每种项目由 `user_specified` 配置是否采用用户指定的地址，用来处理有些服务组件以裸进程形式部署无法从内部发现的case
- 当`user_specified：true`时，对应的`addrs`为采集地址url列表
- 当`user_specified：false`时，则认为由内置的代码来进行动态发现，需要配置好对应的`port` `schema` `metrics_path`等信息

|  采集类型  |  采集地址说明 |  配置/发现说明  | 
|  ----  | ----  | ---- |
| 容器基础资源指标 kubelet-cadvisor	| kubelet 在node上listen分两种情况:<br>  listen 0.0.0.0 <br> listen机器内网ip |	默认为`k8s-mon`自动根据配置的`port`找到对应的地址 |
| k8s资源指标	kube-stats-metrics| 默认为通过coredns 访问service `http://kube-state-metrics.kube-system:8080/metrics` | 同时支持指定  |
| k8s服务组件指标(master侧) <br> apiserver <br> kube-controller-manager <br> kube-scheduler <br> etcd <br> coredns   <br>| 需要注意这些组件的部署方式 : <br> 部署在pod 中 <br> 以裸进程部署 |	`k8s-mon`默认认这些组件部署在pod中，通过**getpod**获取地址列表  |  
| k8s服务组件指标(每node部署) kube-proxy <br> kubelet-node | 需要注意这些组件的部署方式 : <br> 部署在pod 中 <br> 以裸进程部署 |	`k8s-mon`默认认这些组件在每个node都可以以`ip:port/metrics`访问到，通过**getnode**获取internal ip ，对应的服务需要listen 内网ip或`0.0.0.0`|  
| 业务指标(暂不支持)	| pod暴露的metrics接口|	- | -  |


# 使用指南
# 3、安装步骤

## setup01 准备工作

> 准备k8s环境 ，确保每个node节点部署夜莺agent `n9e-agent`

```shell script
# 创建namespace kube-admin
kubectl create ns kube-admin
# 创建访问etcd所需secret，在master上执行（不采集etcd则不需要）
# 注意如果 不采集etcd，没有创建对应的证书(如k8s使用公有云托管的)，默认 deployment中挂载证书那几行是注释掉的，开启etcd采集再打开
# etcd证书信息依据自己环境替换即可
kubectl create secret generic etcd-certs --from-file=/etc/kubernetes/pki/etcd/healthcheck-client.crt --from-file=/etc/kubernetes/pki/etcd/healthcheck-client.key --from-file=/etc/kubernetes/pki/etcd/ca.crt -n kube-admin

```

> 直接使用公共源的镜像
```shell script
# 公共源使用阿里云的
# registry.cn-beijing.aliyuncs.com/n9e/k8s-mon:v1
```


> 或者自己下载代码，打镜像

```shell script
mkdir -pv $GOPATH/github.com/n9e 
cd $GOPATH/github.com/n9e 
git clone https://github.com/n9e/k8s-mon 

# 使用docker 命令，或者ci工具,将镜像同步到仓库中 
# 如需修改镜像名字，需要同步修改daemonset 和deployment yaml文件中的image字段
# 镜像需要同步到所有node，最好上传到仓库中
cd k8s-mon  && docker build -t k8s-mon:v1 . 

```


## setup02 必须修改的配置
> 修改对接夜莺nid标签的名字
- 对应配置为配置文件中的`n9e_nid_label_name`
- 默认为:`N9E_NID`，与之前k8s-mon采集cadvisor指标要求容器环境变量名一致
- 如需修改则需要改 `k8s-config/configMap_deployment.yaml` 和 `k8s-config/configMap_daemonset.yaml` 中的 `n9e_nid_label_name`字段 

> pod yaml文件中传入上述 nid标签，例如：`N9E_NID`  

- 举例：deployment中定义pod的 `N9E_NID`  label，假设test-server01这个模块对应的服务树节点nid为5
- 后续该pod的容器的基础指标出现在nid=5的节点下: 如 cpu.user 
- 后续该pod的k8s的基础指标出现在nid=5的节点下: 如 kube_deployment_status_replicas_available
- 其余自定义标签不采集,如：`region: A` `cluster: B`
                     
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-server01-deployment
  labels:
    app: test-server01
    # 这里表示此deployment的nid为5
    N9E_NID: "5"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-server01
  template:
    metadata:
      labels:
        app: test-server01
        region: A
        cluster: B
        # 这里表示此deployment启动的容器nid为5
        N9E_NID: "5"
```

> 服务组件监控需要指定server_side_nid
-  修改  `k8s-config/configMap_deployment.yaml` 将 server_side_nid: 字段改为指定的服务组件监控叶子节点的nid
- 举例：server_side_nid: "6"：代表6为k8s集群的服务树叶子节点，k8s控制平面的指标都会上报到这里

> k8s服务组件指标(master侧) 如果不是部署在pod中，需要指定采集地址
- apiserver 、kube-scheduler、coredns、etcd等
- `k8s-mon`默认认这些组件部署在pod中，通过**getpod**获取地址列表
- 如果不是部署在pod中，需要指定采集地址(将user_specified设置为true，并指定addr，其余配置保持不变即可)，举例如下
```yaml
apiserver:
  user_specified: true
  addrs:
    - "https://1.1.1.1:6443/metrics"
    - "https://2.2.2.2:6443/metrics"
```


## setup03 可以调整的配置(维持默认值时可跳过此段配置)
> 如果不想采集某类指标可以去掉其配置
- 举例：不想采集`apiserver`的指标
- 则去掉/注释掉 `k8s-config/configMap_deployment.yaml`中 `apiserver`段即可
- deployment中需要采集每node的`kube-proxy` 和`kubelet` (node量大的时候)不需要可以去掉

> 每node的`kube-proxy` 和`kubelet`静态分片采集
- 默认采集所有node的指标，在node数量大时会导致性能问题，则需要开启分片采集
- 举例有1万个node需要采集kube-proxy，则部署3个k8s-mon，配置值开启kube-proxy段
- 其中`hash_mod_num`代表总分片数量 `hash_mod_shard`代表本实例取模后的index(取值范围是0 ~ hash_mod_num-1)
- 那么这三个实例则会将1万个node分片采集

```yaml
# 实例1
kube_proxy:
  hash_mod_num: 3
  hash_mod_shard: 0
```

```yaml
# 实例2
kube_proxy:
  hash_mod_num: 3
  hash_mod_shard: 1
```

```yaml
# 实例3
kube_proxy:
  hash_mod_num: 3
  hash_mod_shard: 2
```


> 想给某个采集项指定采集地址
- 举例：想设置kube-scheduler的采集地址为 `https://1.1.1.1:1234/metrics` 和 `https://2.2.2.2:1234/metrics` 
- 则修改`k8s-config/configMap_deployment.yaml`中 `user_specified` 和`addrs`即可
```yaml
kube_scheduler:
  user_specified: true
  addrs:
    - "https://1.1.1.1:1234/metrics"
    - "https://2.2.2.2:1234/metrics"
```



> 如需给采集的指标添加自定义tag 
- 则修改 `k8s-config/configMap_deployment.yaml` `k8s-config/configMap_daemonset.yaml`中的`append_tags`字段即可
```yaml
append_tags:
  key1: value1
  key2: value2
```

> 如需修改采集间隔
- 修改`k8s-config/configMap_deployment.yaml` `k8s-config/configMap_daemonset.yaml`中的 `collect_step` 字段

> 如需修改某个项目的采集并发
- 修改`k8s-config/configMap_deployment.yaml` 中的指定项目的 `concurrency_limit` 字段，默认10


> 如需服务组件采集多实例时的特征标签
- 修改`k8s-config/configMap_deployment.yaml` 中的 `multi_server_instance_unique_label` 字段

> 调整日志级别
- 修改`k8s-config/deployment.yaml` 中的 spec.containers.command 加上`--log.level=debug`即可看到debug日志，日志样例如下
- 单项数据处理耗时
```shell script
level=debug ts=2021-02-24T15:47:31.810+08:00 caller=kube_controller_manager.go:180 msg=DoCollectSuccessfullyReadyToPush funcName=kube-controller-manager metrics_num=621 time_took_seconds=0.307592776
```
- 单项推送耗时
```shell script
level=debug ts=2021-02-24T15:47:31.863+08:00 caller=push.go:25 msg=PushWorkSuccess funcName=kube-controller-manager url=http://localhost:2080/api/collector/push metricsNum=621 time_took_seconds=0.053355322
```
- 获取pod耗时
```shell script
level=debug ts=2021-02-24T15:50:01.523+08:00 caller=get_pod.go:99 msg=server_pod_ips_result num_kubeSchedulerIps=1 num_kubeControllerIps=1 num_apiServerIps=1 num_coreDnsIps=2 num_kubeProxyIps=2 num_etcdIps=1 time_took_seconds=0.020384107
``` 

## setup04 启动服务 

> 启动ksm服务(部署在kube-system namespace中 ，需要采集才启动)
```shell script
kubectl apply -f k8s-config/kube-stats-metrics
```
> 启动k8s-mon daemonset 和deployment (部署在kube-admin namespace中，按需启动daemonset 和deployment)
```shell script
kubectl apply -f k8s-config
```

## setup05 观察日志，查看指标
> 查看日志 
```shell script
kubectl logs -l app=k8s-mon-deployment  -n kube-admin  -f
kubectl logs -l app=k8s-mon-daemonset  -n kube-admin  -f
``` 

## setup06 查看指标，导入大盘图
> 即时看图查看指标 
```shell script
# 浏览器访问及时看图path： http://<n9e_addr>/mon/dashboard?nid=<nid>
 
``` 
> 导入大盘图 
```shell script
# 大盘图在 metrics-detail/夜莺大盘-xxxjson中
# 将三个大盘json文件放到夜莺服务端机器 <n9e_home>/etc/screen 下 
# 或者 克隆夜莺3.5+代码，内置大盘图json在 etc/screen 下 
# 刷新页面，在对应的节点选择导入内置大盘即可
``` 


## 注意事项
### 采集间隔
- kubelet 内置了cadvisor作为容器采集 ，具体文档可以看这里 [cadvisor housekeeping配置](https://github.com/google/cadvisor/blob/master/docs/runtime_options.md#housekeeping)
- 同时kubelet 命令行透传了[相关配置](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/)
- ```--housekeeping-interval duration     Default: `10s` ``` 模式采集是10秒，所以在`默认配置`下无论prometheus还是k8s-mon采集间隔不应低于10s
- **cpu 和mem指标需要pod设置limit，如果没有limit则某些指标会缺失**

### 白名单问题
- 建议保持默认的metrics名单，metrics_white_list为空则全采集
- tag白名单可按需配置

### histogram数据问题
- 提供基于 histogram的分位值，所有的_bucket指标已经被过滤掉了，提供分位值`quantile`  50 90 95 99
- 线性插值法计算，和prometheus大致相同 :prometheus先算rate再算sum，k8s-mon是先算sum再算rate
- 举例 `coredns_dns_request_duration_seconds_bucket --> coredns_dns_request_duration_seconds_quantile ` 代表coredns 解析平均延迟分位值
- 同时提供平均值 举例 `coredns_dns_request_duration_seconds_bucket -->coredns_dns_request_duration_seconds_avg `


# 原有cadvisor采集模式，即配置文件中collect_mode : cadvisor_plugin
> 作为Nightingale的插件，用于收集docker容器的监控指标

## 快速构建 
```
    $ mkdir -p $GOPATH/src/github.com/n9e
    $ cd $GOPATH/src/github.com/n9e
    $ git clone https://github.com/n9e/k8s-mon.git
    $ cd k8s-mon
    $ make
    $ ./k8s-mon
```

## 前置依赖
1. docker容器所在宿主机已安装并启动了cadvisor
2. docker容器的环境变量中包含 N9E_NID ，N9E_NID 的内容为夜莺服务树节点id，如果设置 N9E_NID = 1，则到节点id为1的节点下，就可以容器的监控指标

## 使用方式
1. 将 k8s-mon、k8s-mon.yml 分发到容器所在的宿主机上
2. 到宿主机所属节点配置插件采集

![k8s-mon](https://s3-gz01.didistatic.com/n9e-pub/image/docker.png)

3. 配置完之后，到即时看图，选择对应的节点，再选择设备无关，即可查看采集到的容器监控指标
![docker-metric](https://s3-gz01.didistatic.com/n9e-pub/image/docker_metric.png)

## 视频教程
[观看地址](https://s3-gz01.didistatic.com/n9e-pub/video/n9e-docker-mon.mp4)

## 指标列表

- CPU    
cpu.user    
cpu.sys    
cpu.idle    
cpu.util    
cpu.periods    
cpu.throttled_periods    
cpu.throttled_time    

- 内存    
mem.bytes.total    
mem.bytes.used    
mem.bytes.used.percent    
mem.bytes.cached    
mem.bytes.rss    
mem.bytes.swap    

- 磁盘    
disk.io.read.bytes    
disk.io.write.bytes	    
disk.bytes.total    
disk.bytes.used    
disk.bytes.used.percent    

- 网络    
net.sockets.tcp.timewait    
net.in.bits    
net.in.pps    
net.in.errs    
net.in.dropped    
net.out.bits    
net.out.pps    
net.out.errs    
net.out.dropped    
net.tcp.established    

- 系统    
sys.ps.process.used    
sys.ps.thread.used    
sys.fd.count.used    
sys.socket.count.used    
sys.restart.count    




