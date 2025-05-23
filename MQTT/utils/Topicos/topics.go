package topics

import (
	"fmt"
	"strings"
)

// Carro → Servidor
func CarroRequestReserva(carID string, serverID string, cidade string) string     { return fmt.Sprintf("car/%s/request/reserva/%s/%s", carID, cidade, serverID) }
func CarroRequestRotas(carID string, cidade string) string     { return fmt.Sprintf("car/%s/request/rotas/%s", carID, strings.ToLower(cidade)) }
func CarroDesconectado(carID string) string {return fmt.Sprintf("car/%s/desconectado", carID)}
func CarroRequestToServer(carID string, cidade string, TipoDeSolicitacao string) string     { return fmt.Sprintf("car/%s/request/%s/%s", carID, cidade, TipoDeSolicitacao) }

func CarroSendsRechargeStart(carID string, serverID, cidade string) string { return fmt.Sprintf("car/%s/recharge/start/%s", carID, strings.ToLower(cidade)) }
func CarroSendsRechargeFinish(carID string) string { return fmt.Sprintf("car/%s/recharge/finish/", carID) }
func CarroRequestStatus(carID string, serverID string, cidade string) string      { return fmt.Sprintf("car/%s/request/status/%s/%s", carID, cidade, serverID) }
func CarroRequestCancel(carID string) string      { return fmt.Sprintf("car/%s/request/cancel//", carID) }


// Servidor → Carro
func ServerResponseToCar(carID string) string { return fmt.Sprintf("server/response/%s", carID) }
func ServerNotifyCar(serverID, carID string) string     { return fmt.Sprintf("server/%s/notify/%s", serverID, carID) }
func ServerReserveStatus(serverID, carID string) string     { return fmt.Sprintf("server/%s/ReserveStatus/%s", serverID, carID) }
func ServerResponteRoutes(carID string, cidade string) string { return fmt.Sprintf("server/%s/rotas/%s", carID, strings.ToLower(cidade))}
func ServerDesconectado(serverID string) string {return fmt.Sprintf("server/%s/desconectado", serverID)}

// Servidor → Posto
func ServerCommandReserve(stationID string) string { return fmt.Sprintf("station/%s/command/reserve", stationID) }
func ServerCommandCancel(stationID string) string  { return fmt.Sprintf("station/%s/command/cancel", stationID) }
func ServerCommandStart(stationID string) string   { return fmt.Sprintf("station/%s/command/start", stationID) }
func ServerCommandStop(stationID string) string    { return fmt.Sprintf("station/%s/command/stop", stationID) }

// Posto → Servidor
func StationStatus(stationID string) string        { return fmt.Sprintf("station/%s/status", stationID) }
func StationEventStarted(stationID string) string  { return fmt.Sprintf("station/%s/event/started", stationID) }
func StationEventFinished(stationID string) string { return fmt.Sprintf("station/%s/event/finished", stationID) }
