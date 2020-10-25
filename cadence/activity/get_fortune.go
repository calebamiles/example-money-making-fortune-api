package activity

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

const (
	fortuneBackendURL = "http://127.0.0.1:8090/fortune"
	figletBackendURL  = "http://127.0.0.1:8091/figlet"
	postContentType   = "application/octet-stream"
)

// GetFigletizedFortune returns a fortune with a Figlet transformation applied by:
// - getting a fortune from the Fortune API
// - appling a Figlet transformation of that fortune using the Figlet API
func GetFigletizedFortune(ctx context.Context) (string, error) {
	fortuneResp, err := http.Get(fortuneBackendURL)
	if err != nil {
		activity.GetLogger(ctx).Error("Failed to GET fortune from fortune backend", zap.Error(err))
		return "", err
	}

	defer fortuneResp.Body.Close()

	fortuneBytes, err := ioutil.ReadAll(fortuneResp.Body)
	if err != nil {
		activity.GetLogger(ctx).Error("Failed to read fortune from backend response", zap.Error(err))
		return "", err
	}

	if len(fortuneBytes) == 0 {
		activity.GetLogger(ctx).Error("Zero length response from fortune backend")
		return "", err
	}

	fortuneReader := bytes.NewBuffer(fortuneBytes)

	figletResp, err := http.Post(figletBackendURL, postContentType, fortuneReader)
	if err != nil {
		activity.GetLogger(ctx).Error("Failed to POST fortune to figlet backend", zap.Error(err))
		return "", err
	}

	defer figletResp.Body.Close()

	figletedFortuneBytes, err := ioutil.ReadAll(figletResp.Body)
	if err != nil {
		activity.GetLogger(ctx).Error("Failed to read figleted fortune from backend response", zap.Error(err))
		return "", err
	}

	return string(figletedFortuneBytes), nil
}
