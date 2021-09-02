//+build integration

package sku_reader_test

import (
	"feeder-service/internal/sku/infrastructure/io/socket/tcp/sku_reader"
	"github.com/stretchr/testify/suite"
	"net"
	"testing"
	"time"
)

const (
	addr    = "localhost:4000"
)

type IntegrationSuite struct {
	suite.Suite
	deadline  time.Time
	skuReader *sku_reader.SkuReaderImpl
}

func (s *IntegrationSuite) SetupTest() {
	s.deadline = time.Now().Add(10 * time.Second)
}

func (s *IntegrationSuite) TearDownTest() {
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func (s *IntegrationSuite) TestRead() {
	skuReader, err := sku_reader.New(addr)
	s.Require().NoError(err)
	s.skuReader = skuReader

	expectedMessage := "KASL-3423"

	go func() {
		message, err := s.skuReader.Read(s.deadline)
		s.Require().NoError(err)
		s.Require().Equal(expectedMessage, message)
	}()

	conn, err := net.Dial("tcp", addr)
	s.Require().NoError(err)
	defer conn.Close()

	_, err = conn.Write([]byte("000"+expectedMessage))
	s.Require().NoError(err)
}

func (s *IntegrationSuite) TestErrorReadWhenAddrDoesNotMatch() {
	skuReader, err := sku_reader.New("localhost:5000")
	s.Require().NoError(err)
	s.skuReader = skuReader
	expectedMessage := "hello"

	go func() {
		message, err := s.skuReader.Read(s.deadline)
		s.Require().Error(err)
		s.Require().Equal(expectedMessage, message)
	}()

	_, err = net.Dial("tcp", addr)
	s.Require().Error(err)
}