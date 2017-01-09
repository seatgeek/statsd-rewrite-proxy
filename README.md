# Introduction

Will rewrite Nomad and Fabio StatsD metrics into Datadog StatsD metrics, with tags, which make the metrics infinitely more useful.

Rules are defined in `rules.go`, and should be pretty self-explanitory.

This product is used in production at Bownty, and seem to work reasonable well.

Pull-Requests for other open source project rules are more than welcome.

A sample `nomad` job file exist in `_infrastrcture/nomad/` - the file is a template, and can't be run directly, please replace the `{{ }}` markers with actual values for your environment.

The agent will listen on UDP port `8126` for statsd, and TCP port `4000` for expvar export data. It will always forward metrics to `127.0.0.1:8125` (Datadog statsd default port)

## Fabio

### Example

`fabio.{fabio_service}.{fabio_host}.{fabio_path}.{fabio_upstream}.{fabio_dimension}`

can be rewritten into

`fabio.service.requests.{fabio_dimension}`

with

`fabio_service,fabio_host,fabio_path,fabio_upstream,fabio_dimension` being converted into named tags (e.g. `fabio_host:example.com`)

### Config

The following Fabio env configuration is assumed, the `FABIO_METRICS_PREFIX` is especially important.

```bash
FABIO_METRICS_TARGET      = "statsd"
FABIO_METRICS_STATSD_ADDR = "127.0.0.1:8126"
FABIO_METRICS_PREFIX      = "fabio."
```

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