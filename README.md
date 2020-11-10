# CodeCoach.Us API

It's a Done Deal.

# Development

```bash
# Copy base .env and setup credentials (secrets).
cp .env .env.local

# Do the same thing for testing.
# Remember to change the database name!
cp .env .env.test.local

# Start up server dependencies:
docker-compose -f deployments/docker-compose-deps.yml up -d

# Ensure everything works:
./test.sh
```

# Local Docker

```bash
# Copy base .env and setup credentials (secrets).
cp .env .env.docker.local

# Build:
docker-compose -f deployments/docker-compose-full.yml build

# Run:
docker-compose -f deployments/docker-compose-full.yml up -d

# Run end-to-end tests:
go run cmd/remote-e2e-test/main.go

# Run Github OAuth tests:
scripts/oauth-e2e.sh
```
