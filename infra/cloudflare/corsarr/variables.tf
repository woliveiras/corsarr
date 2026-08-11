variable "account_id" {
  description = "Cloudflare account that owns the Corsarr website Worker."
  type        = string
  nullable    = false

  validation {
    condition     = length(trimspace(var.account_id)) > 0
    error_message = "account_id must not be empty."
  }
}

variable "zone_id" {
  description = "Cloudflare identifier for the separately managed woliveiras.com zone."
  type        = string
  nullable    = false

  validation {
    condition     = length(trimspace(var.zone_id)) > 0
    error_message = "zone_id must not be empty."
  }
}

variable "worker_name" {
  description = "Wrangler-owned Worker service that receives the Custom Domain."
  type        = string
  default     = "corsarr-site"
  nullable    = false

  validation {
    condition     = var.worker_name == "corsarr-site"
    error_message = "This root is intentionally limited to the corsarr-site Worker."
  }
}

variable "hostname" {
  description = "Canonical hostname for the Corsarr website."
  type        = string
  default     = "corsarr.woliveiras.com"
  nullable    = false

  validation {
    condition     = var.hostname == "corsarr.woliveiras.com"
    error_message = "This root is intentionally limited to corsarr.woliveiras.com."
  }
}

variable "enable_custom_domain" {
  description = "Explicit gate for attaching the Custom Domain after corsarr-site exists."
  type        = bool
  default     = false
  nullable    = false
}
