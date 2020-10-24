package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// HandleGetHealthz returns server health as ok
func HandleGetHealthz(w http.ResponseWriter, req *http.Request) {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		logger.Error("Failed to setup logger", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Status  string `json:"status"`
		Service string `json:"service"`
	}

	response.Status = "ok"
	response.Service = "money-making-figleted-fortune-service"
	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error("encoding status to JSON", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	n, err := w.Write(responseJSON)
	if err != nil {
		logger.Error("writing response: %s", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if n != len(responseJSON) {
		logger.Error(fmt.Sprintf("expected to write %d bytes, but only wrote %d", len(responseJSON), n))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
