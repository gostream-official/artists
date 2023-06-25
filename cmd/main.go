package main

import (
	"fmt"
	"strconv"

	"github.com/gostream-official/artists/impl/funcs/createartist"
	"github.com/gostream-official/artists/impl/funcs/deleteartist"
	"github.com/gostream-official/artists/impl/funcs/getartist"
	"github.com/gostream-official/artists/impl/funcs/getartists"
	"github.com/gostream-official/artists/impl/funcs/updateartist"
	"github.com/gostream-official/artists/impl/inject"
	"github.com/gostream-official/artists/pkg/env"
	"github.com/gostream-official/artists/pkg/router"
	"github.com/gostream-official/artists/pkg/store"

	"github.com/revx-official/output/log"
)

// Description:
//
//	The package initializer function.
//	Initializes the log level to info.
func init() {
	log.Level = log.LevelInfo
}

// Description:
//
//	The main function.
//	Represents the entry point of the application.
func main() {
	log.Infof("booting service instance ...")

	executionPortEnvVar := env.GetEnvironmentVariableWithFallback("PORT", "9871")
	executionPort, err := strconv.Atoi(executionPortEnvVar)

	if err != nil {
		log.Fatalf("Received invalid execution port")
	}

	if executionPort < 0 || executionPort > 65535 {
		log.Fatalf("Received invalid execution port")
	}

	mongoUsername, err := env.GetEnvironmentVariable("MONGO_USERNAME")
	if err != nil {
		log.Fatalf("Cannot retrieve mongo username via environment variable")
	}

	mongoPassword, err := env.GetEnvironmentVariable("MONGO_PASSWORD")
	if err != nil {
		log.Fatalf("Cannot retrieve mongo password via environment variable")
	}

	mongoHost := env.GetEnvironmentVariableWithFallback("MONGO_HOST", "127.0.0.1:27017")

	connectionURI := fmt.Sprintf("mongodb://%s:%s@%s", mongoUsername, mongoPassword, mongoHost)
	instance, err := store.NewMongoInstance(connectionURI)

	log.Infof("establishing database connection ...")
	if err != nil {
		log.Fatalf("failed to connect to mongo instance: %s", err)
	}

	log.Infof("successfully established database connection")

	injector := inject.Injector{
		MongoInstance: instance,
	}

	log.Infof("launching router engine ...")
	engine := router.Default()

	engine.HandleWith("GET", "/artists", getartists.Handler).Inject(injector)
	engine.HandleWith("GET", "/artists/:id", getartist.Handler).Inject(injector)
	engine.HandleWith("POST", "/artists", createartist.Handler).Inject(injector)
	engine.HandleWith("PUT", "/artists/:id", updateartist.Handler).Inject(injector)
	engine.HandleWith("DELETE", "/artists/:id", deleteartist.Handler).Inject(injector)

	err = engine.Run(uint16(executionPort))
	if err != nil {
		log.Fatalf("failed to launch router engine: %s", err)
	}
}
