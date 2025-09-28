package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/syucream/posix_mq/src/posix_mq"
)

var rooms = []string{"general", "deportes", "musica"}

type Message struct {
	Type      string // "register" o "chat"
	ClientId  string
	Name      string
	Text      string
	DateTime  string
	QueueName string // solo en "register"
}

// Cola de entrada por sala
var inputQueues = make(map[string]*posix_mq.MessageQueue)

// Mapa de sala → colas de salida de clientes
var outputQueues = make(map[string]map[string]*posix_mq.MessageQueue)

func initRooms() {
	for _, room := range rooms {
		inName := "/" + room + "_in"
		q, err := posix_mq.NewMessageQueue(inName, posix_mq.O_CREAT|posix_mq.O_RDONLY, 0666, nil)
		if err != nil {
			log.Fatalf("Error creando cola de entrada para sala %s: %v", room, err)
		}
		inputQueues[room] = q
		outputQueues[room] = make(map[string]*posix_mq.MessageQueue)

		fmt.Printf("Sala creada: %s\n", room)
	}
}

func cleanup() {
	for _, q := range inputQueues {
		q.Close()
		q.Unlink()
	}
	for _, clients := range outputQueues {
		for _, q := range clients {
			q.Close()
			q.Unlink()
		}
	}
}

func main() {
	fmt.Println("Servidor iniciado con POSIX MQ")
	initRooms()

	// Capturar señales
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		fmt.Println("\nDeteniendo servidor...")
		cleanup()
		os.Exit(0)
	}()

	// Listener por sala
	for _, room := range rooms {
		go func(room string) {
			inQueue := inputQueues[room]
			for {
				msgBytes, _, err := inQueue.Receive()
				if err != nil {
					log.Printf("[%s] Error recibiendo mensaje: %v", room, err)
					time.Sleep(1 * time.Second)
					continue
				}

				var msg Message
				if err := json.Unmarshal(msgBytes, &msg); err != nil {
					log.Printf("[%s] Error parseando mensaje: %v", room, err)
					continue
				}

				switch msg.Type {
				case "register":
					outQueue, err := posix_mq.NewMessageQueue(msg.QueueName, posix_mq.O_WRONLY, 0666, nil)
					if err != nil {
						log.Printf("[%s] Error creando cola de salida %s: %v", room, msg.QueueName, err)
						continue
					}
					outputQueues[room][msg.ClientId] = outQueue
					log.Printf("[%s] Cliente %s registrado en %s", room, msg.ClientId, msg.QueueName)

				case "chat":
					log.Printf("[%s] Mensaje de %s: %s", room, msg.Name, msg.Text)

					// Enviar a todos los clientes de la sala
					for clientID, outQ := range outputQueues[room] {
						// Clonar el mensaje para cada cliente
						outMsg := msg
						// Opcional: evitar reenviar al mismo cliente si quieres
						// if clientID == msg.ClientId {
						//     continue
						// }

						data, _ := json.Marshal(outMsg)
						if err := outQ.Send(data, 0); err != nil {
							log.Printf("[%s] Error enviando a %s: %v", room, clientID, err)
						}
					}

				default:
					log.Printf("[%s] Tipo de mensaje desconocido: %s", room, msg.Type)
				}
			}
		}(room)
	}

	select {} // Mantener vivo
}
