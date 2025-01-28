package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/teerut26/polstory-go/route"
)

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 10, // 10mb
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://polstory.teerut.com, http://localhost:5173, http://teerut-server:3008",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Static("/", "./web")

	apiGroup := app.Group("/api")
	apiGroup.Post("/generate", route.GenerateHandler)

	log.Fatal(app.Listen(":3000"))
	log.Println("Server is running on port 3000")
}
