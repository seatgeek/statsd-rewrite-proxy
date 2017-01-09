Will rewrite Nomad and Fabio StatsD metrics into Datadog StatsD metrics, with tags, which make the metrics infinitely more useful.

Rules are defined in `rules.go`, and should be pretty self-explanitory.

This product is used in production at Bownty, and seem to work reasonable well.

Pull-Requests for other open source project rules are more than welcome.

A sample `nomad` job file exist in `_infrastrcture/nomad/` - the file is a template, and can't be run directly, please replace the `{{ }}` markers with actual values for your environment.

The agent will listen on UDP port `8126` for statsd, and TCP port `4000` for expvar export data. It will always forward metrics to `127.0.0.1:8125` (Datadog statsd default port)