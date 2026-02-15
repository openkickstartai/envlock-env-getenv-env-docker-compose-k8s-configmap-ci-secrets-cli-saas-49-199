# EnvLock 🔒

**Stop deploying with missing environment variables.** EnvLock scans your code for every `getenv` call, cross-references `.env` / `docker-compose.yml` / K8s ConfigMaps, and catches missing vars and config drift before production crashes.

Single static binary. Zero runtime dependencies. Works with Python, Node.js, Go, Ruby, Java.

## 🚀 Quick Start

```bash
# Install
go install github.com/envlock/envlock@latest

# Scan current directory (auto-detects .env and docker-compose.yml)
envlock --dir .

# Explicit sources
envlock --dir ./src --env .env.production --compose docker-compose.yml --k8s deploy/configmap.yaml

# JSON output for CI pipelines
envlock --dir . --json
```

### GitHub Actions

```yaml
- name: EnvLock Check
  run: envlock --dir ./src --env .env.example --json
```

## 📊 Why Pay for EnvLock?

Every team has had a deploy fail because someone forgot to add `STRIPE_SECRET_KEY` to production. Average cost of one env-var outage: **$2,000-$50,000** (downtime + engineer hours + customer impact). EnvLock pays for itself after preventing **one** incident.

## 💰 Pricing

| Feature | Free (CLI) | Pro ($49/mo) | Enterprise ($199/mo) |
|---|---|---|---|
| Code scanning (all languages) | ✅ | ✅ | ✅ |
| `.env` file parsing | ✅ | ✅ | ✅ |
| `docker-compose.yml` parsing | ✅ | ✅ | ✅ |
| K8s ConfigMap parsing | ✅ | ✅ | ✅ |
| Drift detection | ✅ | ✅ | ✅ |
| JSON output | ✅ | ✅ | ✅ |
| GitHub/GitLab PR comments | — | ✅ | ✅ |
| Slack/Teams alerts | — | ✅ | ✅ |
| Multi-repo scanning | — | ✅ | ✅ |
| Terraform/Vault/AWS SSM sources | — | ✅ | ✅ |
| Type & schema validation | — | ✅ | ✅ |
| CI secret store cross-check | — | — | ✅ |
| SOC2 audit trail & reports | — | — | ✅ |
| SSO + team management | — | — | ✅ |
| SLA & priority support | — | — | ✅ |

## How It Works

1. **Scan** — walks your codebase, extracts env var names via regex patterns
2. **Parse** — reads `.env`, `docker-compose.yml`, K8s ConfigMap YAML
3. **Compare** — finds vars referenced in code but missing from sources
4. **Drift** — detects vars that exist in some sources but not others
5. **Report** — human-readable or JSON output, non-zero exit for CI

## License

BSL 1.1 — Free for teams ≤5. Commercial license required for larger teams.
