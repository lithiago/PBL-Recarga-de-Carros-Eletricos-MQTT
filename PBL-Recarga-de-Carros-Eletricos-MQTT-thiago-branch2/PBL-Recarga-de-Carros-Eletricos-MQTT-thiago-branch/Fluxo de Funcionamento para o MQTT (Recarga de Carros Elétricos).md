
Para iniciar é importante esclarecer os conceitos básicos do protocolo MQTT.
### Sobre o MQTT
1. Cliente MQTT: Qualquer dispositivo que utiliza uma biblioteca MQTT. Um cliente pode publicar mensagens (publicador), receber mensagens (assinante), ou fazer ambos.
2. Agente MQTT (Broker):
	1. Receber mensagens dos publicadores;
	2. Filtrar e encaminhar mensagens aos assinantes correspondentes;
	3. Gerenciar sessões e mensagens perdidas;
3. Conexão MQTT: Iniciada pelo cliente através da mensagem ``CONNECT``, e confirmada pelo agente com `CONNACK`. Toda a comunicação ocorre sobre TCP/IP. Os clientes nunca se conectam entre si diretamente sempre passam pelo agente.
#### Tópicos MQTT
Um tópico é como o endereço da mensagem. Uma string hierárquica, usada pelo broker para organizar e filtrar mensagens. A ideia segue o principio de que alguém publica uma mensagem em um tópico. E outros assim esse tópicos para receber essas mensagens.

- Formato de um possível tópico do nosso problema:
	``` 
	car/car_id/request/reservation

	Legenda:
	car -> Tudo que envolve carro
	car_id -> Id do Carro
	request -> Tipo de Operação
	reservation -> Subcategoria da Operação
	```
	Esse tópico atuará como uma etiqueta para cada mensagem de solicitação de reserva
### Funcionamento do MQTT

1. **Etapa 1: Conexão**
	1. Antes de qualquer interação os componentes (carro, servidor e ponto de recarga) precisam se conectar ao broker (o agente MQTT).
		1. Cada componente envia uma mensagem `CONNECT` para o broker e espera pelo retorno `CONNACK`.
2. **Etapa 2: Assinar (Subscribe)**
	1. Depois de conectado, cada componente diz quais tipos de mensagens quer receber
3. **Etapa 3: Publicar (Publish):
	1. Agora os componentes categorizados como publicadores (publishers) enviam mensagens para o tópico**
	2. O broker recebe a mensagem e verifica quem assinou o tópico (no caso, o servidor).
4. **Etapa 4: O broker entrega a mensagem**
	1. O broker encaminha a mensagem automaticamente para todos os que assinaram esse tópico
5. **Etapa 5: O servidor responde**:
	1. O servidor publica a resposta para um tópico
		1. Possível tópico usado no nosso problema:
		```
		server/response/car_id
		```
		O cliente que estiver inscrito nesse tópico, recebe a mensagem.



# 📋 Tabela Completa de Tópicos do Sistema

| Tópico                                 | Quem Publica     | Quem Assina                       | Descrição                                                                          |
| -------------------------------------- | ---------------- | --------------------------------- | ---------------------------------------------------------------------------------- |
| `car/<car_id>/request/reservation`     | Carro            | Servidor                          | O carro solicita reserva enviando posição, destino e bateria atual.                |
| `car/<car_id>/request/status`          | Carro            | Servidor                          | O carro envia atualizações de status (posição, energia, etc.).                     |
| `car/<car_id>/request/cancel`          | Carro            | Servidor                          | O carro cancela uma reserva feita anteriormente.                                   |
| `server/<server_id>/response/<car_id>` | Servidor         | Carro                             | O servidor responde para o carro com opções de recarga, confirmações ou rotas.     |
| `server/<server_id>/notify/<car_id>`   | Servidor         | Carro                             | Notificações gerais: mudanças na reserva, congestionamento, avisos de parada, etc. |
| `station/<station_id>/command/reserve` | Servidor         | Ponto de Recarga                  | Comando para reservar horário para um carro específico.                            |
| `station/<station_id>/command/cancel`  | Servidor         | Ponto de Recarga                  | Comando para cancelar uma reserva em andamento.                                    |
| `station/<station_id>/command/start`   | Servidor         | Ponto de Recarga                  | Comando para iniciar a recarga do carro no ponto.                                  |
| `station/<station_id>/command/stop`    | Servidor         | Ponto de Recarga                  | Comando para parar a recarga do carro no ponto.                                    |
| `station/<station_id>/status`          | Ponto de Recarga | Servidor                          | Status atual do ponto: livre, reservado, em manutenção, recarregando.              |
| `station/<station_id>/event/started`   | Ponto de Recarga | Servidor                          | Evento indicando que a recarga foi iniciada.                                       |
| `station/<station_id>/event/finished`  | Ponto de Recarga | Servidor                          | Evento indicando que a recarga foi concluída.                                      |
| `station/<station_id>/queue`           | Ponto de Recarga | Servidor                          | Atualização da fila de espera para recarga no ponto.                               |
| `telemetry/car/<car_id>`               | Carro            | Sistema de Monitoramento/Servidor | Dados de telemetria em tempo real: localização, bateria, etc.                      |
| `telemetry/station/<station_id>`       | Ponto de Recarga | Sistema de Monitoramento/Servidor | Telemetria do ponto: energia consumida, tempo de ocupação, etc.                    |