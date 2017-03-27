package main

func createRules() {

	/*********************************************************************************************************************************************************
	 * Nomad Key Metrics
	 *********************************************************************************************************************************************************/

	// nomad.runtime.num_goroutines
	rules = append(rules, NewRule("nomad.runtime.num_goroutines", "nomad.runtime.num_goroutines"))

	// nomad.runtime.alloc_bytes
	rules = append(rules, NewRule("nomad.runtime.alloc_bytes", "nomad.runtime.alloc_bytes"))

	// nomad.runtime.sys_bytes
	rules = append(rules, NewRule("nomad.runtime.sys_bytes", "nomad.runtime.sys_bytes"))

	// nomad.runtime.malloc_count
	rules = append(rules, NewRule("nomad.runtime.malloc_count", "nomad.runtime.malloc_count"))

	// nomad.runtime.free_count
	rules = append(rules, NewRule("nomad.runtime.free_count", "nomad.runtime.free_count"))

	// nomad.runtime.heap_objects
	rules = append(rules, NewRule("nomad.runtime.heap_objects", "nomad.runtime.heap_objects"))

	// nomad.runtime.total_gc_pause_ns
	rules = append(rules, NewRule("nomad.runtime.total_gc_pause_ns", "nomad.runtime.total_gc_pause_ns"))

	// nomad.runtime.total_gc_runs
	rules = append(rules, NewRule("nomad.runtime.total_gc_runs", "nomad.runtime.total_gc_runs"))

	// nomad.uptime
	rules = append(rules, NewRule("nomad.uptime", "nomad.uptime"))

	// nomad.client.uptime.<HostID>
	rules = append(rules, NewRule("nomad.client.uptime.{nomad_client}", "nomad.client.uptime"))

	// nomad.worker.invoke_scheduler.<type>
	rules = append(rules, NewRule("nomad.worker.invoke_scheduler.{nomad_scheduler}", "nomad.worker.invoke_scheduler"))

	/*********************************************************************************************************************************************************
	 * Nomad Host Metrics
	 *********************************************************************************************************************************************************/

	// nomad.client.host.cpu.<HostID>.<CPU-Core>.total
	// nomad.client.host.cpu.<HostID>.<CPU-Core>.user
	// nomad.client.host.cpu.<HostID>.<CPU-Core>.system
	// nomad.client.host.cpu.<HostID>.<CPU-Core>.idle
	rules = append(rules, NewRule("nomad.client.host.cpu.{nomad_client}.{nomad_client_cpu_core}.{nomad_cpu_metric}", "nomad.client.cpu.{nomad_cpu_metric}"))

	// nomad.client.host.disk.<HostID>.<Device-Name>.size
	// nomad.client.host.disk.<HostID>.<Device-Name>.used
	// nomad.client.host.disk.<HostID>.<Device-Name>.available
	// nomad.client.host.disk.<HostID>.<Device-Name>.used_percent
	// nomad.client.host.disk.<HostID>.<Device-Name>.inodes_percent
	rules = append(rules, NewRule("nomad.client.host.disk.{nomad_client}.{nomad_client_device}.{nomad_disk_metric}", "nomad.client.disk.{nomad_disk_metric}"))

	// nomad.client.host.memory.<HostID>.total
	// nomad.client.host.memory.<HostID>.available
	// nomad.client.host.memory.<HostID>.used
	// nomad.client.host.memory.<HostID>.free
	rules = append(rules, NewRule("nomad.client.host.memory.{nomad_client}.{nomad_client_memory_metric}", "nomad.client.host.memory.{nomad_client_memory_metric}"))

	// nomad.client.allocated.cpu.<HostID>
	rules = append(rules, NewRule("nomad.client.allocated.cpu.{nomad_client}", "nomad.client.allocated.cpu"))

	// nomad.client.allocated.memory.<HostID>
	rules = append(rules, NewRule("nomad.client.allocated.memory.{nomad_client}", "nomad.client.allocated.memory"))

	// nomad.client.allocated.disk.<HostID>
	rules = append(rules, NewRule("nomad.client.allocated.disk.{nomad_client}", "nomad.client.allocated.disk"))

	// nomad.client.allocated.iops.<HostID>
	rules = append(rules, NewRule("nomad.client.allocated.iops.{nomad_client}", "nomad.client.allocated.iops"))

	// nomad.client.allocated.network.<Device-Name>.<HostID>
	rules = append(rules, NewRule("nomad.client.allocated.network.{nomad_device_name}.{nomad_client}", "nomad.client.allocated.network"))

	// nomad.client.unallocated.cpu.<HostID>
	rules = append(rules, NewRule("nomad.client.unallocated.cpu.{nomad_client}", "nomad.client.unallocated.cpu"))

	// nomad.client.unallocated.memory.<HostID>
	rules = append(rules, NewRule("nomad.client.unallocated.memory.{nomad_client}", "nomad.client.unallocated.memory"))

	// nomad.client.unallocated.disk.<HostID>
	rules = append(rules, NewRule("nomad.client.unallocated.disk.{nomad_client}", "nomad.client.unallocated.disk"))

	// nomad.client.unallocated.iops.<HostID>
	rules = append(rules, NewRule("nomad.client.unallocated.iops.{nomad_client}", "nomad.client.unallocated.iops"))

	// nomad.client.unallocated.network.<Device-Name>.<HostID>
	rules = append(rules, NewRule("nomad.client.unallocated.network.{nomad_device_name}.{nomad_client}", "nomad.client.unallocated.network"))

	/*********************************************************************************************************************************************************
	 * Nomad Allocation Metrics
	 *********************************************************************************************************************************************************/

	// nomad.client.allocations.migrating.<HostID>
	rules = append(rules, NewRule("nomad.client.allocations.migrating.{nomad_client}", "nomad.client.allocations.migrating"))

	// nomad.client.allocations.blocked.<HostID>
	rules = append(rules, NewRule("nomad.client.allocations.blocked.{nomad_client}", "nomad.client.allocations.blocked"))

	// nomad.client.allocations.pending.<HostID>
	rules = append(rules, NewRule("nomad.client.allocations.pending.{nomad_client}", "nomad.client.allocations.pending"))

	// nomad.client.allocations.running.<HostID>
	rules = append(rules, NewRule("nomad.client.allocations.running.{nomad_client}", "nomad.client.allocations.running"))

	// nomad.client.allocations.terminal.<HostID>
	rules = append(rules, NewRule("nomad.client.allocations.terminal.{nomad_client}", "nomad.client.allocations.terminal"))

	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.rss
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.cache
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.swap
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.max_usage
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.kernel_usage
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.kernel_max_usage
	rules = append(rules, NewRule("nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.memory.{nomad_job_memory_metric}", "nomad.allocation.memory.{nomad_job_memory_metric}"))

	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.cpu.total_percent
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.cpu.system
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.cpu.user
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.cpu.throttled_time
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.cpu.total_ticks
	rules = append(rules, NewRule("nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.cpu.{nomad_job_cpu_metric}", "nomad.allocation.cpu.{nomad_job_cpu_metric}"))
}
