# Corsarr Cloudflare infrastructure

The Corsarr repository owns only the `corsarr-site` Worker delivery configuration
and its `corsarr.woliveiras.com` Custom Domain. The shared `woliveiras.com` zone
remains externally managed and must not be imported into this state.

- Wrangler owns website assets, Worker versions, deployments, and `workers.dev`.
- Terraform owns only the long-lived Workers Custom Domain.
- The Terraform root uses its own R2 backend key.
- `enable_custom_domain` remains `false` until `corsarr-site` has been deployed.
- Every Terraform apply and first production publication require explicit approval.

The GitHub environments expected by the workflows are:

- `cloudflare-corsarr-plan`
- `cloudflare-corsarr-production`

Both Terraform environments need the R2 state variables and credentials documented
in the workflow. The production environment additionally owns the narrow Cloudflare
apply or Wrangler deployment token.
