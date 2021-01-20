## cadvisor指标说明
### **cpu 和mem指标需要pod设置limit，如果没有limit则某些指标会缺失**

### cpu指标
|  夜莺指标名   |含义| prometheus metrics或计算方式|说明 |
|  ----  | ----  | ---- | --- |
| cpu.util| 容器cpu使用占其申请的百分比 | sum (rate (container_cpu_usage_seconds_total[1m])) by( container) /( sum (container_spec_cpu_quota) by(container) /100000) * 100| 0-100的范围|
| cpu.idle| 容器cpu空闲占其申请的百分比 |100 - cpu.util  | 0-100的范围|
| cpu.user| 容器cpu用户态使用占其申请的百分比 | sum (rate (container_cpu_user_seconds_total[1m])) by( container) /( sum (container_spec_cpu_quota) by(container) /100000) * 100| 0-100的范围|
| cpu.sys| 容器cpu内核态使用占其申请的百分比 | sum (rate (container_cpu_sys_seconds_total[1m])) by( container) /( sum (container_spec_cpu_quota) by(container) /100000) * 100| 0-100的范围|
| cpu.cores.occupy | 容器cpu使用占用机器几个核 | rate(container_cpu_usage_seconds_total[1m]) | 0到机器核数上限,结果为1就是占用1个核 |
| cpu.spec.quota | 容器的CPU配额 | container_spec_cpu_quota | 为容器指定的CPU个数*100000 |
| cpu.throttled.util | 容器CPU执行周期受到限制的百分比 | sum by(container_name, pod_name, namespace) (increase(container_cpu_cfs_throttled_periods_total{container_name!=""}[5m])) / <br>sum by(container_name, pod_name, namespace) (increase(container_cpu_cfs_periods_total[5m])) * 100 |  0-100的范围|
| cpu.periods | 容器生命周期中度过的cpu周期总数 | counter型无需计算 |  使用rate/increase 查看|
| cpu.throttled.periods | 容器生命周期中度过的受限的cpu周期总数 | counter型无需计算 |  使用rate/increase 查看|
| cpu.throttled.time | 容器被节流的总时间 )| counter型无需计算 |  单位(纳秒|


### mem指标

|  夜莺指标名   |含义| prometheus metrics或计算方式|说明 |
|  ----  | ----  | ---- | --- |
| mem.bytes.total | 容器的内存限制 | 无需计算 | 单位byte 对应pod yaml中resources.limits.memory|
| mem.bytes.used | 当前内存使用情况，包括所有内存，无论何时访问 | container_memory_rss + container_memory_cache + kernel memory | 单位byte |
| mem.bytes.used.percent | 容器内存使用率 | container_memory_usage_bytes/container_spec_memory_limit_bytes *100 | 范围0-100|
| mem.bytes.workingset| 容器真实使用的内存量，也是limit限制时的 oom 判断依据  | container_memory_max_usage_bytes > container_memory_usage_bytes >= container_memory_working_set_bytes > container_memory_rss |单位byte|
| mem.bytes.workingset.percent| 容器真实使用的内存量百分比  | container_memory_working_set_bytes/container_spec_memory_limit_bytes *100 | 范围0-100|
| mem.bytes.cached| 容器cache内存量| container_memory_cache | 单位byte |
| mem.bytes.rss| 容器rss内存量| container_memory_rss | 单位byte |
| mem.bytes.swap| 容器cache内存量| container_memory_swap | 单位byte |


### filesystem && disk.io指标
 
|  夜莺指标名   |含义| prometheus metrics或计算方式|说明 |
|  ----  | ----  | ---- | --- |
| disk.bytes.total | 容器可以使用的文件系统总量| container_fs_limit_bytes | (单位：字节) |
| disk.bytes.used | 容器已经使用的文件系统总量| container_fs_usage_bytes | (单位：字节) |
| disk.bytes.used.percent | 容器文件系统使用百分比| container_fs_usage_bytes/container_fs_limit_bytes *100 | 范围0-100 |
| disk.io.read.bytes | 容器io.read qps| rate(container_fs_reads_bytes_total)[1m] | (单位：bps) |
| disk.io.write.bytes | 容器io.write qps| rate(container_fs_write_bytes_total)[1m] | (单位：bps) |

### network指标
#### **网卡指标都应该求所有interface的和计算**
 
|  夜莺指标名   |含义| prometheus metrics或计算方式|说明 |
|  ----  | ----  | ---- | --- |
| net.in.bytes | 容器网络接收数据总数| rate(container_network_receive_bytes_total)[1m] | (单位：bytes/s) |
| net.out.bytes | 容器网络积传输数据总数）| rate(container_network_transmit_bytes_total)[1m] | (单位：bytes/s) |
| net.in.pps | 容器网络接收数据包pps| rate(container_network_receive_packets_total)[1m] | (单位：p/s) |
| net.out.pps | 容器网络发送数据包pps| rate(container_network_transmit_packets_total)[1m] | (单位：p/s) |
| net.in.errs | 容器网络接收数据错误数| rate(container_network_receive_errors_total)[1m] | (单位：bytes/s)|
| net.out.errs | 容器网络发送数据错误数| rate(container_network_transmit_errors_total)[1m] |(单位：bytes/s)|
| net.in.dropped | 容器网络接收数据包drop pps| rate(container_network_receive_packets_dropped_total)[1m] |  (单位：p/s)  |
| net.out.dropped | 容器网络发送数据包drop pps| rate(container_network_transmit_packets_dropped_total)[1m] |  (单位：p/s)  |
| container_network_{tcp,udp}_usage_total 默认不采集是因为 --disable_metrics=tcp, udp ,因为开启cpu压力大|



### system指标
 
|  夜莺指标名   |含义| prometheus metrics或计算方式|说明 |
|  ----  | ----  | ---- | --- |
| sys.ps.process.count | 容器中running进程个数| container_processes | (单位：个) |
| sys.ps.thread.count | 容器中进程running线程个数| container_threads | (单位：个) |
| sys.fd.count.used  | 容器中打开文件描述符个数| container_file_descriptors | (单位：个) |
| sys.fd.soft.ulimits  | 容器中root process Soft ulimit| container_ulimits_soft | (单位：个) |
| sys.socket.count.used  | 容器中打开套接字个数| container_sockets | (单位：个)|
| sys.task.state  | 容器中task 状态分布| container_tasks_state | (单位：个)|
