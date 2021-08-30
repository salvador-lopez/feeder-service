//+build integration

package sku_reader_test

import (
	"feeder-service/internal/sku/ui/socket/tcp/sku_reader"
	"github.com/stretchr/testify/suite"
	"net"
	"testing"
)

const addr = "localhost:4000"

type IntegrationSuite struct {
	suite.Suite
	skuReader *sku_reader.SkuReaderImpl
}

func (s *IntegrationSuite) SetupTest() {

}

func (s *IntegrationSuite) TearDownTest() {
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func (s *IntegrationSuite) TestRead() {
	s.skuReader = sku_reader.New(addr)

	expectedMessage := "hello"

	go func() {
		message, err := s.skuReader.Read()
		s.Require().NoError(err)
		s.Require().Equal(expectedMessage, message)
	}()

	conn, err := net.Dial("tcp", addr)
	s.Require().NoError(err)
	defer conn.Close()

	_, err = conn.Write([]byte(expectedMessage))
	s.Require().NoError(err)
}

func (s *IntegrationSuite) TestErrorReadWhenAddrDoesNotMatch() {
	s.skuReader = sku_reader.New("localhost:5000")
	expectedMessage := "hello"

	go func() {
		message, err := s.skuReader.Read()
		s.Require().Error(err)
		s.Require().Equal(expectedMessage, message)
	}()

	_, err := net.Dial("tcp", addr)
	s.Require().Error(err)
}