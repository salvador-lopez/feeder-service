package main

import (
	"context"
	"feeder-service/internal/sku/application/command/create_sku"
	"feeder-service/internal/sku/domain"
	"feeder-service/internal/sku/infrastructure/io/socket/tcp/server"
	"feeder-service/internal/sku/infrastructure/io/socket/tcp/sku_reader"
	mongoSku "feeder-service/internal/sku/infrastructure/persistence/mongo"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"strconv"
	"time"
)

type config struct {
	socketAddr  string
	mongoUri string
	mongoDatabase string
	logFileName string
	maxConcurrentConnections int
	timeout time.Duration
}

func newConfigDefault() *config {
	return &config{
		socketAddr:               "localhost:4000",
		mongoUri:                 "mongodb://localhost:27017",
		mongoDatabase:            "sku",
		logFileName:              "server_report_file.txt",
		maxConcurrentConnections: 5,
		timeout:                  60 * time.Second,
	}
}

func main() {
	cfg, err := fetchConfigFromEnvVars()
	if err != nil {
		log.Fatalf("error fetching application config from env vars: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	serverTCP, err := bootstrapApplication(ctx, cfg)
	if err != nil {
		log.Fatalf("error bootstraping application: %v", err)
	}
	fmt.Println("Starting listening tcp connections in "+cfg.socketAddr)
	serverTCP.Run(ctx, cfg.maxConcurrentConnections, time.Now().Add(cfg.timeout))
	fmt.Println("Server finished successfully")
}

func fetchConfigFromEnvVars() (*config, error) {
	cfg := newConfigDefault()
	fetchSocketAddrEnvVar(cfg)
	fetchLogFileNameEnvVar(cfg)
	err := fetchMaxConcurrentConnectionsEnvVar(cfg)
	if err != nil {
		return nil, err
	}

	err = fetchTimeoutEnvVar(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func fetchSocketAddrEnvVar(cfg *config) {
	socketAddr, ok := os.LookupEnv("SOCKET_ADDR")
	if ok {
		cfg.socketAddr = socketAddr
	}
}

func fetchLogFileNameEnvVar(cfg *config) {
	logFileName, ok := os.LookupEnv("LOG_FILE_NAME")
	if ok {
		cfg.logFileName = logFileName
	}
}

func fetchMaxConcurrentConnectionsEnvVar(cfg *config) error{
	maxConcurrentConnsAsString, ok := os.LookupEnv("MAX_CONCURRENT_CONNECTIONS")
	if ok {
		maxConcurrentConns, err := strconv.Atoi(maxConcurrentConnsAsString)
		if err != nil {
			return err
		}
		cfg.maxConcurrentConnections = maxConcurrentConns
	}

	return nil
}

func fetchTimeoutEnvVar(cfg *config) error{
	timeoutAsString, ok := os.LookupEnv("TIMEOUT_IN_SECS")
	if ok {
		timeout, err := strconv.Atoi(timeoutAsString)
		if err != nil {
			return err
		}
		cfg.timeout = time.Duration(timeout) * time.Second
	}

	return nil
}

func bootstrapApplication(ctx context.Context, cfg *config) (*server.Server, error) {
	skuReader, err := sku_reader.New(cfg.socketAddr)
	if err != nil {
		return nil, err
	}

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(cfg.mongoUri))
	if err != nil {
		return nil, err
	}
	err = mongoClient.Connect(ctx)
	if err != nil {
		return nil, err
	}

	db := mongoClient.Database(cfg.mongoDatabase)
	skuRepository, err := mongoSku.NewSkuRepository(db, domain.NewHydrator())
	if err != nil {
		return nil, err
	}

	createSkuCommandHandler := create_sku.NewCommandHandler(skuRepository)

	logFile, err := os.OpenFile(cfg.logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	logger := log.New(logFile, "", log.Lmsgprefix)

	return server.New(skuReader, createSkuCommandHandler, logger), nil
}
