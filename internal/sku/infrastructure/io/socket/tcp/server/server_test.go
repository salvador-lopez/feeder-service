//+build unit

package server_test

import (
	"context"
	"feeder-service/internal/sku/application/command/create_sku"
	applicationMock "feeder-service/internal/sku/application/command/create_sku/mock"
	"feeder-service/internal/sku/infrastructure/io/socket/tcp/server"
	"feeder-service/internal/sku/infrastructure/io/socket/tcp/sku_reader/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"log"
	"strings"
	"sync"
	"testing"
)

const sku = "KASL-3423"

type UnitSuite struct {
	suite.Suite
	ctx                         context.Context
	skuReaderMock               *mock.MockSkuReader
	createSkuCommandHandlerMock *applicationMock.MockCommandHandlerInterface
	mockCtrl                    *gomock.Controller
	logger                      *log.Logger
	loggerBuffer				*strings.Builder
	server                      *server.Server
}

func (s *UnitSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.skuReaderMock = mock.NewMockSkuReader(s.mockCtrl)
	s.createSkuCommandHandlerMock = applicationMock.NewMockCommandHandlerInterface(s.mockCtrl)
	s.loggerBuffer = &strings.Builder{}
	s.logger = log.New(s.loggerBuffer, "", log.Lmsgprefix)
	s.server = server.New(s.skuReaderMock, s.createSkuCommandHandlerMock, s.logger)
}

func (s *UnitSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(UnitSuite))
}

func (s *UnitSuite) TestSkuReaderReadIsCalledFiveTimesAndReportIsLoggedAsExpected() {
	s.skuReaderMock.EXPECT().Read().Times(4).Return(sku, nil)
	s.skuReaderMock.EXPECT().Read().AnyTimes().Return("terminate", nil)

	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: sku}).Return(nil).Times(2)
	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: sku}).Return(create_sku.ErrSkuAlreadyExists).Times(1)
	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: sku}).Return(create_sku.ErrCreatingSku).Times(1)

	s.server.Run(s.ctx, 1)
	s.Require().Equal("Received 2 unique product skus, 1 duplicates, 1 discard values\n", s.loggerBuffer.String())
}

func (s *UnitSuite) TestSkuReaderReadIsNotCalledWhenMaxConnectionsIsZero() {
	s.skuReaderMock.EXPECT().Read().Times(0)
	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, gomock.Any()).Times(0)
	s.server.Run(s.ctx, 0)
}

func (s *UnitSuite) TestServerFinishAndAReportIsGeneratedWhenContextIsDoneDueToCancel() {
	s.skuReaderMock.EXPECT().Read().AnyTimes().Return("terminate", nil)
	ctx, cancelFunc := context.WithCancel(s.ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.server.Run(ctx, 5)
		wg.Done()
	}()
	wg.Wait()
	cancelFunc()
	s.Require().NotEmpty(s.loggerBuffer.String())
}

func (s *UnitSuite) TestServerFinishAndAReportIsGeneratedWhenContextIsDoneDueToTimeout() {
	s.skuReaderMock.EXPECT().Read().AnyTimes().Return(sku, nil)
	s.createSkuCommandHandlerMock.EXPECT().Handle(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	ctx, cancelFunc := context.WithTimeout(s.ctx, 0)
	defer cancelFunc()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.server.Run(ctx, 5)
		wg.Done()
	}()
	wg.Wait()

	s.Require().NotEmpty(s.loggerBuffer.String())
}