package main

import (
	"net/http"

	"github.com/calebamiles/example-money-making-fortune-api/cadence/activity"
	"github.com/calebamiles/example-money-making-fortune-api/cadence/workflow"
	"github.com/calebamiles/example-money-making-fortune-api/service"

	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/worker"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// hostPort is the location of the cadence frontend
	hostPort = "127.0.0.1:7933"

	// domain is the namespace to use
	domain = "hcp"

	// clientName is the name of the worker
	clientName = "figletized-fortune-worker"

	// cadenceService is the Cadence service to connect to
	cadenceService = "cadence-frontend"
)

func main() {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	var err error
	logger, err := config.Build()
	if err != nil {
		logger.Fatal("Failed to setup logger", zap.Error(err))
	}

	ch, err := tchannel.NewChannelTransport(tchannel.ServiceName(clientName))
	if err != nil {
		logger.Fatal("Failed to setup tchannel", zap.Error(err))
	}
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: clientName,
		Outbounds: yarpc.Outbounds{
			cadenceService: {Unary: ch.NewSingleOutbound(hostPort)},
		},
	})

	if err := dispatcher.Start(); err != nil {
		logger.Fatal("Failed to start dispatcher", zap.Error(err))
	}

	thriftService := workflowserviceclient.New(dispatcher.ClientConfig(cadenceService))

	workerOptions := worker.Options{
		Logger:       logger,
		MetricsScope: tally.NewTestScope(workflow.TaskList, map[string]string{}),
	}

	worker := worker.New(
		thriftService,
		domain,
		workflow.TaskList,
		workerOptions,
	)

	worker.RegisterActivity(activity.GetFigletizedFortune)
	worker.RegisterWorkflow(workflow.GetFigletizedFortune)

	err = worker.Start()
	if err != nil {
		logger.Error("Failed to start worker", zap.Error(err))
	}

	logger.Info("Started Fortune Cadence worker.", zap.String("worker", workflow.TaskList))

	http.HandleFunc("/fortune", service.HandleGetFigletizedFortuneCadence)
	http.HandleFunc("/healthz", service.HandleGetHealthz)

	logger.Info("Starting HTTP server.", zap.String("service", "figleted-fortune"))
	http.ListenAndServe(":8092", nil)
}
