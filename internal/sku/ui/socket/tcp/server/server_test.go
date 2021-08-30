//+build unit

package server_test

import (
	"context"
	"feeder-service/internal/sku/application/command/create_sku"
	applicationMock "feeder-service/internal/sku/application/command/create_sku/mock"
	"feeder-service/internal/sku/ui/socket/tcp/server"
	"feeder-service/internal/sku/ui/socket/tcp/sku_reader/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"log"
	"os"
	"strings"
	"testing"
)

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
	s.skuReaderMock.EXPECT().Read().Times(4).Return("patata", nil)
	s.createSkuCommandHandlerMock.EXPECT().Handle(create_sku.Command{Sku: "patata"}).Times(4).Return(nil)
	s.skuReaderMock.EXPECT().Read().Times(1).Return("terminate", nil)
	s.server.Run(s.ctx, 5)
	s.Require().Equal("Received 4 unique product skus, 0 duplicates, 0 discard values\n", s.loggerBuffer.String())
}

func (s *UnitSuite) TestSkuReaderReadIsNotCalledWhenMaxConnectionsIsZero() {
	s.skuReaderMock.EXPECT().Read().Times(0)
	s.createSkuCommandHandlerMock.EXPECT().Handle(gomock.Any()).Times(0)
	s.server.Run(s.ctx, 0)
}

func (s *UnitSuite) TestServerFinishAndAReportIsGeneratedWhenContextIsDone() {
	s.skuReaderMock.EXPECT().Read().Return("patata", nil).AnyTimes()
	s.createSkuCommandHandlerMock.EXPECT().Handle(gomock.Any()).AnyTimes().Return(nil)
	go s.server.Run(s.ctx, 5)
	proc, err := os.FindProcess(os.Getpid())
	s.Require().NoError(err)
	err = proc.Signal(os.Interrupt)
	s.Require().NoError(err)
	s.Require().Contains(s.loggerBuffer.String(), "Received")
}