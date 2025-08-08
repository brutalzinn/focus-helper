# ♾️ Focus Helper: ATC Edition

## Um assistente de foco pessoal que simula uma torre de Controle de Tráfego Aéreo (ATC) para ajudar você a se desconectar do computador e evitar o hiperfoco.

### Por que este projeto existe?

Para muitos indivíduos autistas e neurodivergentes, o **hiperfoco** é uma experiência comum. Embora possa ser um superpoder para a produtividade, também pode levar ao esgotamento e à negligência de necessidades básicas. Alertas genéricos para "fazer uma pausa" são frequentemente abruptos, ineficazes e fáceis de ignorar. O **Focus Helper** oferece uma solução inovadora, transformando essas interrupções em uma experiência imersiva e envolvente que respeita os interesses especiais do usuário.

Ao simular uma torre de **Controle de Tráfego Aéreo (ATC)**, ele utiliza uma abordagem criativa e com tema de aviação para entregar alertas críticos. Em vez de uma simples notificação, os usuários recebem uma "chamada de rádio" com instruções urgentes, geradas dinamicamente pela **IA Llama 3.2**, tornando o processo de desconexão mais eficaz e agradável.

### Funcionalidades Principais ♾️

* **Alertas Imersivos e Envolventes**: Em vez de simples pop-ups, a aplicação entrega "chamadas de rádio" com instruções urgentes e com tema de aviação, transformando um alerta disruptivo em uma experiência imersiva.
* **Personalização com IA**: Os alertas são gerados dinamicamente usando um modelo de linguagem, como o **Llama 3.2**, o que torna a experiência mais imprevisível e personalizada do que uma notificação estática.
* **Intervenção no Hiperfoco**: A simulação de ATC é uma forma criativa de interromper um estado de hiperfoco, aproveitando um interesse especial (aviação) para encorajar o usuário a se desconectar do trabalho e fazer uma pausa necessária.
* **Respeito à Neurodiversidade**: O projeto reconhece que lembretes de pausa genéricos podem ser incômodos para muitas pessoas e oferece uma solução que é funcional e respeitosa com diferentes estilos cognitivos.

### Instruções de Configuração com Docker e Makefile 🐳

Este projeto usa o Docker para criar um ambiente portátil e um `Makefile` para simplificar o processo de construção e execução do contêiner.

#### Pré-requisitos

* **Docker e Docker Compose**: Certifique-se de que o Docker e o Docker Compose estão instalados e em execução em seu sistema.
* **Make**: O utilitário `make` é necessário para executar os comandos definidos no `Makefile`.
* **Modelos de Voz**: Você precisa baixar os modelos de voz para que o projeto funcione.
* **Ollama**: Você precisa ter o ollama instalado e rodando um modelo de IA como o llama3.2
#### Baixando os Modelos de Voz

Antes de construir o projeto, você precisa baixar os arquivos do modelo de voz. O projeto usa o modelo de voz **Cadu** para o português do Brasil.

1.  **Crie o diretório `voices`**: Na raiz do seu projeto, crie uma pasta chamada `voices`.

    ```bash
    mkdir voices
    ```

2.  **Baixe o arquivo `.onnx` do modelo**: Este arquivo contém o modelo de rede neural treinado.

    ```bash
    wget -O voices/pt_BR-cadu-medium.onnx https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/pt/pt_BR/cadu/medium/pt_BR-cadu-medium.onnx?download=true
    ```

3.  **Baixe o arquivo de configuração `.json`**: Este arquivo fornece os metadados e as configurações necessárias.

    ```bash
    wget -O voices/pt_BR-cadu-medium.onnx.json https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/pt/pt_BR/cadu/medium/pt_BR-cadu-medium.onnx.json?download=true
    ```

#### Como Usar o `Makefile`

O `Makefile` define vários alvos (comandos) que você pode executar a partir do seu terminal.

1.  **Construir a Imagem Docker**: Este comando compila sua aplicação Go dentro do contêiner e cria a imagem Docker final.
    ```bash
    make build
    ```

2.  **Executar o Contêiner**: Este comando executa o script `entrypoint.sh`, que inicia o contêiner com as configurações necessárias para áudio e GUI.
    ```bash
    make run
    ```
    *Nota: O script `entrypoint.sh` deve estar presente e corretamente configurado para executar o contêiner Docker.*

3.  **Reconstruir a Imagem**: Se você fez alterações no `Dockerfile` ou no código-fonte, este comando removerá a imagem antiga e construirá uma nova.
    ```bash
    make rebuild
    ```

4.  **Parar o Contêiner**: Para parar e remover o contêiner em execução, use:
    ```bash
    make stop
    ```

5.  **Ver os Logs**: Para ver a saída do seu contêiner em tempo real, use:
    ```bash
    make logs
    ```

6.  **Limpar**: Este comando para o contêiner e remove a imagem Docker do seu sistema.
    ```bash
    make clean
    ```