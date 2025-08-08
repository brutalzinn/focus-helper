# Makefile para o projeto Focus Helper

IMAGE_NAME := focus-helper
CONTAINER_NAME := focus-helper-container

.PHONY: all build rebuild run stop clean logs help

all: build

build:
	@echo "--> Construindo a imagem Docker '${IMAGE_NAME}'..."
	@docker build -t $(IMAGE_NAME) .

rebuild: clean build

# O comando 'run' agora simplesmente executa nosso script robusto.
run:
	@echo "--> Executando o contêiner via script 'run-docker.sh'..."
	@./entrypoint.sh

stop:
	@echo "--> Parando e removendo o contêiner '${CONTAINER_NAME}'..."
	@docker stop $(CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(CONTAINER_NAME) 2>/dev/null || true

clean: stop
	@echo "--> Removendo a imagem Docker '${IMAGE_NAME}'..."
	@docker rmi $(IMAGE_NAME) 2>/dev/null || true

logs:
	@echo "--> Exibindo logs do contêiner '${CONTAINER_NAME}'..."
	@docker logs -f $(CONTAINER_NAME)

help:
	@echo "Uso: make [alvo]"
	@echo "  build    Constrói a imagem Docker."
	@echo "  rebuild  Força a remoção e reconstrução da imagem."
	@echo "  run      Executa o contêiner com acesso a GUI/Áudio."
	@echo "  stop     Para e remove o contêiner."
	@echo "  clean    Remove a imagem Docker."
	@echo "  logs     Exibe os logs do contêiner."

.DEFAULT_GOAL := help