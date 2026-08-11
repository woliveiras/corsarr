# Corsarr website

Static bilingual landing page for `corsarr.woliveiras.com`.

```sh
pnpm install
pnpm dev
pnpm quality
pnpm build
```

Portuguese is served from `/` and English from `/en/`. Product screenshots must
come from the real Corsarr Desktop application. Third-party names and logos are
used only to identify supported integrations and remain the property of their
respective projects.

Production delivery is triggered by a semantic release tag. Wrangler owns the
Worker release; Terraform owns only the Custom Domain.
