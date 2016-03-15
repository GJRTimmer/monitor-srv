package main

import (
	"log"
	"github.com/micro/go-micro"
	"github.com/micro/monitor-srv/handler"
	"github.com/micro/monitor-srv/monitor"
	proto "github.com/micro/monitor-srv/proto/monitor"
)

func main() {
	service := micro.NewService(
		micro.Name("go.micro.srv.monitor"),
		// before starting
		micro.BeforeStart(func() error {
			monitor.DefaultMonitor.Run()
			return nil
		}),
	)

	service.Init()

	// healthchecks
	service.Server().Subscribe(
		service.Server().NewSubscriber(
			monitor.HealthCheckTopic,
			monitor.DefaultMonitor.ProcessHealthCheck,
		),
	)

	// status
	service.Server().Subscribe(
		service.Server().NewSubscriber(
			monitor.StatusTopic,
			monitor.DefaultMonitor.ProcessStatus,
		),
	)

	// stats
	service.Server().Subscribe(
		service.Server().NewSubscriber(
			monitor.StatsTopic,
			monitor.DefaultMonitor.ProcessStats,
		),
	)

	proto.RegisterMonitorHandler(service.Server(), new(handler.Monitor))

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
