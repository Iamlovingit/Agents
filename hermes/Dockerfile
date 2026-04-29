# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.23-bookworm AS hermes-agent-builder

WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH
ARG AGENT_VERSION=dev

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN set -eux; \
    arch="${TARGETARCH:-amd64}"; \
    CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${arch}" \
        go build -trimpath -ldflags="-s -w -X main.version=${AGENT_VERSION}" \
        -o /out/hermes-agent ./cmd/hermes-agent

FROM lscr.io/linuxserver/webtop:ubuntu-kde

ARG HERMES_BRANCH=main
ARG NODE_MAJOR=22
ARG TARGETARCH
ARG DEBIAN_FRONTEND=noninteractive

ENV HERMES_HOME=/config/.hermes \
    HERMES_SKILL_DIRS=/config/hermes/skills:/config/.hermes/skills \
    PLAYWRIGHT_BROWSERS_PATH=/usr/local/share/ms-playwright \
    TITLE="Hermes Runtime" \
    SUBFOLDER="/" \
    PATH=/opt/node/bin:/usr/local/bin:${PATH}

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        bash \
        build-essential \
        ca-certificates \
        curl \
        ffmpeg \
        git \
        libffi-dev \
        libjpeg-dev \
        libolm-dev \
        libssl-dev \
        libxml2-dev \
        libxslt1-dev \
        openssh-client \
        pkg-config \
        portaudio19-dev \
        ripgrep \
        sudo \
        tar \
        unzip \
        xz-utils \
        zlib1g-dev; \
    rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    arch="${TARGETARCH:-$(dpkg --print-architecture)}"; \
    case "${arch}" in \
        amd64|x86_64) node_arch="x64" ;; \
        arm64|aarch64) node_arch="arm64" ;; \
        arm|armhf) node_arch="armv7l" ;; \
        *) echo "Unsupported architecture for Node.js: ${arch}" >&2; exit 1 ;; \
    esac; \
    node_index="https://nodejs.org/dist/latest-v${NODE_MAJOR}.x/"; \
    node_tar="$(curl -fsSL "${node_index}" | grep -oE "node-v${NODE_MAJOR}\.[0-9]+\.[0-9]+-linux-${node_arch}\.tar\.xz" | head -1)"; \
    test -n "${node_tar}"; \
    curl -fsSL "${node_index}${node_tar}" -o /tmp/node.tar.xz; \
    mkdir -p /opt/node; \
    tar -xJf /tmp/node.tar.xz -C /opt/node --strip-components=1; \
    ln -sf /opt/node/bin/node /usr/local/bin/node; \
    ln -sf /opt/node/bin/npm /usr/local/bin/npm; \
    ln -sf /opt/node/bin/npx /usr/local/bin/npx; \
    node --version; \
    npm --version; \
    rm -f /tmp/node.tar.xz

RUN set -eux; \
    mkdir -p "${HERMES_HOME}" "${PLAYWRIGHT_BROWSERS_PATH}"; \
    export HOME=/root; \
    curl -fsSL https://raw.githubusercontent.com/NousResearch/hermes-agent/main/scripts/install.sh \
        | bash -s -- --branch "${HERMES_BRANCH}" --hermes-home "${HERMES_HOME}" --skip-setup; \
    /usr/local/lib/hermes-agent/venv/bin/python -m ensurepip --upgrade; \
    /usr/local/lib/hermes-agent/venv/bin/python -m pip install --no-cache-dir "hermes-agent[dingtalk]"; \
    chmod -R a+rX /usr/local/lib/hermes-agent "${PLAYWRIGHT_BROWSERS_PATH}"; \
    chown -R abc:abc "${HERMES_HOME}" || true; \
    hermes --help >/dev/null; \
    rm -rf /root/.cache /tmp/*

COPY --from=hermes-agent-builder /out/hermes-agent /usr/local/bin/hermes-agent
COPY rootfs/ /

ENV SHELL=/bin/bash

RUN <<'EOF'
set -eux
hermes_python="$(readlink -f /usr/local/lib/hermes-agent/venv/bin/python)"
hermes_python_root="$(dirname "$(dirname "${hermes_python}")")"
hermes_new_python_root="/opt/uv/python/$(basename "${hermes_python_root}")"
mkdir -p /opt/uv/python
cp -a "${hermes_python_root}" /opt/uv/python/
ln -sfn "${hermes_new_python_root}/bin/$(basename "${hermes_python}")" /usr/local/lib/hermes-agent/venv/bin/python
sed -i "s|/root/.local/share/uv/python/[^/]*/bin|${hermes_new_python_root}/bin|g" /usr/local/lib/hermes-agent/venv/pyvenv.cfg
chmod -R a+rX /opt/uv /usr/local/lib/hermes-agent/venv
rm -rf /root/.local/share/uv/python
chmod 755 /usr/local/bin/hermes-agent /usr/local/bin/hermes-apply-runtime-config /usr/local/bin/start-hermes-terminal /usr/local/bin/start-hermes-gateway

mkdir -p /custom-cont-init.d
cat >/custom-cont-init.d/99-hermes-config <<'SCRIPT'
#!/usr/bin/with-contenv bash
set -e

HERMES_HOME="${HERMES_HOME:-/config/.hermes}"
HERMES_INSTALL_DIR="/usr/local/lib/hermes-agent"
export HERMES_HOME

mkdir -p \
    /config/.cache/npm \
    /config/.cache/uv \
    /config/.config \
    /config/.config/autostart \
    /config/.config/openbox \
    /config/.local/share \
    "${HERMES_HOME}/cron" \
    "${HERMES_HOME}/sessions" \
    "${HERMES_HOME}/logs" \
    "${HERMES_HOME}/pairing" \
    "${HERMES_HOME}/hooks" \
    "${HERMES_HOME}/image_cache" \
    "${HERMES_HOME}/audio_cache" \
    "${HERMES_HOME}/memories" \
    "${HERMES_HOME}/skills" \
    "${CLAWMANAGER_AGENT_PERSISTENT_DIR:-/config}/hermes-agent"

if [ ! -f "${HERMES_HOME}/.env" ]; then
    if [ -f "${HERMES_INSTALL_DIR}/.env.example" ]; then
        cp "${HERMES_INSTALL_DIR}/.env.example" "${HERMES_HOME}/.env"
    else
        touch "${HERMES_HOME}/.env"
    fi
fi

if [ ! -f "${HERMES_HOME}/config.yaml" ] && [ -f "${HERMES_INSTALL_DIR}/cli-config.yaml.example" ]; then
    cp "${HERMES_INSTALL_DIR}/cli-config.yaml.example" "${HERMES_HOME}/config.yaml"
fi

if [ ! -f "${HERMES_HOME}/SOUL.md" ]; then
    cat >"${HERMES_HOME}/SOUL.md" <<'SOUL'
# Hermes Agent Persona

<!--
This file defines the agent's personality and tone.
Edit this to customize how Hermes communicates with you.
Delete the contents or this file to use the default personality.
-->
SOUL
fi

if [ -x "${HERMES_INSTALL_DIR}/venv/bin/python" ] && [ -f "${HERMES_INSTALL_DIR}/tools/skills_sync.py" ]; then
    "${HERMES_INSTALL_DIR}/venv/bin/python" "${HERMES_INSTALL_DIR}/tools/skills_sync.py" >/dev/null 2>&1 || true
fi

if [ -x /usr/local/bin/hermes-apply-runtime-config ]; then
    /usr/local/bin/hermes-apply-runtime-config || true
fi

cat >/config/.config/autostart/hermes-runtime.desktop <<'DESKTOP'
[Desktop Entry]
Type=Application
Name=Hermes Runtime
Comment=Open Hermes automatically
Exec=/usr/local/bin/start-hermes-terminal
Terminal=false
X-KDE-Autostart-enabled=true
DESKTOP
chmod 755 /config/.config/autostart/hermes-runtime.desktop

touch /config/.config/openbox/autostart
if ! grep -q 'start-hermes-terminal' /config/.config/openbox/autostart 2>/dev/null; then
    {
        echo ''
        echo '# Hermes Runtime autostart'
        echo '(/usr/local/bin/start-hermes-terminal) &'
    } >> /config/.config/openbox/autostart
fi
chmod 755 /config/.config/openbox/autostart

if id abc >/dev/null 2>&1; then
    chown -R abc:abc /config/.cache /config/.config /config/.local || true
    chown -R abc:abc "${HERMES_HOME}" || true
    chown -R abc:abc "${CLAWMANAGER_AGENT_PERSISTENT_DIR:-/config}/hermes-agent" || true
fi
SCRIPT
chmod 755 /custom-cont-init.d/99-hermes-config

mkdir -p \
    /etc/s6-overlay/s6-rc.d/hermes-agent/dependencies.d \
    /etc/s6-overlay/s6-rc.d/hermes-gateway/dependencies.d \
    /etc/s6-overlay/s6-rc.d/user/contents.d
echo longrun >/etc/s6-overlay/s6-rc.d/hermes-agent/type
touch /etc/s6-overlay/s6-rc.d/hermes-agent/dependencies.d/init-services
cat >/etc/s6-overlay/s6-rc.d/hermes-agent/run <<'SCRIPT'
#!/usr/bin/with-contenv bash
set -euo pipefail

if [ "${CLAWMANAGER_AGENT_ENABLED:-false}" != "true" ]; then
    echo "ClawManager Hermes agent disabled"
    sleep infinity
fi

export HOME="${HOME:-/config}"
export HERMES_HOME="${HERMES_HOME:-/config/.hermes}"
export XDG_CACHE_HOME="${XDG_CACHE_HOME:-/config/.cache}"
export XDG_CONFIG_HOME="${XDG_CONFIG_HOME:-/config/.config}"
export XDG_DATA_HOME="${XDG_DATA_HOME:-/config/.local/share}"
export NPM_CONFIG_CACHE="${NPM_CONFIG_CACHE:-/config/.cache/npm}"
export UV_CACHE_DIR="${UV_CACHE_DIR:-/config/.cache/uv}"

mkdir -p \
    "${CLAWMANAGER_AGENT_PERSISTENT_DIR:-/config}/hermes-agent" \
    "${XDG_CACHE_HOME}/npm" \
    "${XDG_CACHE_HOME}/uv" \
    "${XDG_CONFIG_HOME}" \
    "${XDG_DATA_HOME}" \
    "${HERMES_HOME}"

/usr/local/bin/hermes-apply-runtime-config || true

if command -v s6-setuidgid >/dev/null 2>&1 && id abc >/dev/null 2>&1; then
    chown -R abc:abc "${XDG_CACHE_HOME}" "${XDG_CONFIG_HOME}" "${XDG_DATA_HOME}" "${HERMES_HOME}" || true
    chown -R abc:abc "${CLAWMANAGER_AGENT_PERSISTENT_DIR:-/config}/hermes-agent" || true
    exec s6-setuidgid abc /usr/local/bin/hermes-agent
fi

exec /usr/local/bin/hermes-agent
SCRIPT
chmod 755 /etc/s6-overlay/s6-rc.d/hermes-agent/run
touch /etc/s6-overlay/s6-rc.d/user/contents.d/hermes-agent

echo longrun >/etc/s6-overlay/s6-rc.d/hermes-gateway/type
touch /etc/s6-overlay/s6-rc.d/hermes-gateway/dependencies.d/init-services
cat >/etc/s6-overlay/s6-rc.d/hermes-gateway/run <<'SCRIPT'
#!/usr/bin/with-contenv bash
set -euo pipefail

exec /usr/local/bin/start-hermes-gateway
SCRIPT
chmod 755 /etc/s6-overlay/s6-rc.d/hermes-gateway/run
touch /etc/s6-overlay/s6-rc.d/user/contents.d/hermes-gateway
EOF
