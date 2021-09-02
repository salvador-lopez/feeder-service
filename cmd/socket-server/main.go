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
	"time"
)

const (
	socketAddr     = "localhost:4000"
	logFileName    = "server_report_file.txt"
	maxConnections = 5
	timeout        = 60 * time.Second
)

func main() {
	fmt.Println("Starting server")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	serverTCP, err := bootstrapApplication(ctx)
	if err != nil {
		log.Fatalf("error bootstraping application: %v", err)
	}
	serverTCP.Run(ctx, maxConnections, time.Now().Add(timeout))
}

func bootstrapApplication(ctx context.Context) (*server.Server, error) {
	skuReader, err := sku_reader.New(socketAddr)
	if err != nil {
		return nil, err
	}

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}
	err = mongoClient.Connect(ctx)
	if err != nil {
		return nil, err
	}

	db := mongoClient.Database("sku")
	skuRepository, err := mongoSku.NewSkuRepository(db, domain.NewHydrator())
	if err != nil {
		return nil, err
	}

	createSkuCommandHandler := create_sku.NewCommandHandler(skuRepository)

	logFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	logger := log.New(logFile, "", log.Lmsgprefix)

	return server.New(skuReader, createSkuCommandHandler, logger), nil
}

