package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var connections = make(map[string]*websocket.Conn)

func main() {
	app := fiber.New()

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

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		log.Println(c.Locals("allowed"))
		log.Println(c.Params("id"))
		log.Println(c.Query("v"))
		log.Println(c.Cookies("session"))
		connections[c.Params("id")] = c
		var (
			mt  int
			msg []byte
			err error
		)
		for {
			fmt.Println(connections)
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				delete(connections, c.Params("id"))
				break
			}
			log.Printf("recv: %s", msg)

			for _, connection := range connections {
				if err = connection.WriteMessage(mt, msg); err != nil {
					log.Println("write:", err)
					delete(connections, c.Params("id"))
				}
			}
		}

	}))
	log.Fatal(app.Listen(":8000"))
}
