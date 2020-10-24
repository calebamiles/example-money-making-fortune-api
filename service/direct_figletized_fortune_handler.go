package service

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	fortuneBackendURL = "http://127.0.0.1:8090/fortune"
	figletBackendURL  = "http://127.0.0.1:8091/figlet"
	postContentType   = "application/octet-stream"
)

// HandleGetFigletizedFortuneDirect returns a fortune with the Figlet transformation applied
// by directly interacting with backend fortune, and figlet services
func HandleGetFigletizedFortuneDirect(w http.ResponseWriter, req *http.Request) {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		logger.Error("Failed to setup logger", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fortuneResp, err := http.Get(fortuneBackendURL)
	if err != nil {
		logger.Error("Failed to GET fortune from fortune backend", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	defer fortuneResp.Body.Close()

	fortuneBytes, err := ioutil.ReadAll(fortuneResp.Body)
	if err != nil {
		logger.Error("Failed to read fortune from backend response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	if len(fortuneBytes) == 0 {
		logger.Error("Zero length response from fortune backend")
		w.WriteHeader(http.StatusInternalServerError)
	}

	fortuneReader := bytes.NewBuffer(fortuneBytes)

	figletResp, err := http.Post(figletBackendURL, postContentType, fortuneReader)
	if err != nil {
		logger.Error("Failed to POST fortune to figlet backend", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	defer figletResp.Body.Close()

	figletedFortuneBytes, err := ioutil.ReadAll(figletResp.Body)
	if err != nil {
		logger.Error("Failed to read figleted fortune from backend response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	n, err := w.Write(figletedFortuneBytes)
	if err != nil {
		logger.Error("Writing HandleGetFigletizedFortuneDirect response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if n != len(figletedFortuneBytes) {
		logger.Error(fmt.Sprintf("Expected to write %d bytes, but only wrote %d", len(figletedFortuneBytes), n))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
