//+build integration

package mongo_test

import (
	"context"
	"feeder-service/internal/sku/domain"
	mongo2 "feeder-service/internal/sku/infrastructure/persistence/mongo"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

const skuValue = "KASL-3423"

type IntegrationSuite struct {
	suite.Suite
	ctx        context.Context
	db         *mongo.Database
	repository *mongo2.SkuRepository
}

func (s *IntegrationSuite) SetupSuite() {
	s.ctx = context.Background()

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	s.Require().NoError(err)
	err = mongoClient.Connect(s.ctx)
	s.Require().NoError(err)

	s.db = mongoClient.Database("sku_integration_test")
	s.repository, err = mongo2.NewSkuRepository(s.db, domain.NewHydrator())
	s.Require().NoError(err)
}

func (s *IntegrationSuite) TearDownSuite() {
	err := s.db.Drop(s.ctx)
	s.Require().NoError(err)
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func (s *IntegrationSuite) TestSaveAndFind() {
	skuId, err := domain.NewSkuId(skuValue)
	s.Require().NoError(err)
	sku := domain.NewSku(skuId)

	err = s.repository.Save(s.ctx, sku)
	s.Require().NoError(err)

	skuFromRepository, err := s.repository.Find(s.ctx, skuId)
	s.Require().NoError(err)
	s.Require().True(sku.Id().Equal(skuFromRepository.Id()))
}