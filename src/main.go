package main

import (
	"context"
	"github.com/GSH-LAN/Unwindia_common/src/go/workitemLock"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-pulsar/pkg/pulsar"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/GSH-LAN/Unwindia_MS_dotlan/src/database"
	"github.com/GSH-LAN/Unwindia_MS_dotlan/src/dotlan"
	"github.com/GSH-LAN/Unwindia_MS_dotlan/src/environment"
	"github.com/GSH-LAN/Unwindia_MS_dotlan/src/worker"
	"github.com/GSH-LAN/Unwindia_common/src/go/config"
	"github.com/gammazero/workerpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	mainContext, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		os.Exit(1)
	}()

	err := godotenv.Load()
	if err != nil && !strings.Contains(err.Error(), "no such file") {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	env := environment.Get()

	var configClient config.ConfigClient
	if env.ConfigFileName != "" {
		configClient, err = config.NewConfigFile(mainContext, env.ConfigFileName, env.ConfigTemplatesDir)
	} else {
		configClient, err = config.NewConfigClient()
	}

	if err != nil {
		cancel()
		log.Fatal().Err(err).Msg("Error initializing config")
	}

	wp := workerpool.New(env.WorkerCount)

	dotlanClient, err := dotlan.NewClient(mainContext, env, wp, configClient)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating dotlan client")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(env.MongoDbURI))
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating mongo client")
	}

	ctx, _ := context.WithTimeout(mainContext, 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to mongodb")
	}
	db := client.Database(database.DATABASE_NAME)
	dbClient, err := database.NewClientWithDatabase(db, wp)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating database client")
	}

	publisher, err := pulsar.NewPublisher(
		pulsar.PublisherConfig{
			URL:            env.PulsarURL,
			Authentication: env.PulsarAuth,
		},
		watermill.NewStdLoggerWithOut(log.Logger, true, false),
	)

	lock := workitemLock.NewMemoryWorkItemLock()

	w := worker.NewWorker(mainContext, wp, configClient, dotlanClient, dbClient, publisher, lock)

	err = w.Start(time.NewTicker(env.ProcessInterval))
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting worker")
	}
}
