# Introduction

Will rewrite Nomad StatsD metrics into DataDog StatsD metrics, with tags, which make the metrics infinitely more useful.

Rules are defined in `rules.go`, and should be pretty self-explanitory.

Pull-Requests for other open source project rules are more than welcome.

A sample `nomad` job file exist in `_infrastrcture/nomad/` - the file is a template, and can't be run directly, please replace the `{{ }}` markers with actual values for your environment.

The agent will listen on UDP port `8126` for statsd, and TCP port `4000` for expvar export data. It will always forward metrics to `127.0.0.1:8125` (DataDog StatsD default port)

## Nomad

### Example

`nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.memory.{nomad_metric}`

can be rewritten into

`nomad.allocation.memory.{nomad_metric}`

with

`nomad_job,nomad_task_group,nomad_allocation_id,nomad_task,nomad_metric` being converted into named tags (e.g. `nomad_job:example-job`)

### Config

The following Nomad client telemetry configuration is assumed.

```hcl
telemetry {
  statsd_address = "localhost:8126"

  disable_hostname = true

  publish_node_metrics = true
  publish_allocation_metrics = true
}
```
