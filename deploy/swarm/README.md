# Docker Swarm deployment

This deployment path runs three services in Docker Swarm:

- `nginx` as the public entrypoint on port `80`
- `web` as the static Vue application
- `api` as the Go backend

## Prerequisites

- Docker Engine with Swarm mode available
- An `api/.env` file created from `api/.env.example`

## One-command deploy

Run:

```bash
./deploy/swarm/deploy.sh
```

The script will:

1. Initialize Swarm if it is not active.
2. Build the `api` and `web` images locally.
3. Create Docker Swarm secrets from `api/.env`.
4. Deploy the stack with `docker stack deploy`.

For a single-node swarm, the default local image tags are enough.

After deployment, the script prints both the local URL and the LAN URL. Open the
LAN URL on a phone connected to the same network. For example:

```text
LAN URL:   http://192.168.0.110/
```

Use `http://`, not `localhost`: on a phone, `localhost` refers to the phone.

To publish a different host port:

```bash
PUBLISHED_PORT=8080 ./deploy/swarm/deploy.sh
```

## Checking the published port

Swarm publishes ports through its ingress routing mesh. The individual task
container may therefore show only `80/tcp` in `docker ps`; that does not mean the
host port is missing. Check the service instead:

```bash
docker stack services crackoa
docker service inspect crackoa_nginx --format '{{json .Endpoint.Ports}}'
```

The expected stack output includes `*:80->80/tcp` (or the port selected with
`PUBLISHED_PORT`).

If the LAN URL works on the host but not on another device, make sure both
devices are on the same non-guest Wi-Fi network and that client/AP isolation,
a VPN, or the host firewall is not blocking local traffic.

## Secrets

The API reads these values from Docker secrets through `*_FILE` environment variables:

- `ACCESS_KEY`
- `JWT_SECRET`
- `OPENAI_API_KEY`
- `GEMINI_API_KEY`

`ACCESS_KEY` and `JWT_SECRET` are required. The AI provider keys may be empty, in which case the provider is disabled.

## Image tags

By default the deploy script builds:

- `crackoa/api:swarm-local`
- `crackoa/web:swarm-local`

Override them if needed:

```bash
IMAGE_PREFIX=registry.example.com/crackoa IMAGE_TAG=prod ./deploy/swarm/deploy.sh
```

For a multi-node swarm, push those images to a registry reachable by every node:

```bash
IMAGE_PREFIX=registry.example.com/crackoa IMAGE_TAG=prod PUSH_IMAGES=1 ./deploy/swarm/deploy.sh
```
