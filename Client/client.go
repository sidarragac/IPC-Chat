package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/syucream/posix_mq/src/posix_mq"
)

type Message struct {
	Type      string // "register" o "chat"
	ClientId  string
	Name      string
	Text      string
	DateTime  string
	QueueName string // solo en mensajes de registro
}

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
)

func colorForName(name string) string {
	// Asignar un color basado en hash del nombre para consistencia
	colors := []string{Red, Green, Yellow, Blue, Purple, Cyan}
	hash := 0
	for _, c := range name {
		hash += int(c)
	}
	return colors[hash%len(colors)]
}


func main() {
	reader := bufio.NewReader(os.Stdin)
	clientId := uuid.New().String()

	fmt.Print("Ingresa tu nombre: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error leyendo nombre:", err)
	}
	name = strings.TrimSpace(name)

	fmt.Print("Elige una sala (general, deportes, musica): ")
	room, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error leyendo sala:", err)
	}
	room = strings.TrimSpace(room)

	if room == "" {
		log.Fatal("Sala no válida")
	}

	inQueueName := "/" + room + "_in"
	recvQueueName := "/" + room + "_out_" + clientId

	// Cola de envío (al servidor)
	sendQueue, err := posix_mq.NewMessageQueue(inQueueName, posix_mq.O_WRONLY, 0666, nil)
	if err != nil {
		log.Fatalf("No se pudo abrir la cola de envío %s: %v", inQueueName, err)
	}
	defer sendQueue.Close()

	// Crear cola personal de recepción
	recvQueue, err := posix_mq.NewMessageQueue(recvQueueName, posix_mq.O_CREAT|posix_mq.O_RDONLY, 0666, nil)
	if err != nil {
		log.Fatalf("No se pudo crear la cola de recepción %s: %v", recvQueueName, err)
	}
	defer func() {
		recvQueue.Close()
		recvQueue.Unlink()
	}()

	// Enviar mensaje de registro al servidor
	regMsg := Message{
		Type:      "register",
		ClientId:  clientId,
		Name:      name,
		QueueName: recvQueueName,
	}

	msgBytes, _ := json.Marshal(regMsg)
	if err := sendQueue.Send(msgBytes, 0); err != nil {
		log.Fatalf("No se pudo enviar mensaje de registro: %v", err)
	}

	fmt.Printf("Conectado a la sala [%s] como [%s]\n", room, name)

	// Escuchar mensajes
	go func() {
		for {
			msgBytes, _, err := recvQueue.Receive()
			if err != nil {
				log.Println("Error al recibir mensaje:", err)
				continue
			}

			var msg Message
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Error deserializando mensaje:", err)
				continue
			}

			if msg.ClientId == clientId {
        continue
      }

			color := colorForName(msg.Name)
			fmt.Print("\n")
			fmt.Printf("%s%s%s - %s%s%s: %s\n", Cyan, msg.DateTime, Reset, color, msg.Name, Reset, msg.Text)
			fmt.Print("> ")
		}
	}()

	// Enviar mensajes
	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error leyendo entrada:", err)
			continue
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		msg := Message{
			Type:     "chat",
			ClientId: clientId,
			Name:     name,
			Text:     input,
			DateTime: time.Now().Format("2006-01-02 15:04:05"),
		}

		msgBytes, _ := json.Marshal(msg)
		if err := sendQueue.Send(msgBytes, 0); err != nil {
			log.Println("Error enviando mensaje:", err)
		}
	}
}
