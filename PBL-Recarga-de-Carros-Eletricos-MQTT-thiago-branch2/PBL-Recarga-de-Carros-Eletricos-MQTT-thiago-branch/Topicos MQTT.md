
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
