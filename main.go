package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
)

var clients = make(map[*websocket.Conn]struct{})
var save = make(chan *websocket.Conn)
var del = make(chan *websocket.Conn)
var data = make(chan []byte)

func hub() {
	for {
		select {
		case con := <-save:
			clients[con] = struct{}{}

		case con := <-del:
			delete(clients, con)

		case message := <-data:
			for con := range clients {
				if err := con.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					log.Println("err : ", err)

					con.Close()
					delete(clients, con)
				}
			}
		}
	}
}

func main() {
	app := fiber.New()

	app.Use(cors.New())

	if err := godotenv.Load(); err != nil {
		log.Fatal("Could not load env file")
	}

	app.Static("/", "dist")

	app.Get("/", func(c *fiber.Ctx) error {
		return nil
	})
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	go hub()

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		log.Println(c.Locals("allowed"))
		log.Println(c.Params("id"))
		log.Println(c.Query("v"))
		log.Println(c.Cookies("session"))

		defer func() {
			del <- c
			c.Close()
		}()

		save <- c
		for {
			fmt.Println("c => ", c)
			messageType, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("err : ", err)
				return
			}
			log.Printf("recv: %s", msg)
			if messageType == websocket.TextMessage {
				data <- msg
			}

		}

	}))
	log.Fatal(app.Listen(os.Getenv("port")))
}
