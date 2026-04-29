# Agents Runtime Images

This repository contains Docker image definitions for the runtimes used by the ClawManager Agents project.

Each top-level runtime directory can produce one container image. At the moment, the repository includes:

- `hermes`
- `openclaw`

## Repository layout

- `hermes/`: Hermes runtime image
- `openclaw/`: OpenClaw runtime image

## Manual builds

You can build each runtime image directly with Docker from the repository root.

### Hermes

```bash
docker build \
  -f hermes/Dockerfile \
  -t hermes:local \
  ./hermes
```

### OpenClaw

```bash
docker build \
  -f openclaw/Dockerfile.openclaw \
  -t openclaw:local \
  ./openclaw
```

## Manual multi-architecture builds

If you want to build multi-architecture images manually, use Docker Buildx.

### Hermes

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f hermes/Dockerfile \
  -t <registry>/hermes:latest \
  --push \
  ./hermes
```

### OpenClaw

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f openclaw/Dockerfile.openclaw \
  -t <registry>/openclaw:latest \
  --push \
  ./openclaw
```

The supported architectures are:

- `linux/amd64`
- `linux/arm64`

This publishes a single multi-arch image manifest, so both architectures are available under the same tag.
