package httpserver

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/zacscoding/burrow-all/internal/kafka"
	"net/http"
)

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
		c.SetShouldFail(params.shouldFail)
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
