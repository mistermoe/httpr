_help:
  @just -l

lint:
  @echo "Running linter..."
  @golangci-lint run

test:
  @echo "Running tests..."
  @go clean -testcache && go test -cover ./...

docs:
  #!/bin/bash
  set -euo pipefail

  cd docs
  pnpm install
  pnpm start

observe:
  #!/bin/bash
  set -euo pipefail

  docker run \
    --name otel-lgtm \
    --platform linux/amd64 \
    -p 3000:3000 \
    -p 4317:4317 \
    -p 4318:4318 \
    -e ENABLE_LOGS_ALL=true \
    -e GF_AUTH_ANONYMOUS_ENABLED=true \
    -e GF_AUTH_ANONYMOUS_ORG_ROLE=Admin \
    -e GF_AUTH_DISABLE_LOGIN_FORM=true \
    grafana/otel-lgtm