locals {
  cloud_run_hostname = try(replace(google_cloud_run_v2_service.finance.uri, "https://", ""), "placeholder.run.app")
}

data "cloudflare_zones" "domain" {
  filter {
    name = var.domain
  }
}

locals {
  zone_id = data.cloudflare_zones.domain.zones[0].id
}

moved {
  from = cloudflare_record.root
  to   = cloudflare_record.app
}

resource "cloudflare_record" "app" {
  zone_id = local.zone_id
  name    = "app"
  content = local.cloud_run_hostname
  type    = "CNAME"
  proxied = true
  ttl     = 1
}

resource "cloudflare_workers_script" "origin_rewrite" {
  account_id = var.cloudflare_account_id
  name       = "finance-origin-rewrite"
  content    = <<-JS
    export default {
      async fetch(request) {
        const url = new URL(request.url);
        url.hostname = "${local.cloud_run_hostname}";
        return fetch(new Request(url, request));
      }
    }
  JS
  module = true
}

resource "cloudflare_workers_route" "root" {
  zone_id     = local.zone_id
  pattern     = "app.${var.domain}/*"
  script_name = cloudflare_workers_script.origin_rewrite.name
}

# Landing page (Cloudflare Pages)

resource "cloudflare_pages_project" "landing" {
  account_id        = var.cloudflare_account_id
  name              = "dnafami-landing"
  production_branch = "main"
}

resource "cloudflare_record" "root" {
  zone_id = local.zone_id
  name    = "@"
  content = "dnafami-landing.pages.dev"
  type    = "CNAME"
  proxied = true
  ttl     = 1
}

resource "cloudflare_record" "www" {
  zone_id = local.zone_id
  name    = "www"
  content = "dnafami-landing.pages.dev"
  type    = "CNAME"
  proxied = true
  ttl     = 1
}

resource "cloudflare_pages_domain" "root" {
  account_id   = var.cloudflare_account_id
  project_name = cloudflare_pages_project.landing.name
  domain       = var.domain
}

resource "cloudflare_pages_domain" "www" {
  account_id   = var.cloudflare_account_id
  project_name = cloudflare_pages_project.landing.name
  domain       = "www.${var.domain}"
}

