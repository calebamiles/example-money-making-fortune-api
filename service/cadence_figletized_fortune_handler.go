package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/calebamiles/example-money-making-fortune-api/cadence/workflow"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/client"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// hostPort is the location of the cadence frontend
	hostPort = "127.0.0.1:7933"

	// clientName is the name of the client
	clientName = "get-figleted-fortune-http-handler"

	// clientService is the Cadence service to connect to
	clientService = "cadence-frontend"
)

// HandleGetFigletizedFortuneCadence gets a fortune with a Figlet transformation applied
// using the Cadence backend
func HandleGetFigletizedFortuneCadence(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cadenceOpts := &client.Options{
		Identity: clientName,
	}

	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		logger.Fatal("Failed to setup logger", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ch, err := tchannel.NewChannelTransport(tchannel.ServiceName(clientName))
	if err != nil {
		logger.Error("Failed to setup tchannel", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: clientName,
		Outbounds: yarpc.Outbounds{
			clientService: {Unary: ch.NewSingleOutbound(hostPort)},
		},
	})

	if err := dispatcher.Start(); err != nil {
		logger.Error("Failed to start dispatcher", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	thriftService := workflowserviceclient.New(dispatcher.ClientConfig(clientService))
	cadence := client.NewClient(thriftService, "hcp", cadenceOpts)

	startOpts := client.StartWorkflowOptions{
		TaskList:                        workflow.TaskList,
		ExecutionStartToCloseTimeout:    60 * time.Second,
		DecisionTaskStartToCloseTimeout: 20 * time.Second,
		WorkflowIDReusePolicy:           client.WorkflowIDReusePolicyAllowDuplicateFailedOnly,
		Memo:                            map[string]interface{}{"workflow-type": "local-development"},
	}

	future, err := cadence.ExecuteWorkflow(ctx, startOpts, workflow.GetFigletizedFortune)
	if err != nil {
		logger.Error("Executing GetFigletizedFortune workflow", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var fortune string
	err = future.Get(ctx, &fortune)
	if err != nil {
		logger.Error("Getting GetFigletizedFortune workflow result", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	n, err := w.Write([]byte(fortune))
	if err != nil {
		logger.Error("Writing HandleGetFigletizedFortuneCadence response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if n != len(fortune) {
		logger.Error(fmt.Sprintf("Expected to write %d bytes, but only wrote %d", len(fortune), n))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
