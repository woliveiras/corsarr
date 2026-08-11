provider "cloudflare" {}

resource "cloudflare_workers_custom_domain" "corsarr" {
  count = var.enable_custom_domain ? 1 : 0

  account_id = var.account_id
  zone_id    = var.zone_id
  hostname   = var.hostname
  service    = var.worker_name

  lifecycle {
    prevent_destroy = true
  }
}
