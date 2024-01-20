package environment

import (
	"encoding/json"
	environment2 "github.com/GSH-LAN/Unwindia_common/src/go/environment"
	"github.com/GSH-LAN/Unwindia_common/src/go/logger"
	"github.com/GSH-LAN/Unwindia_common/src/go/messagebroker"
	pulsarClient "github.com/apache/pulsar-client-go/pulsar"
	envLoader "github.com/caarlos0/env/v10"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/ksuid"
	"runtime"
	"time"
)

var (
	env *Environment
)

// environment holds all the environment variables with primitive types
type environment struct {
	environment2.BaseEnvironment

	ServiceUid          string
	DotlanMySQLHost     string `env:"MYSQL_HOST"`
	DotlanMySQLPort     int    `env:"MYSQL_PORT" envDefault:"3306"`
	DotlanMySQLUser     string `env:"MYSQL_USER"`
	DotlanMySQLPassword string `env:"MYSQL_PASSWORD"`
	DotlanMySQLDatabase string `env:"MYSQL_DATABASE"`
	// TODO: add background process which deletes finished contests+Tournaments older than configured PreserveHours
	PreserveHours int `env:"PRESERVE_HOURS" envDefault:"168" envDescription:"Number of hours to preserve processed data in the database"`

	ProcessInterval time.Duration `env:"PROCESS_INTERVAL" envDefault:"10s"`
}

// Environment holds all environment configuration with more advanced typing and validation
type Environment struct {
	environment
	PulsarAuth pulsarClient.Authentication
}

// Load initialized the environment variables
func load() *Environment {
	e := environment{}
	if err := envLoader.Parse(&e); err != nil {
		log.Panic().Err(err)
	}

	if err := logger.SetLogLevel(e.LogLevel); err != nil {
		log.Panic().Err(err)
	}

	if e.WorkerCount <= 0 {
		e.WorkerCount = runtime.NumCPU() + e.WorkerCount
	}

	if e.ServiceUid == "" {
		e.ServiceUid = ksuid.New().String()
	}

	var pulsarAuthParams = make(map[string]string)
	if e.PulsarAuthParams != "" {
		if err := json.Unmarshal([]byte(e.PulsarAuthParams), &pulsarAuthParams); err != nil {
			log.Panic().Err(err)
		}
	}

	var mbpulsarAuth messagebroker.PulsarAuth
	if err := mbpulsarAuth.Unmarshal(e.PulsarAuth); err != nil {
		log.Panic().Err(err)
	}

	var pulsarAuth pulsarClient.Authentication

	switch mbpulsarAuth {
	case messagebroker.AUTH_TOKEN:
		pulsarAuth = pulsarClient.NewAuthenticationToken(pulsarAuthParams["token"])
	case messagebroker.AUTH_OAUTH2:
		pulsarAuth = pulsarClient.NewAuthenticationOAuth2(pulsarAuthParams)
	}

	e2 := Environment{
		environment: e,
		PulsarAuth:  pulsarAuth,
	}

	log.Info().Interface("environemt", e2).Msgf("Loaded Environment")

	return &e2
}

func Get() *Environment {
	if env == nil {
		env = load()
	}

	return env
}
