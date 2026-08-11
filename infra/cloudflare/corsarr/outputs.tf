output "custom_domain_id" {
  description = "Custom Domain identifier when the explicit gate is enabled."
  value       = try(cloudflare_workers_custom_domain.corsarr[0].id, null)
}

output "canonical_hostname" {
  description = "Stable browser origin for the Corsarr website."
  value       = var.hostname
}
