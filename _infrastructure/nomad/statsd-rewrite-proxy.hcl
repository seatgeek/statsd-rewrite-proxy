job "{{PROJECT_NAME}}" {
  region      = "global"
  type        = "system"
  datacenters = ["production", "vagrant"]

  task "server" {
    driver = "raw_exec"
    user   = "root"

    config {
      command = "statsd-rewrite-proxy"
    }

    artifact {
      source = "https://storage.googleapis.com/bownty-deploy-artifacts/{{PROJECT_NAME}}/{{APP_ENV}}/{{APP_VERSION}}/statsd-rewrite-proxy"
    }

    service {
      name = "dd-go-expvar"
      port = "http"

      check {
        type     = "tcp"
        port     = "http"
        interval = "10s"
        timeout  = "2s"
      }
    }

    resources {
      cpu    = 512
      memory = 256

      network {
        mbits = 1
        port  "http" {}
        port "statsd" {
          static = 8126
        }
      }
    }
  }
}
