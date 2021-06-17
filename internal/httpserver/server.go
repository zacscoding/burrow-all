package httpserver

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/zacscoding/burrow-all/internal/kafka"
	"net/http"
	"sync"
)

type Server struct {
	producers map[string]*kafka.Producer
	consumers map[string]*kafka.Consumer
	e         *echo.Echo
	kafkaCfg  *kafka.Config

	pmutex sync.Mutex
	cmutex sync.Mutex
}

func NewServer(e *echo.Echo, kafkaCfg *kafka.Config) *Server {
	s := &Server{
		producers: make(map[string]*kafka.Producer),
		consumers: make(map[string]*kafka.Consumer),
		e:         e,
		kafkaCfg:  kafkaCfg,
		pmutex:    sync.Mutex{},
		cmutex:    sync.Mutex{},
	}

	s.e.POST("/v1/producer/:name/:topic", s.handleStartProducer)
	s.e.DELETE("/v1/producer/:name/:topic", s.handleStopProducer)

	s.e.POST("/v1/consumer/:name/:topic", s.handleStartConsumer)
	s.e.PUT("/v1/consumer/:name/:topic", s.handleUpdateConsumer)
	s.e.DELETE("/v1/consumer/:name/:topic", s.handleStopConsumer)

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

func (s *Server) handleStartProducer(ctx echo.Context) error {
	params := new(ProducerStartRequest)
	if err := params.Bind(ctx); err != nil {
		return err
	}

	id := makeID(params.Name, params.Topic)
	s.pmutex.Lock()
	defer s.pmutex.Unlock()
	if _, ok := s.producers[id]; ok {
		return echo.NewHTTPError(http.StatusConflict, "already exist producer")
	}
	p, err := kafka.NewProducer(&kafka.ProducerConfig{
		KafkaCfg: s.kafkaCfg,
		Name:     id,
		Topic:    params.Topic,
		Interval: params.interval,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	s.producers[id] = p
	return ctx.JSON(http.StatusCreated, &StatusResponse{
		Status: "created",
		Metadata: echo.Map{
			"name":     id,
			"topic":    params.Topic,
			"interval": params.Interval,
		},
	})
}

func (s *Server) handleStopProducer(ctx echo.Context) error {
	params := new(ProducerStopRequest)
	if err := params.Bind(ctx); err != nil {
		return err
	}

	id := makeID(params.Name, params.Topic)
	s.pmutex.Lock()
	defer s.pmutex.Unlock()

	p, ok := s.producers[id]

	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("producer %s not found", id))
	}
	p.Stop()
	delete(s.producers, id)
	return ctx.JSON(http.StatusOK, &StatusResponse{
		Status: "deleted",
		Metadata: echo.Map{
			"name":  params.Name,
			"topic": params.Topic,
		},
	})
}

func (s *Server) handleStartConsumer(ctx echo.Context) error {
	params := new(ConsumerStartRequest)
	if err := params.Bind(ctx); err != nil {
		return err
	}

	id := makeID(params.Name, params.Topic)
	s.cmutex.Lock()
	defer s.cmutex.Unlock()
	if _, ok := s.consumers[id]; ok {
		return echo.NewHTTPError(http.StatusConflict, "already exist consumer")
	}
	c, err := kafka.NewConsumer(&kafka.ConsumerConfig{
		KafkaCfg:        s.kafkaCfg,
		Name:            id,
		GroupID:         params.GroupID,
		Topic:           params.Topic,
		ConsumeInterval: params.consumeInterval,
		ShouldFail:      params.shouldFail,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	s.consumers[id] = c
	return ctx.JSON(http.StatusCreated, &StatusResponse{
		Status: "created",
		Metadata: echo.Map{
			"name":       id,
			"topic":      params.Topic,
			"groupId":    params.GroupID,
			"interval":   params.consumeInterval,
			"shouldFail": params.ShouldFail,
		},
	})
}

func (s *Server) handleStopConsumer(ctx echo.Context) error {
	params := new(ConsumerStopRequest)
	if err := params.Bind(ctx); err != nil {
		return err
	}

	id := makeID(params.Name, params.Topic)
	s.cmutex.Lock()
	defer s.cmutex.Unlock()

	c, ok := s.consumers[id]

	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("consumer %s not found", id))
	}
	c.Stop()
	delete(s.consumers, id)
	return ctx.JSON(http.StatusOK, &StatusResponse{
		Status: "deleted",
		Metadata: echo.Map{
			"name":  params.Name,
			"topic": params.Topic,
		},
	})
}

func (s *Server) handleUpdateConsumer(ctx echo.Context) error {
	params := new(ConsumerUpdateRequest)
	if err := params.Bind(ctx); err != nil {
		return err
	}

	id := makeID(params.Name, params.Topic)
	s.cmutex.Lock()
	defer s.cmutex.Unlock()

	c, ok := s.consumers[id]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("consumer %s not found", id))
	}

	metadata := echo.Map{
		"name":  id,
		"topic": params.Topic,
	}
	if params.ConsumeInterval != "" {
		c.SetConsumeInterval(params.consumeInterval)
		metadata["interval"] = params.consumeInterval
	}
	if params.ShouldFail != "" {
		c.SetConsumeInterval(params.consumeInterval)
		metadata["shouldFail"] = params.shouldFail
	}
	return ctx.JSON(http.StatusCreated, &StatusResponse{
		Status:   "updated",
		Metadata: metadata,
	})
}

func makeID(name, topic string) string {
	return name + "_" + topic
}
