# 模式说明


|  模式名称  |  部署运行方式 |  collect_mode配置   |  说明 | 
|  ----  | ----  | ---- |---| 
| 夜莺插件形式采集cadvisor raw api	| 可执行的插件由夜莺agent调用|	cadvisor_plugin | 文档在readme最下面 (原有cadvisor采集模式) |  
| 容器基础资源指标采集	|  k8s daemonset 部署在每一个node上|	kubelet_agent | 统称为新模式   |  
| 集中采集k8s服务组件	|k8s deployment 部署 |	server_side | 统称为新模式  |  


# 新模式：部署在k8s中采集相关指标
## 原理说明

- 通过抓取各个组件的/metrics接口获得prometheus的数据，ETL后push到夜莺
- 各个采集项目有metirc和tag的白名单过滤
- `cadvisor`数据需要hold点做计算比率类指标，多用在百分比的情况，其余不需要
- `counter`类型将有夜莺agent转换为`gauge`型，即数值已经转为`rate`了
- 指标说明在`metrcs-detail`文件夹里
- k8s yaml配置在 `k8s-config`中
- 服务组件监控时多实例问题：用户无需关心，k8s-mon自动发现并采集 


## 采集内容说明

- 一般来说在k8s集群汇总我们关注一下4类指标 

|  指标类型  |  采集源 |  应用举例  |  部署方式 | 
|  ----  | ----  | ---- |---| 
| 容器基础资源指标	| kubelet 内置cadvisor |	查看容器cpu、mem等 | k8s daemonset   |  
| k8s资源指标	| [kube-stats-metrics](https://github.com/kubernetes/kube-state-metrics) (简称ksm)|	查看pod状态、查看deployment信息等 | k8s deployment (需要提前部署ksm)   |  
| k8s服务组件指标	|各个服务组件的metrics接口(多实例自动发现)<br> apiserver <br> kube-controller-manager <br> kube-scheduler <br> etcd <br> coredns  |	查看请求延迟/QPS等 | 和ksm同一套代码，部署在  k8s deployment   |  
| 业务指标(暂不支持)	| pod暴露的metrics接口|	- | -  |  

# 使用指南
# 3、安装步骤

## setup01 准备工作

> 准备k8s环境 ，确保每个node节点部署夜莺agent `n9e-agent`

```shell script
# 创建namespace kube-admin
kubectl create ns kube-admin
```

> 下载代码，打镜像

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

## setup03 可以调整的配置

> 如需给采集的指标添加自定义tag 
- 则修改 `k8s-config/configMap_deployment.yaml` `k8s-config/configMap_daemonset.yaml`中的`append_tags`字段即可
```yaml
append_tags:
  key1: value1
  key2: value2
```

> 如需修改采集间隔
- 修改`k8s-config/configMap_deployment.yaml` `k8s-config/configMap_daemonset.yaml`中的 `collect_step` 字段

> 如需服务组件采集多实例时的特征标签
- 修改`k8s-config/configMap_deployment.yaml` 中的 `multi_server_instance_unique_label` 字段


## setup04 启动服务 

> 启动ksm服务(部署在kube-system namespace中)
```shell script
kubectl apply -f k8s-config/kube-stats-metrics
```
> 启动k8s-mon daemonset 和deployment (部署在kube-admin namespace中)
```shell script
kubectl apply -f k8s-config
```

## setup04 观察日志，查看指标
> 查看日志 
```shell script
kubectl logs -l app=k8s-mon-deployment  -n kube-admin  -f
kubectl logs -l app=k8s-mon-daemonset  -n kube-admin  -f
``` 

## setup04 查看指标，导入大盘图
> 即时看图查看指标 
```shell script
# 浏览器访问及时看图path： http://<n9e_addr>/mon/dashboard?nid=<nid>
 
``` 
> 导入大盘图 
```shell script
# 大盘图在 metrics-detail/夜莺大盘-xxxjson中
``` 


## 注意事项
### 采集间隔
- kubelet 内置了cadvisor作为容器采集 ，具体文档可以看这里 [cadvisor housekeeping配置](https://github.com/google/cadvisor/blob/master/docs/runtime_options.md#housekeeping)
- 同时kubelet 命令行透传了[相关配置](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/)
- ```--housekeeping-interval duration     Default: `10s` ``` 模式采集是10秒，所以在`默认配置`下无论prometheus还是k8s-mon采集间隔不应低于10s

### 白名单问题
- 建议保持默认的metrics名单，metrics_white_list为空则全采集
- tag白名单可按需配置

### histogram数据问题
- 暂时不提供基于 histogram的分位值，所以所有的_bucket指标已经被过滤掉了
- 取而代之的是提供平均值，即： xx_sum/xx_count 举例 ` etcd请求平均延迟 = etcd_request_duration_seconds_sum /etcd_request_duration_seconds_count `



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




