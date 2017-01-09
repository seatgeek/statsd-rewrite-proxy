package main

func createRules() {
	// fabio rewrites
	rules = append(rules, NewRule("fabio.http.status.{code}", "fabio.http.response_code"))
	rules = append(rules, NewRule("fabio.{fabio_service}.{fabio_host}.{fabio_path}.{fabio_upstream}.{fabio_dimension}", "fabio.service.requests.{fabio_dimension}"))
	rules = append(rules, NewRule("fabio.{fabio_service}.{fabio_host}.{fabio_path}", "fabio.service.requests"))

	// nomad rewrites
	rules = append(rules, NewRule("nomad.client.uptime.*", "nomad.client.uptime"))

	rules = append(rules, NewRule("nomad.client.host.memory.{client_id}.{nomad_metric}", "nomad.client.memmory.{nomad_metric}"))
	rules = append(rules, NewRule("nomad.client.host.cpu.{client_id}.{nomad_cpu_core}.{nomad_metric}", "nomad.client.cpu.{nomad_metric}"))
	rules = append(rules, NewRule("nomad.client.host.disk.{client_id}.{nomad_device}.{nomad_metric}", "nomad.client.disk.{nomad_metric}"))

	rules = append(rules, NewRule("nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.memory.{nomad_metric}", "nomad.allocation.memory.{nomad_metric}"))
	rules = append(rules, NewRule("nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.cpu.{nomad_metric}", "nomad.allocation.cpu.{nomad_metric}"))
}
