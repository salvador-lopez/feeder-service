//+build unit

package create_sku_test

import (
	"context"
	"errors"
	"feeder-service/internal/sku/application/command/create_sku"
	"feeder-service/internal/sku/domain"
	"feeder-service/internal/sku/domain/mock"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"testing"
)

const sku = "KASL-3423"

type UnitSuite struct {
	suite.Suite
	ctx            context.Context
	repositoryMock *mock.MockSkuRepository
	mockCtrl       *gomock.Controller
	handler        *create_sku.CommandHandler
}

func (s *UnitSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.repositoryMock = mock.NewMockSkuRepository(s.mockCtrl)
	s.handler = create_sku.NewCommandHandler(s.repositoryMock)
}

func (s *UnitSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(UnitSuite))
}

func (s *UnitSuite) TestSaveSku() {
	s.repositorySaveNoErrorExpectation(sku)

	err := s.executeCommandHandler(sku)
	s.Require().NoError(err)
}

func (s *UnitSuite) TestReturnErrCreatingSkuWhenCallToRepositorySaveReturnError() {
	repositoryError := errors.New("repository error")
	s.repositoryMock.EXPECT().Save(s.ctx, gomock.Any()).Times(1).Return(repositoryError)
	s.executeTestErrCreatingSku(repositoryError.Error())
}

func (s *UnitSuite) TestReturnErrCreatingSkuWhenAlreadyExists() {
	skuId, err := domain.NewSkuId(sku)
	s.Require().NoError(err)
	alreadyExistingSku := domain.NewSku(skuId)
	s.repositoryMock.EXPECT().Save(s.ctx, alreadyExistingSku).Times(1).Return(fmt.Errorf("%w: %s", domain.ErrSkuAlreadyExists, sku))

	s.executeTestErrSkuAlreadyExists()
}

func (s *UnitSuite) TestReturnErrInvalidSkuWhenHasLessThanNineCharacters() {
	s.executeTestInvalidSku("ABCD-123")
}

func (s *UnitSuite) TestReturnErrInvalidSkuWhenHasMoreThanNineCharacters() {
	s.executeTestInvalidSku("ABCD-12345")
}

func (s *UnitSuite) TestReturnErrInvalidSkuWhenTheFifthCharacterIsNotADash() {
	s.executeTestInvalidSku("ABCDE1234")
}

func (s *UnitSuite) TestReturnErrInvalidSkuWhenTheFirstFourCharactersAreNotLetters() {
	s.executeTestInvalidSku("1234-1234")
}

func (s *UnitSuite) TestReturnErrInvalidSkuWhenTheLastFourCharactersAreNotNumbers() {
	s.executeTestInvalidSku("ABCD-ABCD")
}

func (s *UnitSuite) repositorySaveNoErrorExpectation(sku string) {
	s.repositoryMock.EXPECT().Save(s.ctx, gomock.AssignableToTypeOf(&domain.Sku{})).Times(1).DoAndReturn(func(ctx context.Context, skuEntity *domain.Sku) error {
		expectedSkuId, err := domain.NewSkuId(sku)
		s.Require().NoError(err)
		s.Require().True(expectedSkuId.Equal(skuEntity.Id()))
		return nil
	})
}

func (s *UnitSuite) executeCommandHandler(sku string) error{
	return s.handler.Handle(s.ctx, create_sku.Command{
		Sku: sku,
	})
}

func (s *UnitSuite) executeTestInvalidSku(invalidSku string) {
	err := s.executeCommandHandler(invalidSku)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, domain.ErrInvalidSku))
	s.Require().Equal("invalid Sku provided: "+invalidSku, err.Error())
}

func (s *UnitSuite) executeTestErrCreatingSku(errorMsg string) {
	err := s.executeCommandHandler(sku)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, create_sku.ErrCreatingSku))
	s.Require().Equal("error creating sku "+sku+": "+errorMsg, err.Error())
}

func (s *UnitSuite) executeTestErrSkuAlreadyExists() {
	err := s.executeCommandHandler(sku)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, domain.ErrSkuAlreadyExists))
	s.Require().Equal("sku already exists: "+sku, err.Error())
}
