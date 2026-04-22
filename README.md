# GrokOA

GrokOA is an interview and online assessment helper.

[![Video Demo](https://img.youtube.com/vi/f_PUWYjuU7k/0.jpg)](https://www.youtube.com/watch?v=f_PUWYjuU7k)


## Local Setup

### Prerequisites

- Docker Engine
- Docker Compose

### 1. Setup API env

Create `api/.env` from `api/.env.example` and set:

- `ACCESS_KEY`
- `JWT_SECRET`
- `OPENAI_API_KEY`
- `GEMINI_API_KEY`

If one of AI provider keys is empty, it will be disabled.

### 2. Start the web client and api

From root:

```bash
docker compose up --build
```

Web client starts on `http://localhost`

Stop with:
```bash
docker compose down
```

## Project Structure

- `/api` Go backend application
- `/web` Vue frontend
- `/desktop` desktop client

## Deployment

For Docker Swarm deployment, see [deploy/swarm/README.md](/Users/rostyk/Documents/Projects/crackoa/deploy/swarm/README.md).
