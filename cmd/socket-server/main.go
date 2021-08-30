package main

import (
	"context"
	"feeder-service/internal/sku/application/command/create_sku"
	"feeder-service/internal/sku/infrastructure/io/socket/tcp/server"
	"feeder-service/internal/sku/infrastructure/io/socket/tcp/sku_reader"
	"feeder-service/internal/sku/infrastructure/persistence/mongo"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Println("Starting server")

	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()

	bootstrapTCPServer().Run(ctx, 5)
}

func bootstrapTCPServer() *server.Server {
	skuReader := sku_reader.New("localhost:4000")
	skuRepository := mongo.NewSkuRepository()
	createSkuCommandHandler := create_sku.NewCommandHandler(skuRepository)

	logFile, err := os.OpenFile("server_report_file", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.Lmsgprefix)

	return server.New(skuReader, createSkuCommandHandler, logger)
}

