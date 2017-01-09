package main

func createRules() {
	// fabio rewrites
	rules = append(rules, NewRule("fabio.http.status.{code}", "x.fabio.http.response_code"))
	rules = append(rules, NewRule("fabio.{fabio_service}.{fabio_host}.{fabio_path}.{fabio_upstream}.{fabio_dimension}", "x.fabio.service.requests.{fabio_dimension}"))
	rules = append(rules, NewRule("fabio.{fabio_service}.{fabio_host}.{fabio_path}", "x.fabio.service.requests"))

	// nomad rewrites
	rules = append(rules, NewRule("nomad.client.uptime.*", "x.nomad.client.uptime"))

	rules = append(rules, NewRule("nomad.client.host.memory.{client_id}.{nomad_metric}", "x.nomad.client.memmory.{nomad_metric}"))
	rules = append(rules, NewRule("nomad.client.host.cpu.{client_id}.{nomad_cpu_core}.{nomad_metric}", "x.nomad.client.cpu.{nomad_metric}"))
	rules = append(rules, NewRule("nomad.client.host.disk.{client_id}.{nomad_device}.{nomad_metric}", "x.nomad.client.disk.{nomad_metric}"))

	rules = append(rules, NewRule("nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.memory.{nomad_metric}", "x.nomad.allocation.memory.{nomad_metric}"))
	rules = append(rules, NewRule("nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.cpu.{nomad_metric}", "x.nomad.allocation.cpu.{nomad_metric}"))
}
