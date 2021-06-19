package httpserver

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/zacscoding/burrow-all/internal/kafka"
	"sync"
)

type Server struct {
	producers   map[string]*kafka.Producer
	consumers   map[string]*kafka.Consumer
	e           *echo.Echo
	kafkaCfg    *kafka.Config
	openEvents  []map[string]interface{}
	closeEvents []map[string]interface{}

	pmutex sync.Mutex
	cmutex sync.Mutex
	emutex sync.RWMutex
}

func NewServer(e *echo.Echo, kafkaCfg *kafka.Config) *Server {
	s := &Server{
		producers: make(map[string]*kafka.Producer),
		consumers: make(map[string]*kafka.Consumer),
		e:         e,
		kafkaCfg:  kafkaCfg,
		pmutex:    sync.Mutex{},
		cmutex:    sync.Mutex{},
		emutex:    sync.RWMutex{},
	}

	producer := s.e.Group("/v1/producer")
	producer.POST("/:name/:topic", s.handleStartProducer)
	producer.DELETE("/:name/:topic", s.handleStopProducer)

	consumer := s.e.Group("/v1/consumer")
	consumer.POST("/:name/:topic", s.handleStartConsumer)
	consumer.PUT("/:name/:topic", s.handleUpdateConsumer)
	consumer.DELETE("/:name/:topic", s.handleStopConsumer)

	event := s.e.Group("/v1/event")
	event.GET("", s.handleGetEvents)
	event.POST("", s.handleOpenEvent)
	event.DELETE("", s.handleOpenEvent)

	return s
}

func (s *Server) Start(addr string) error {
	return s.e.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.pmutex.Lock()
	defer s.pmutex.Unlock()
	s.cmutex.Lock()
	defer s.cmutex.Unlock()
	for _, p := range s.producers {
		p.Stop()
	}
	for _, c := range s.consumers {
		c.Stop()
	}
	return s.e.Shutdown(ctx)
}
