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
	listener net.Listener
	skuReader *sku_reader.SkuReaderImpl
}

func (s *IntegrationSuite) SetupTest() {
	s.deadline = time.Now().Add(1 * time.Second)
	listener, err := net.Listen("tcp", addr)
	s.Require().NoError(err)
	s.listener = listener
	skuReader, err := sku_reader.New(listener)
	s.Require().NoError(err)
	s.skuReader = skuReader
}

func (s *IntegrationSuite) TearDownTest() {
	err := s.listener.Close()
	s.Require().NoError(err)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func (s *IntegrationSuite) TestRead() {
	expectedMessage := "KASL-3423"

	var readErrorChan = make(chan error)
	var readMessageChan = make(chan string)
	go func() {
		readMessage, readError := s.skuReader.Read(s.deadline)
		if readError != nil {
			readErrorChan <- readError
			return
		}
		readMessageChan <- readMessage
	}()

	s.sendMessageFromAClient("000"+expectedMessage)

	testFinishDeadline := s.deadline.Add(1 * time.Second)
	select {
		case readMessage := <- readMessageChan:
			s.Require().Equal(expectedMessage, readMessage)
		case readError := <- readErrorChan:
			s.FailNow(readError.Error())
		case <-time.After(testFinishDeadline.Sub(time.Now())):
			s.FailNow("skuReader.Read() operation does not finish on time")
	}
}

func (s *IntegrationSuite) TestReadDeadlineExceed() {
	var readErrorChan = make(chan error)
	var readMessageChan = make(chan string)
	go func() {
		readMessage, readError := s.skuReader.Read(s.deadline)
		if readError != nil {
			readErrorChan <- readError
			return
		}
		readMessageChan <- readMessage
	}()

	testFinishDeadline := s.deadline.Add(1 * time.Second)
	select {
	case <- readMessageChan:
		s.FailNow("no message was expected")
	case readError := <- readErrorChan:
		s.Require().Error(readError)
	case <-time.After(testFinishDeadline.Sub(time.Now())):
		s.FailNow("skuReader.Read() operation does not finish on time")
	}
}

func (s *IntegrationSuite) sendMessageFromAClient(messageToSend string) {
	conn, err := net.Dial("tcp", addr)
	s.Require().NoError(err)
	defer conn.Close()

	_, err = conn.Write([]byte(messageToSend))
	s.Require().NoError(err)
}
