package httpserver

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"log"
	"net/http"
)

func (s *Server) handleGetEvents(ctx echo.Context) error {
	result := echo.Map{}
	eventType := ctx.QueryParam("type")
	s.emutex.RLock()
	defer s.emutex.RUnlock()

	if eventType == "" || eventType == "open" {
		var openEvents []map[string]interface{}
		copy(openEvents, s.openEvents)
		result["open"] = openEvents
	}
	if eventType == "" || eventType == "close" {
		var closeEvents []map[string]interface{}
		copy(closeEvents, s.closeEvents)
		result["close"] = closeEvents
	}
	return ctx.JSON(http.StatusOK, result)
}

func (s *Server) handleOpenEvent(ctx echo.Context) error {
	log.Println("[Notification] Called POST /v1/event")
	body, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		log.Println("> Failed to read body. err:", err)
		return err
	}
	var alert map[string]interface{}
	if err := json.Unmarshal(body, &alert); err != nil {
		log.Println("> Failed to unmarshal data. err:", err)
		return err
	}
	s.emutex.Lock()
	defer s.emutex.Unlock()
	s.openEvents = append(s.openEvents, alert)
	return ctx.JSON(http.StatusOK, nil)
}

func (s *Server) handleCloseEvent(ctx echo.Context) error {
	log.Println("[Notification] Called DELETE /v1/event")
	body, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		log.Println("> Failed to read body. err:", err)
		return err
	}
	var alert map[string]interface{}
	if err := json.Unmarshal(body, &alert); err != nil {
		log.Println("> Failed to unmarshal data. err:", err)
		return err
	}
	s.emutex.Lock()
	defer s.emutex.Unlock()
	s.closeEvents = append(s.closeEvents, alert)
	return ctx.JSON(http.StatusOK, nil)
}
