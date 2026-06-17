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
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// // // remove this before deploy ###################
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}
	// /// ############################################

	// Set Timezone to Asia/Bangkok
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err == nil {
		time.Local = loc
	}

	app := fiber.New(configuration.NewFiberConfiguration())
	middlewares.Logger(app)
	app.Use(recover.New())
	app.Use(cors.New())

	mongodb := ds.NewMongoDB(10)

	debtorRepo := repo.NewDebtorsRepository(mongodb)
	callListItemRepo := repo.NewCallListItemsRepository(mongodb)
	callAttemptRepo := repo.NewCallAttemptsRepository(mongodb)
	callSessionRepo := repo.NewCallSessionsRepository(mongodb)
	callRecordsRepo := repo.NewCallRecordsRepository(mongodb)
	workspacesRepo := repo.NewWorkspacesRepository(mongodb)

	sv1 := sv.NewDebtorsService(debtorRepo)
	sv2 := sv.NewCallListItemsService(callListItemRepo)
	sv3 := sv.NewCallAttemptsService(callAttemptRepo, callListItemRepo)
	sv4 := sv.NewCallSessionsService(callSessionRepo)
	callRecordsSv := sv.NewCallRecordsService(callRecordsRepo)
	sv6 := sv.NewWorkspacesService(workspacesRepo)
	voicebotMakeCallSv := sv.NewVoicebotMakeCallService()

	gw.NewHTTPGateway(app, sv6, callRecordsSv, sv1, sv2, sv3, sv4, voicebotMakeCallSv)

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "8080"
	}

	app.Listen(":" + PORT)
}
