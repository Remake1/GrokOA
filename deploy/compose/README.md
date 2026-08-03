# Docker Compose deployment

This deployment path runs the same three service shape as the Swarm stack, but
uses Docker Compose bridge networking and direct host port publishing:

- `nginx` as the public entrypoint
- `web` as the static Vue application
- `api` as the Go backend

Use this path for local deployment when Docker Swarm ingress is unreliable.

## Prerequisites

- Docker Engine
- Docker Compose
- An `api/.env` file created from `api/.env.example`

## One-command deploy

Run:

```bash
./deploy/compose/deploy.sh
```

The script will:

1. Read `api/.env`.
2. Write local Compose secret files under `deploy/compose/.secrets/`.
3. Build the `api` and `web` images locally.
4. Start the `nginx`, `web`, and `api` services with Docker Compose.

The generated `.secrets/` directory is ignored by git.

## Stop

Run:

```bash
docker compose --project-name crackoa-compose --file deploy/compose/compose.yml down
```

## Port

The default published port is `80`:

```text
http://127.0.0.1/
```

To publish a different host port:

```bash
PUBLISHED_PORT=8080 ./deploy/compose/deploy.sh
```

If the Swarm stack is still running on port `80`, stop it first or use a
different `PUBLISHED_PORT`.

## Secrets

The API reads these values from Compose secrets through `*_FILE` environment
variables:

- `ACCESS_KEY`
- `JWT_SECRET`
- `OPENAI_API_KEY`
- `GEMINI_API_KEY`

`ACCESS_KEY` and `JWT_SECRET` are required. The AI provider keys may be empty,
in which case the provider is disabled.

## Image tags

By default the deploy script builds:

- `crackoa/api:compose-local`
- `crackoa/web:compose-local`

Override them if needed:

```bash
IMAGE_PREFIX=registry.example.com/crackoa IMAGE_TAG=local ./deploy/compose/deploy.sh
```
