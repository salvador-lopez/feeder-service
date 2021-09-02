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
	"sync"
	"syscall"
	"time"
)

type Server struct {
	skuReader               sku_reader.SkuReader
	createSkuCommandHandler create_sku.CommandHandlerInterface
	logger                  *log.Logger
}

type Report struct {
	CreatedSkus int
	DuplicatedSkus int
	InvalidSkus int
}

func New(skuReader sku_reader.SkuReader, createSkuCommandHandler create_sku.CommandHandlerInterface, logger *log.Logger) *Server {
	return &Server{skuReader: skuReader, createSkuCommandHandler: createSkuCommandHandler, logger: logger}
}

func (s *Server) Run(ctx context.Context, maxConnections int, deadline time.Time) Report {
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

	report := Report{}
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
						report.DuplicatedSkus++
						connectionSlots.FreesASlot()
						wg.Done()
						return
					}
					report.InvalidSkus++
					connectionSlots.FreesASlot()
					wg.Done()
					return
				}
				report.CreatedSkus++
				s.logger.Println(message)
				connectionSlots.FreesASlot()
				wg.Done()
			}()
		}
	}
	wg.Wait()

	return report
}
