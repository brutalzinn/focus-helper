#DONT WORK YET

FROM golang:1.23.2-bookworm AS builder

RUN apt-get update && apt-get install -y \
    build-essential \
    ffmpeg \
    sox \
    libsox-fmt-mp3 \
    mplayer \
    libasound2-dev \
    wget \
    ca-certificates \
    libgtk-3-dev \
    libx11-dev \
    libxtst-dev \
    libpng-dev \
    libxtst6 \
    libxext6 \
    && rm -rf /var/lib/apt/lists/* && rm -rf /var/lib/apt/lists/*


RUN wget https://github.com/rhasspy/piper/releases/download/2023.11.14-2/piper_linux_x86_64.tar.gz && \
    tar -xzvf piper_linux_x86_64.tar.gz && \
    mv piper/piper /usr/local/bin/ && \
    rm piper_linux_x86_64.tar.gz && \
    rm -rf piper

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /focus-helper .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ffmpeg \
    sox \
    libsox-fmt-mp3 \
    mplayer \
    libasound2 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /focus-helper /usr/local/bin/focus-helper

COPY --from=builder /usr/local/bin/piper /usr/local/bin/piper

WORKDIR /app
COPY ./voices ./voices
COPY ./audio ./audio
COPY ./assets ./assets

CMD ["/usr/local/bin/focus-helper"]
