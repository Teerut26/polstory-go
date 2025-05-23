package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/teerut26/polstory-go/route"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: No .env file found")
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 20, // 10mb
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://polstory.teerut.com, http://localhost:5173, http://teerut-server:3008",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Static("/", "./web")

	apiGroup := app.Group("/api")

	apiGroup.Post("/916/generate", route.Generate916Handler)
	apiGroup.Post("/45/generate", route.Generate45Handler)

	log.Fatal(app.Listen(":3000"))
	log.Println("Server is running on port 3000")
}
