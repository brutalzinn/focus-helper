#!/bin/bash

# Encerra o script imediatamente se qualquer comando falhar.
set -e

# --- Variáveis de Configuração ---
IMAGE_NAME="focus-helper"
CONTAINER_NAME="focus-helper-container"

# --- Validação do Ambiente do Host ---
echo "--> Validando ambiente do host..."
if [ -z "$DISPLAY" ]; then
    echo "ERRO: A variável de ambiente DISPLAY não está definida. Não é possível conectar a uma GUI."
    exit 1
fi
if [ -z "$XDG_RUNTIME_DIR" ]; then
    echo "ERRO: A variável de ambiente XDG_RUNTIME_DIR não está definida. Não é possível conectar ao áudio."
    exit 1
fi
if [ ! -S "${XDG_RUNTIME_DIR}/pulse/native" ]; then
    echo "ERRO: O socket do PulseAudio não foi encontrado em ${XDG_RUNTIME_DIR}/pulse/native."
    exit 1
fi
echo "--> Ambiente validado com sucesso."

# --- Preparação ---
echo "--> Adicionando permissões de acesso ao X11..."
xhost +local:docker

AUDIO_GID=$(getent group audio | cut -d: -f3 || echo 'audio')

# --- Execução do Contâiner ---
echo "--> Executando o contêiner '${CONTAINER_NAME}'..."
docker run --rm -it \
    --name "${CONTAINER_NAME}" \
    --security-opt seccomp=unconfined \
    -v "$(pwd)/assets:/app/assets" \
    -v "$(pwd)/voices:/app/voices" \
    -e OTO_DRIVER=pulseaudio \
    -e PULSE_SERVER="unix:${XDG_RUNTIME_DIR}/pulse/native" \
    -v "${XDG_RUNTIME_DIR}/pulse/native:${XDG_RUNTIME_DIR}/pulse/native" \
    -v "$HOME/.config/pulse/cookie:/root/.config/pulse/cookie" \
    --group-add "${AUDIO_GID}" \
    -e DISPLAY="$DISPLAY" \
    -v /tmp/.X11-unix:/tmp/.X11-unix:ro \
    "${IMAGE_NAME}"