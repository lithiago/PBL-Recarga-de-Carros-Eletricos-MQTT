package topics

import "fmt"

// Carro → Servidor
func CarroConnect() string      { return "car/+/request/connect" }
func CarroRequestReserva(carID string) string     { return fmt.Sprintf("car/%s/request/reservation", carID) }
func CarroRequestRotas(carID string) string     { return fmt.Sprintf("car/%s/request/routes", carID) }
func CarroRequestStatus(carID string) string      { return fmt.Sprintf("car/%s/request/status", carID) }
func CarroRequestCancel(carID string) string      { return fmt.Sprintf("car/%s/request/cancel", carID) }

// Servidor → Carro
func ServerResponseToCar(carID string) string { return fmt.Sprintf("server/response/%s", carID) }
func ServerNotifyCar(serverID, carID string) string     { return fmt.Sprintf("server/%s/notify/%s", serverID, carID) }

// Servidor → Posto
func ServerCommandReserve(stationID string) string { return fmt.Sprintf("station/%s/command/reserve", stationID) }
func ServerCommandCancel(stationID string) string  { return fmt.Sprintf("station/%s/command/cancel", stationID) }
func ServerCommandStart(stationID string) string   { return fmt.Sprintf("station/%s/command/start", stationID) }
func ServerCommandStop(stationID string) string    { return fmt.Sprintf("station/%s/command/stop", stationID) }

// Posto → Servidor
func StationStatus(stationID string) string        { return fmt.Sprintf("station/%s/status", stationID) }
func StationEventStarted(stationID string) string  { return fmt.Sprintf("station/%s/event/started", stationID) }
func StationEventFinished(stationID string) string { return fmt.Sprintf("station/%s/event/finished", stationID) }
