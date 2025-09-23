package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/syucream/posix_mq/src/posix_mq"
)

var rooms = []string{"general", "deportes", "musica"}

type Message struct {
	ClientId string
	Name     string
	Text     string
	DateTime string
}

// Mapa de sala → *MessageQueue
var queues = make(map[string]*posix_mq.MessageQueue)

func initRooms() {
	for _, room := range rooms {
		name := "/" + room
		mq, err := posix_mq.NewMessageQueue(name, posix_mq.O_CREAT|posix_mq.O_RDWR, 0666, nil)
		if err != nil {
			log.Fatalf("Error al crear o abrir cola para sala %s: %v", room, err)
		}
		queues[room] = mq
		fmt.Printf("Sala creada/abierta: %s\n", room)
	}
}

func cleanup() {
	for _, room := range rooms {
		if q, ok := queues[room]; ok {
			q.Close()
			// unlink para eliminar cola
			err := q.Unlink()
			if err != nil {
				fmt.Printf("Error unlink sala %s: %v\n", room, err)
			} else {
				fmt.Printf("Sala unlink: %s\n", room)
			}
		}
	}
}

func main() {
	fmt.Println("Servidor iniciado con POSIX MQ")

	initRooms()

	// Capturar señales para limpiar al salir
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		fmt.Println("\nRecibida señal de terminación, limpiando...")
		cleanup()
		os.Exit(0)
	}()

	// Para cada sala, correr un listener que “rebota” los mensajes
	for _, room := range rooms {
		go func(room string) {
			q := queues[room]
			for {
				msg, _, err := q.Receive()
				if err != nil {
					log.Printf("Error recibiendo mensaje en sala %s: %v", room, err)
					time.Sleep(1 * time.Second)
					continue
				}
				text := string(msg)
				text = strings.TrimSpace(text)
				if text == "" {
					continue
				}
				// Mostrar en servidor
				fmt.Printf("[%s] Recibido: %s\n", room, text)

				// Reenviar el mensaje a la cola de la sala (broadcast)
				// Todos los clientes escuchan la misma cola de sala
				err2 := q.Send([]byte(text), 0)
				if err2 != nil {
					log.Printf("Error reenviando mensaje en sala %s: %v", room, err2)
				}
			}
		}(room)
	}

	// Mantener al servidor vivo
	select {}
}
