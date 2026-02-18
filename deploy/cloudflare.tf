locals {
  cloud_run_hostname = replace(google_cloud_run_v2_service.finance.uri, "https://", "")
}

data "cloudflare_zones" "domain" {
  filter {
    name = var.domain
  }
}

locals {
  zone_id = data.cloudflare_zones.domain.zones[0].id
}

resource "cloudflare_record" "root" {
  zone_id = local.zone_id
  name    = "@"
  content = local.cloud_run_hostname
  type    = "CNAME"
  proxied = true
  ttl     = 1
}

resource "cloudflare_record" "tenant" {
  for_each = toset(var.tenants)

  zone_id = local.zone_id
  name    = each.value
  content = local.cloud_run_hostname
  type    = "CNAME"
  proxied = true
  ttl     = 1
}

resource "cloudflare_ruleset" "origin_rules" {
  zone_id     = local.zone_id
  name        = "Origin Rules"
  description = "Rewrite Host header to Cloud Run hostname"
  kind        = "zone"
  phase       = "http_request_origin"

  rules {
    action = "route"
    action_parameters {
      host_header = local.cloud_run_hostname
    }
    expression  = "true"
    description = "Set Host header to Cloud Run hostname"
    enabled     = true
  }
}