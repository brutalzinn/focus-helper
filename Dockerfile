# Stage 1: The Build Environment
# Stage 1: The Build Environment
FROM ubuntu:22.04 AS builder

# Set environment variables to avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install Go 1.24.2 and build dependencies
RUN apt-get update && apt-get install -y \
    software-properties-common \
    && add-apt-repository -y ppa:longsleep/golang-backports \
    && apt-get update \
    && apt-get install -y golang-1.24-go \
    && ln -s /usr/lib/go-1.24/bin/go /usr/local/bin/go \
    && ln -s /usr/lib/go-1.24/bin/gofmt /usr/local/bin/gofmt

# Enable universe repository and install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    pkg-config \
    libgtk-3-dev \
    libasound2-dev \
    libx11-dev \
    libxtst-dev \
    x11proto-dev \
    libportaudio2 \
    portaudio19-dev \
    git \
    && rm -rf /var/lib/apt/lists/*

# Clone whisper.cpp
WORKDIR /app
RUN git clone https://github.com/ggerganov/whisper.cpp.git
WORKDIR /app/whisper.cpp

# ==================== MUDANÇAS AQUI ====================

# Compila whisper.cpp e ggml usando cmake para um build mais robusto e estático
RUN cmake -S . -B build -DBUILD_SHARED_LIBS=OFF
RUN cmake --build build --target whisper ggml -j$(nproc)

# Retorna ao diretório da aplicação Go
WORKDIR /app/src

# Copia e baixa as dependências Go
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Compila o binário Go com as flags CGO corrigidas
RUN CGO_ENABLED=1 \
    CGO_CFLAGS="-I/app/whisper.cpp -I/app/whisper.cpp/ggml" \
    CGO_LDFLAGS="-L/app/whisper.cpp/build -lwhisper -lggml -lstdc++ -lm -lportaudio" \
    go build -o /focus-helper .

# ... (resto do seu arquivo)

# Stage 3: Final Runtime Image
FROM ubuntu:22.04

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    libasound2 \
    libgtk-3-0 \
    libxtst6 \
    libx11-6 \
    libespeak-ng1 \
    libxext6 \
    libxrandr2 \
    libcanberra-gtk-module \
    libcanberra-gtk3-module \
    libportaudio2 \
    sox \
    ffmpeg \
    pulseaudio-utils \
    wget \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Download and install Piper
RUN wget https://github.com/rhasspy/piper/releases/download/v1.2.0/piper_amd64.tar.gz -O /tmp/piper.tar.gz && \
    mkdir -p /opt/piper && \
    tar -zxvf /tmp/piper.tar.gz -C /opt/piper --strip-components=1 && \
    ln -s /opt/piper/piper /usr/local/bin/piper && \
    rm /tmp/piper.tar.gz

# Copy the built binary and whisper.cpp library
WORKDIR /app
COPY --from=builder /focus-helper /app/focus-helper
COPY --from=builder /app/whisper.cpp/libwhisper.a /app/libwhisper.a

# Create a non-root user for security
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# Set the entry point
ENTRYPOINT ["/app/focus-helper"]