package httpserver

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

type KafkaRequest struct {
	Name    string `param:"name"`
	Topic   string `param:"topic"`
	GroupID string `query:"groupId"`
}

func (r *KafkaRequest) CheckNameAndTopic() error {
	if r.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": "require name in path",
		})
	}
	if r.Topic == "" {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": "require topic in path",
		})
	}
	return nil
}

func (r *KafkaRequest) CheckNameTopicGroupID() error {
	if err := r.CheckNameAndTopic(); err != nil {
		return err
	}
	if r.GroupID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": "require groupId in query",
		})
	}
	return nil
}

type ProducerStartRequest struct {
	KafkaRequest
	Interval string `query:"interval"`
	interval time.Duration
}

func (r *ProducerStartRequest) Bind(ctx echo.Context) error {
	if err := ctx.Bind(r); err != nil {
		return err
	}
	if err := (&echo.DefaultBinder{}).BindQueryParams(ctx, r); err != nil {
		return err
	}
	if err := r.CheckNameAndTopic(); err != nil {
		return err
	}
	interval, err := parseDuration(r.Interval)
	if err != nil {
		return err
	}
	r.interval = interval
	return nil
}

type ProducerStopRequest struct {
	KafkaRequest
}

func (r *ProducerStopRequest) Bind(ctx echo.Context) error {
	if err := ctx.Bind(r); err != nil {
		return err
	}
	if err := (&echo.DefaultBinder{}).BindQueryParams(ctx, r); err != nil {
		return err
	}
	if err := r.CheckNameAndTopic(); err != nil {
		return err
	}
	return nil
}

type ConsumerStartRequest struct {
	KafkaRequest
	ConsumeInterval string `query:"interval"`
	ShouldFail      string `query:"shouldFail"`
	consumeInterval time.Duration
	shouldFail      bool
}

func (r *ConsumerStartRequest) Bind(ctx echo.Context) error {
	if err := ctx.Bind(r); err != nil {
		return err
	}
	if err := (&echo.DefaultBinder{}).BindQueryParams(ctx, r); err != nil {
		return err
	}
	if err := r.CheckNameTopicGroupID(); err != nil {
		return err
	}
	if r.ConsumeInterval != "" {
		interval, err := parseDuration(r.ConsumeInterval)
		if err != nil {
			return err
		}
		r.consumeInterval = interval
	}
	if strings.EqualFold(r.ShouldFail, "true") {
		r.shouldFail = true
	}
	return nil
}

type ConsumerStopRequest struct {
	KafkaRequest
}

func (r *ConsumerStopRequest) Bind(ctx echo.Context) error {
	if err := ctx.Bind(r); err != nil {
		return err
	}
	if err := (&echo.DefaultBinder{}).BindQueryParams(ctx, r); err != nil {
		return err
	}
	if err := r.CheckNameAndTopic(); err != nil {
		return err
	}
	return nil
}

type ConsumerUpdateRequest struct {
	KafkaRequest
	ConsumeInterval string `query:"interval"`
	ShouldFail      string `query:"shouldFail"`
	consumeInterval time.Duration
	shouldFail      bool
}

func (r *ConsumerUpdateRequest) Bind(ctx echo.Context) error {
	if err := ctx.Bind(r); err != nil {
		return err
	}
	if err := (&echo.DefaultBinder{}).BindQueryParams(ctx, r); err != nil {
		return err
	}
	if err := r.CheckNameAndTopic(); err != nil {
		return err
	}
	if r.ConsumeInterval != "" {
		interval, err := parseDuration(r.ConsumeInterval)
		if err != nil {
			return err
		}
		r.consumeInterval = interval
	}
	if strings.EqualFold(r.ShouldFail, "true") {
		r.shouldFail = true
	}
	return nil
}

type StatusResponse struct {
	Status   string   `json:"status"`
	Metadata echo.Map `json:"metadata"`
}

func parseDuration(val string) (time.Duration, error) {
	if val == "" {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "require interval in query params")
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid duration: %s", val))
	}
	return d, nil
}
