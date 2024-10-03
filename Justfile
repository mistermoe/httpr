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