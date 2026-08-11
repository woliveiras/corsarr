terraform {
  required_version = "= 1.13.3"

  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "= 5.22.0"
    }
  }

  backend "s3" {}
}
