FROM golang:1.23.2-bookworm AS builder

RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \
    libgtk-3-dev \
    libasound2-dev \
    libx11-dev \
    libxtst-dev \
    x11proto-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /focus-helper .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    sox \
    ffmpeg \
    pulseaudio-utils \
    libgtk-3-0 \
    libasound2 \
    libxtst6 \
    libx11-6 \
    libespeak-ng1 \
    libxext6 \
    libxrandr2 \
    libasound2-plugins \
    libcanberra-gtk-module \
    libcanberra-gtk3-module \
    x11-xserver-utils \
    wget \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

RUN wget https://github.com/rhasspy/piper/releases/download/v1.2.0/piper_amd64.tar.gz -O /tmp/piper.tar.gz && \
    mkdir -p /opt/piper && \
    tar -zxvf /tmp/piper.tar.gz -C /opt/piper --strip-components=1 && \
    ln -s /opt/piper/piper /usr/local/bin/piper && \
    rm /tmp/piper.tar.gz

ENV LD_LIBRARY_PATH=/opt/piper

WORKDIR /app

COPY --from=builder /focus-helper /app/focus-helper
COPY ./assets ./assets
COPY ./voices ./voices

CMD ["/app/focus-helper"]
