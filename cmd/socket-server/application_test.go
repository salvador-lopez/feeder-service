//+build acceptance

package main

import (
	"context"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net"
	"os"
	"sync"
	"testing"
)

type ApplicationSuite struct {
	suite.Suite
	cfg *config
}

func (s *ApplicationSuite) SetupSuite() {
	cfg, err := fetchConfigFromEnvVars()
	s.Require().NoError(err)
	s.cfg = cfg
}

func (s *ApplicationSuite) TearDownSuite() {
	s.removeLogFile()
	s.dropMongoDatabase()
}

func TestAcceptance(t *testing.T) {
	suite.Run(t, new(ApplicationSuite))
}

func (s *ApplicationSuite) TestServeToFiveConcurrentClients() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		main()
		wg.Done()
	}()

	const (
		aUniqueSku       = "KASL-3423"
		anotherUniqueSku = "SLOS-4332"
	)

	s.sendMessageFromAClient(aUniqueSku)
	s.sendMessageFromAClient("LPOS-32411)")
	s.sendMessageFromAClient(anotherUniqueSku)
	s.sendMessageFromAClient("invalid-sku")
	s.sendMessageFromAClient("KASL-3423")
	s.sendMessageFromAClient("terminate")
	wg.Wait()

	fileData, err := os.ReadFile(s.cfg.logFileName)
	s.Require().NoError(err)
	s.Require().Contains(string(fileData), aUniqueSku)
	s.Require().Contains(string(fileData), anotherUniqueSku)
}

func (s *ApplicationSuite) sendMessageFromAClient(messageToSend string) {
	conn, err := net.Dial("tcp", s.cfg.socketAddr)
	s.Require().NoError(err)
	defer conn.Close()

	_, err = conn.Write([]byte(messageToSend))
	s.Require().NoError(err)
}

func (s *ApplicationSuite) removeLogFile() {
	err := os.Remove(s.cfg.logFileName)
	s.Require().NoError(err)
}

func (s *ApplicationSuite) dropMongoDatabase() {
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(s.cfg.mongoUri))
	s.Require().NoError(err)
	ctx := context.Background()
	err = mongoClient.Connect(ctx)
	s.Require().NoError(err)

	db := mongoClient.Database(s.cfg.mongoDatabase)
	err = db.Drop(ctx)
	s.Require().NoError(err)
}
