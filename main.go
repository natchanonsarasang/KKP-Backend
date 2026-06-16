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
	debtorRepo := repo.NewDebtorsRepository(mongodb)
	callListItemRepo := repo.NewCallListItemsRepository(mongodb)
	callAttemptRepo := repo.NewCallAttemptsRepository(mongodb)
	callSessionRepo := repo.NewCallSessionsRepository(mongodb)
	callRecordsRepo := repo.NewCallRecordsRepository(mongodb)

	sv0 := sv.NewUsersService(userRepo)
	sv1 := sv.NewDebtorsService(debtorRepo, sv0)
	sv2 := sv.NewCallListItemsService(callListItemRepo, sv0)
	sv3 := sv.NewCallAttemptsService(callAttemptRepo, callListItemRepo, sv0)
	sv4 := sv.NewCallSessionsService(callSessionRepo)
	callRecordsSv := sv.NewCallRecordsService(callRecordsRepo)

	gw.NewHTTPGateway(app, sv0, callRecordsSv, sv1, sv2, sv3, sv4)

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "8080"
	}

	app.Listen(":" + PORT)

}
