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
	ClientId string
	Name     string
	Text     string
	DateTime string
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

	qname := "/" + room

	// Abrir la cola de la sala, con lectura y escritura para recibir y enviar
	q, err := posix_mq.NewMessageQueue(qname, posix_mq.O_RDWR, 0666, nil)
	if err != nil {
		log.Fatalf("No se pudo abrir la sala %s: %v", room, err)
	}
	defer q.Close()

	fmt.Printf("Conectado a la sala [%s] como [%s]\n", room, name)

	// Hilo para recibir mensajes
	go func() {
		for {
			msgBytes, _, err := q.Receive()
			if err != nil {
				log.Println("Error al recibir mensaje:", err)
				continue
			}

			var receivedMsg Message
			err = json.Unmarshal(msgBytes, &receivedMsg)

			if err != nil {
				log.Println("Error deserializando mensaje:", err)
				continue
			}

			if receivedMsg.ClientId == clientId {
				continue
			}

			fmt.Println("\n" + receivedMsg.DateTime + " - " + receivedMsg.Name + ": " + receivedMsg.Text)
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
			ClientId: clientId,
			Name:     name,
			Text:     input,
			DateTime: time.Now().UTC().String(),
		}

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			log.Println("Error serializando el mensaje:", err)
			continue
		}

		err = q.Send(msgBytes, 0)
		if err != nil {
			log.Println("Error enviando el mensaje:", err)
		}
	}
}
