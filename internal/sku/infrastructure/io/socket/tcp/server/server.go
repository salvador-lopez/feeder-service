package server

import (
	"context"
	"errors"
	"feeder-service/internal/sku/application/command/create_sku"
	"feeder-service/internal/sku/domain"
	"feeder-service/internal/sku/infrastructure/io/socket/tcp/sku_reader"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	skuReader               sku_reader.SkuReader
	createSkuCommandHandler create_sku.CommandHandlerInterface
	logger                  *log.Logger
}

func New(skuReader sku_reader.SkuReader, createSkuCommandHandler create_sku.CommandHandlerInterface, logger *log.Logger) *Server {
	return &Server{skuReader: skuReader, createSkuCommandHandler: createSkuCommandHandler, logger: logger}
}

func (s *Server) Run(ctx context.Context, maxConnections int, deadline time.Time) {
	liveCondition := true
	if maxConnections <= 0 {
		liveCondition = false
	}

	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
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
	connectionSlots := NewConnectionSlotStatus(maxConnections)
	var wg sync.WaitGroup
	for liveCondition {
		if connectionSlots.UseFreeSlot() {
			wg.Add(1)
			go func() {
				message, err := s.skuReader.Read(deadline)
				if err != nil {
					liveCondition = false
					wg.Done()
					return
				}
				if message == "terminate" {
					liveCondition = false
					wg.Done()
					return
				}
				err = s.createSkuCommandHandler.Handle(ctx, create_sku.Command{Sku: message})
				if err != nil {
					if errors.Is(err, domain.ErrSkuAlreadyExists) {
						duplicatedSkus++
						connectionSlots.FreesASlot()
						wg.Done()
						return
					}
					invalidSkus++
					connectionSlots.FreesASlot()
					wg.Done()
					return
				}
				createdSkus++
				connectionSlots.FreesASlot()
				wg.Done()
			}()
		}
	}
	wg.Wait()

	s.logger.Println("Received "+strconv.Itoa(createdSkus)+" unique product skus, "+strconv.Itoa(duplicatedSkus)+" duplicates, "+strconv.Itoa(invalidSkus)+" discard values")
}
