package server

import (
	"context"
	"errors"
	"feeder-service/internal/sku/application/command/create_sku"
	"feeder-service/internal/sku/ui/socket/tcp/sku_reader"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type Server struct {
	skuReader sku_reader.SkuReader
	createSkuCommandHandler create_sku.CommandHandlerInterface
	logger *log.Logger
}

func New(skuReader sku_reader.SkuReader, createSkuCommandHandler create_sku.CommandHandlerInterface, logger *log.Logger) *Server {
	return &Server{skuReader: skuReader, createSkuCommandHandler: createSkuCommandHandler, logger: logger}
}

func (s *Server) Run(ctx context.Context, maxConnections int) {
	liveCondition := true
	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigChannel:
			liveCondition = false
		case <-ctx.Done():
			liveCondition = false
		}
	}()

	createdSkus := 0
	duplicatedSkus := 0
	invalidSkus := 0
	connectionSlots := NewConnectionSlots(maxConnections)

	if !connectionSlots.HasFreeSlot() {
		liveCondition = false
	}

	for liveCondition {
		if connectionSlots.HasFreeSlot() {
			connectionSlots.UseFreeSlot()
			go func() {
				message, _ := s.skuReader.Read()
				if message == "terminate" {
					liveCondition = false
					return
				}
				err := s.createSkuCommandHandler.Handle(create_sku.Command{Sku: message})
				if err != nil {
					if errors.Is(err, create_sku.ErrSkuAlreadyExists) {
						duplicatedSkus++
						return
					}
					invalidSkus++
				}
				createdSkus++
			}()
		}
	}

	s.logger.Println("Received "+strconv.Itoa(createdSkus)+" unique product skus, "+strconv.Itoa(duplicatedSkus)+" duplicates, "+strconv.Itoa(invalidSkus)+" discard values")
}