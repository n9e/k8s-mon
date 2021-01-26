## golang process
|  指标名   | 类型|含义  | 说明| 
|  ----  | ----  | ---- | ---- |
| process_resident_memory_bytes	| gauge|	mem rss大小 |   |  
| go_memstats_alloc_bytes	| gauge|	还在使用中的已分配byte |   |  
| go_memstats_stack_inuse_bytes	| gauge|	栈inuse |   |  
| go_memstats_heap_inuse_bytes	| gauge|	堆inuse |   |  
| process_open_fds	| gauge|	打开fd数 |   |  
| go_goroutines	| gauge|	goroutine数|     |  
| go_gc_duration_seconds	| summary|	gc耗时|     |  
| process_cpu_seconds_total	| counter|	cpu |     |  