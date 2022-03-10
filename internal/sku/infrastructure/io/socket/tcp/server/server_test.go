//+build unit

package server_test

import (
	"context"
	"feeder-service/internal/sku/application/command/create_sku"
	applicationMock "feeder-service/internal/sku/application/command/create_sku/mock"
	"feeder-service/internal/sku/domain"
	"feeder-service/internal/sku/infrastructure/io/socket/tcp/server"
	"feeder-service/internal/sku/infrastructure/io/socket/tcp/sku_reader/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"log"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	sku            = "KASL-3423"
	anotherSku     = "SLOS-4332"
	maxConnections = 5
)

type UnitSuite struct {
	suite.Suite
	ctx                         context.Context
	skuReaderMock               *mock.MockSkuReader
	createSkuCommandHandlerMock *applicationMock.MockCommandHandlerInterface
	mockCtrl                    *gomock.Controller
	logger                      *log.Logger
	loggerBuffer				*strings.Builder
	deadline					time.Time
	server                      *server.Server
}

func (s *UnitSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.skuReaderMock = mock.NewMockSkuReader(s.mockCtrl)
	s.createSkuCommandHandlerMock = applicationMock.NewMockCommandHandlerInterface(s.mockCtrl)
	s.loggerBuffer = &strings.Builder{}
	s.logger = log.New(s.loggerBuffer, "", log.Lmsgprefix)
	s.deadline = time.Now().Add(10 * time.Second)
	s.server = server.New(s.skuReaderMock, s.createSkuCommandHandlerMock, s.logger)
}

func (s *UnitSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(UnitSuite))
}

func (s *UnitSuite) TestSkuReaderReadIsCalledFiveTimesAndCreatedSkusAreLoggedAndReportIsReturned() {
	s.skuReaderMock.EXPECT().Read(s.deadline).Times(1).Return(sku, nil)
	s.skuReaderMock.EXPECT().Read(s.deadline).Times(3).Return(anotherSku, nil)
	s.skuReaderMock.EXPECT().Read(s.deadline).AnyTimes().Return("terminate", nil)

	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: sku}).Return(nil).Times(1)
	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: anotherSku}).Return(nil).Times(1)
	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: anotherSku}).Return(domain.ErrSkuAlreadyExists).Times(2)

	report := s.server.Run(s.ctx, maxConnections, s.deadline)
	s.Require().Equal(2, report.CreatedSkus)
	s.Require().Equal(2, report.DuplicatedSkus)
	s.Require().Equal(0, report.InvalidSkus)
	s.Require().Contains(s.loggerBuffer.String(), sku)
	s.Require().Contains(s.loggerBuffer.String(), anotherSku)
}

func (s *UnitSuite) TestDuplicatedSkusCanBeUpdatedInAConcurrentWayWithNoRaceConditions() {
	s.skuReaderMock.EXPECT().Read(s.deadline).Times(1).Return(sku, nil)
	s.skuReaderMock.EXPECT().Read(s.deadline).Times(15000).Return(anotherSku, nil)
	s.skuReaderMock.EXPECT().Read(s.deadline).AnyTimes().Return("terminate", nil)

	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: sku}).Return(nil).Times(1)
	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: anotherSku}).Return(nil).Times(1)
	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: anotherSku}).Return(domain.ErrSkuAlreadyExists).AnyTimes()

	report := s.server.Run(s.ctx, 500, s.deadline)
	s.Require().Equal(2, report.CreatedSkus)
	s.Require().Equal(14999, report.DuplicatedSkus)
	s.Require().Equal(0, report.InvalidSkus)
}

func (s *UnitSuite) TestCreatedSkusCanBeUpdatedInAConcurrentWayWithNoRaceConditions() {
	for i := 0; i < 10000; i++ {
		randomSku := "KASL-"+strconv.Itoa(i)
		s.skuReaderMock.EXPECT().Read(s.deadline).Times(1).Return(randomSku, nil)
		s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: randomSku}).Return(nil).Times(1)
	}

	s.skuReaderMock.EXPECT().Read(s.deadline).AnyTimes().Return("terminate", nil)

	report := s.server.Run(s.ctx, 500, s.deadline)
	s.Require().Equal(10000, report.CreatedSkus)
	s.Require().Equal(0, report.DuplicatedSkus)
	s.Require().Equal(0, report.InvalidSkus)
}

func (s *UnitSuite) TestInvalidSkusCanBeUpdatedInAConcurrentWayWithNoRaceConditions() {
	invalidSku := "invalid-sku"
	s.skuReaderMock.EXPECT().Read(s.deadline).Times(10000).Return(invalidSku, nil)
	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, create_sku.Command{Sku: invalidSku}).Return(domain.ErrInvalidSku).Times(10000)

	s.skuReaderMock.EXPECT().Read(s.deadline).AnyTimes().Return("terminate", nil)

	report := s.server.Run(s.ctx, 500, s.deadline)
	s.Require().Equal(0, report.CreatedSkus)
	s.Require().Equal(0, report.DuplicatedSkus)
	s.Require().Equal(10000, report.InvalidSkus)
}

func (s *UnitSuite) TestSkuReaderReadIsNotCalledWhenMaxConnectionsIsZeroAndEmptyReportIsReturned() {
	s.skuReaderMock.EXPECT().Read(s.deadline).Times(0)
	s.createSkuCommandHandlerMock.EXPECT().Handle(s.ctx, gomock.Any()).Times(0)
	s.requireEmptyReportAndNoSkusLogged(s.server.Run(s.ctx, 0, s.deadline))
}

func (s *UnitSuite) requireEmptyReportAndNoSkusLogged(report server.Report) {
	s.Require().Equal(0, report.CreatedSkus)
	s.Require().Equal(0, report.DuplicatedSkus)
	s.Require().Equal(0, report.InvalidSkus)
	s.Require().Empty(s.loggerBuffer.String())
}

func (s *UnitSuite) TestServerFinishAndAnEmptyReportIsReturnedWhenContextIsDoneDueToCancel() {
	s.skuReaderMock.EXPECT().Read(s.deadline).AnyTimes().Return("terminate", nil)
	ctx, cancelFunc := context.WithCancel(s.ctx)
	var report server.Report
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		report = s.server.Run(ctx, maxConnections, s.deadline)
		wg.Done()
	}()
	wg.Wait()
	cancelFunc()
	s.requireEmptyReportAndNoSkusLogged(report)
}

func (s *UnitSuite) TestServerFinishAndTheSkuIsLoggedWhenContextIsDoneDueToTimeout() {
	s.skuReaderMock.EXPECT().Read(s.deadline).AnyTimes().Return(sku, nil)
	s.createSkuCommandHandlerMock.EXPECT().Handle(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	s.createSkuCommandHandlerMock.EXPECT().Handle(gomock.Any(), gomock.Any()).AnyTimes().Return(domain.ErrSkuAlreadyExists)
	ctx, cancelFunc := context.WithTimeout(s.ctx, 0)
	defer cancelFunc()
	var report server.Report
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		report = s.server.Run(ctx, maxConnections, s.deadline)
		wg.Done()
	}()
	wg.Wait()

	s.Require().Equal(1, report.CreatedSkus)
	s.Require().GreaterOrEqual(report.DuplicatedSkus, 0)
	s.Require().Equal(0, report.InvalidSkus)
	s.Require().Equal(sku+"\n", s.loggerBuffer.String())
}