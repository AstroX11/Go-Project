package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	// "time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func main() {
	// Set up logging
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	clientLog := waLog.Stdout("Client", "DEBUG", true)

	// Initialize the SQLite database
	container, err := sqlstore.New("sqlite3", "file:whatsmeow.db?_foreign_keys=on", dbLog)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	// Get the first device (or create a new one)
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		log.Fatalf("Failed to get device: %v", err)
	}

	// Create the WhatsApp client
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Add an event handler to listen for incoming messages
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			fmt.Printf("Received a message from %s: %s\n", v.Info.Sender.String(), v.Message.GetConversation())
		}
	})

	// Connect to WhatsApp
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Wait for a QR code to be scanned
	qrChan, _ := client.GetQRChannel(context.Background())
	err = client.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	for evt := range qrChan {
		if evt.Event == "code" {
			fmt.Println("Scan the QR code with your phone:")
			fmt.Println(evt.Code) // Display the QR code in the terminal
		} else {
			fmt.Println("Logged in!")
			break
		}
	}

	// Send a test message after logging in
	// go func() {
	// 	time.Sleep(3 * time.Second) // Wait for a few seconds to ensure the client is ready
	// 	_, err := client.SendMessage(context.Background(), "1234567890@s.whatsapp.net", &whatsmeow.TextMessage{
	// 		Content: "Hello, this is a test message from Whatsmeow!",
	// 	})
	// 	if err != nil {
	// 		log.Printf("Failed to send message: %v", err)
	// 	} else {
	// 		log.Println("Message sent successfully!")
	// 	}
	// }()

	// Wait for an interrupt signal to disconnect
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Disconnect the client
	client.Disconnect()
	log.Println("Client disconnected.")
}