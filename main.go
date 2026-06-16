package main

import (
	"go-fiber-template/configuration"
	ds "go-fiber-template/domain/datasources"
	repo "go-fiber-template/domain/repositories"
	gw "go-fiber-template/src/gateways"
	"go-fiber-template/src/middlewares"
	sv "go-fiber-template/src/services"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {

	// // // remove this before deploy ###################
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// /// ############################################

	app := fiber.New(configuration.NewFiberConfiguration())
	middlewares.Logger(app)
	app.Use(recover.New())
	app.Use(cors.New())

	mongodb := ds.NewMongoDB(10)

	userRepo := repo.NewUsersRepository(mongodb)
	workspaceMembersRepo := repo.NewWorkspaceMembersRepository(mongodb)
	workspacesRepo := repo.NewWorkspacesRepository(mongodb)

	sv0 := sv.NewUsersService(userRepo)
	sv1 := sv.NewWorkspaceMembersService(workspaceMembersRepo)
	sv2 := sv.NewWorkspacesService(workspacesRepo)

	gw.NewHTTPGateway(app, sv0, sv1, sv2)

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "8080"
	}

	app.Listen(":" + PORT)
}