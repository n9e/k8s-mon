## v2.0.5 / 2021-02-24
* [BUGFIX] curl请求采集接口时，http.resp.status_code 非200直接报错返回，避免权限错误引起的解析错误的误导
* [CHANGE] 多实例采集时，最终0结果改为不push
* [CHANGE] 把一些常规info日志降级成debug，--log.level=debug可以调整日志级别

## v2.0.4 / 2021-01-28
* [FEATURE] 新增ksm指标计算节点cpu/mem 请求/限制率等指标
* [BUGFIX] ksm启动不再sleep等待，因为push的瓶颈在transfer已经解决了



## v2.0.3 / 2021-01-27
* [ENHANCEMENT] 修改大盘文件，测试导入内还在功能
* [CHANGE] 完善readme
* [CHANGE] 新增changelog.md


## v2.0.2 / 2021-01-26
* [FEATURE] 新增服务组件histogram计算分位值
* [FEATURE] 新增计算avg 和成功率
* [FEATURE] 新增etcd采集
* [FEATURE] 新增golang进程指标采集
* [FEATURE] preaggregation.md 作为预聚合指标文档
* [ENHANCEMENT] 所有counter类型metric_name 加_rate后缀,加以区分
* [CHANGE] 完善readme和大盘


## v2.0.1 / 2021-01-20
* [BUGFIX] cpu 和mem指标需要pod设置limit，如果没有limit则某些指标会缺失
* [BUGFIX] daemonset默认设置limit
* [BUGFIX] GetPortListenAddr获取内网ip时没有及时close导致 fd会泄露
* [BUGFIX] GetPortListenAddr改为ticker运行前获取一次传入
* [BUGFIX] 修复/var/run/docker.sock泄露问题
* [FEATURE] 基础指标添加node_ip node_name作为宿主机标识标签
* [ENHANCEMENT] ksm指标没有nid的默认上报到服务节点server_side_nid ，例如kube_node_status_allocatable_cpu_cores这种共享指标
* [CHANGE] import package fmt
* [CHANGE] 修改readme


## v1.0 / 2021-01-19
* [FEATURE] 采集每node上的kube-proxy kubelet-node指标时支持并发数配置和静态分片
* [FEATURE] 服务组件每个项目支持用户指定地址来覆盖getpod，getnode获取不到地址的case
* [CHANGE] 每种项目配置了相关配置段才会开启，如果不想采集某类指标可以去掉其配置