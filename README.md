# Recarga de Carros Elétricos Inteligentes


- **Componentes**: Lucas de Andrade Pereira Mendes, Thiago Ramon Santos de Jesus e Felipe Pinto Silva

- **Professor**: Elinaldo Santos de Gois Junior



## Descrição Rápida
Este projeto simula um sistema distribuído para recarga de carros elétricos, onde carros solicitam rotas e reservam postos de recarga em diferentes cidades. A comunicação entre os componentes é realizada utilizando o protocolo MQTT, e a coordenação de reservas entre múltiplos servidores é feita através de um algoritmo de Two-Phase Commit (2PC).

O sistema consiste em três componentes principais:
* **Carro**: Simula um carro elétrico que busca rotas, solicita reservas em postos de recarga e gerencia sua bateria. Ele interage com o usuário para escolher destinos e visualizar informações. Os dados gerados pelo software do carro são fictícios e aleatórios, simulando a descarga da bateria.
* **Servidor**: Atua como um coordenador central em cada cidade, responsável por receber as solicitações dos carros, calcular rotas ideais com base na autonomia do veículo e na disponibilidade dos postos, e coordenar as reservas em postos que podem estar sob a gerência de outros servidores através do 2PC.
* **Mosquitto (Broker MQTT)**: Servidor de mensagens que permite a comunicação assíncrona e desacoplada entre os carros e os servidores.

O sistema demonstra a aplicação de conceitos de sistemas distribuídos, como comunicação assíncrona, tolerância a falhas (através do LWT do MQTT e da lógica de desconexão/cancelamento), e consistência de dados (com o 2PC para reservas multi-servidor).

## Objetivos do Projeto

O principal objetivo deste projeto é aprimorar um sistema de recarga inteligente de veículos elétricos para suportar o planejamento e a reserva antecipada de múltiplos pontos de recarga ao longo de uma rota específica entre cidades e estados. Para isso, foram estabelecidos os seguintes objetivos específicos:

* **Garantir a disponibilidade sequencial**: O sistema deve assegurar a disponibilidade dos pontos de recarga necessários para que o veículo complete sua viagem dentro de um cronograma previsto, com paradas planejadas de forma otimizada e segura
* **Requisição Atômica**: Através de uma requisição atômica, o sistema deve consultar a disponibilidade e reservar uma sequência de pontos de recarga, eliminando o risco de o veículo ficar sem energia ou sofrer atrasos imprevistos devido à indisponibilidade de carregadores.
* **Comunicação Padronizada e Coordenada**: É essencial que exista uma forma padronizada e coordenada de comunicação entre os servidores das empresas conveniadas envolvidas
* **API de Comunicação entre Servidores**: A comunicação entre os servidores deve ser realizada por meio de uma API REST, projetada para permitir que um cliente (carro), a partir de qualquer servidor, reserve pontos de carregamento disponíveis em diferentes empresas conveniadas, seguindo as regras do sistema centralizado original.
* **Emulação Realista**: Os elementos da arquitetura devem ser executados em contêineres Docker, simulando um cenário distribuído em computadores distintos
* **Comunicação Carro-Servidor via MQTT**: A comunicação entre os carros e os servidores deve adotar o protocolo MQTT (Message Queue Telemetry Transport), um padrão utilizado na Internet das Coisas (IoT)

## Tecnologias e Requisitos

* **Linguagem de Programação**: Go (Golang).
* **Comunicação Principal**: MQTT (Message Queuing Telemetry Transport) com o broker Mosquitto.
    * Utiliza tópicos bem definidos para diferentes tipos de mensagens (solicitações de rotas, reservas, status, desconexões).
    * Implementa Last Will and Testament (LWT) para detecção de desconexões inesperadas de carros.
* **Comunicação API (Interna entre Servidores)**: HTTP RESTful API com o framework Gin.
    * Utilizada para que os servidores troquem informações sobre postos de recarga e coordenem o Two-Phase Commit.
* **Coordenação Distribuída**: Two-Phase Commit (2PC) para garantir atomicidade nas operações de reserva que envolvem múltiplos postos gerenciados por diferentes servidores.
* **Orquestração/Containerização**: Docker e Docker Compose.
    * Facilita a configuração e execução dos múltiplos componentes (broker MQTT, carros e servidores de diferentes cidades) em um ambiente isolado.
* **Armazenamento de Dados**: Arquivos JSON para simular o armazenamento de dados de postos de recarga e rotas.

## Como Executar

Para executar o projeto, siga os passos abaixo:

### Pré-requisitos

* Docker e Docker Compose instalados em sua máquina.

### Passos

1.  **Navegue até a pasta `MQTT`**:
    ```bash
    cd <caminho_do_repositorio>/MQTT
    ```

2.  **Construa as imagens Docker**:
    Este comando irá construir as imagens para o `carro`, `servidor-feiradesantana`, `servidor-ilheus`, e `servidor-salvador` com base nos `Dockerfile`s e o serviço `mosquitto`.
    ```bash
    docker-compose build
    ```

3.  **Inicie os serviços**:
    Este comando levantará todos os contêineres definidos no `docker-compose.yml`, incluindo o broker Mosquitto, três instâncias de servidores (uma para cada cidade: Feira de Santana, Ilhéus e Salvador) e uma instância de carro.
    ```bash
    docker-compose up
    ```
4. **Executando o carro**:
    Este comando irá permitir a interação com o carro no terminal.
   ```bash
   docker-compose run --rm carro
   ``` 

6.  **Interação com o Carro**:
    Após os serviços serem iniciados, o terminal do contêiner `carro` exibirá um menu interativo:
    ```
    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
            🚀 MENU PRINCIPAL 🚀
    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
      🆔 Carro ID: <IP_DO_CARRO>
      🔋 Bateria: XX.XX%
      1️⃣  | Solicitar Nova Rota
      2️⃣  | Cancelar Rota Atual
      3️⃣  | Finalizar Recarga
      4️⃣  | Encerrar Conexão
    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    Digite a opção desejada:
    ```
    Você pode interagir com o sistema digitando as opções no terminal do carro.

7.  **Visualizando Logs**:
    Para visualizar os logs de um serviço específico (por exemplo, `servidor-feiradesantana`), abra outro terminal e execute:
    ```bash
    docker-compose logs -f servidor-feiradesantana
    ```
    Isso permitirá que você acompanhe o fluxo das mensagens e o funcionamento do sistema em tempo real.

8.  **Parando os serviços**:
    Para parar e remover todos os contêineres, redes e volumes criados pelo `docker-compose`, pressione `Ctrl+C` no terminal onde o `docker-compose up` está rodando e então execute:
    ```bash
    docker-compose down
    ```
