package main

import (
	log "github.com/golang/glog"
	"github.com/micro/go-micro"
	"github.com/micro/monitoring-srv/handler"
	"github.com/micro/monitoring-srv/monitor"
	proto "github.com/micro/monitoring-srv/proto/monitor"
)

func main() {
	service := micro.NewService(
		micro.Name("go.micro.srv.monitoring"),
		// before starting
		micro.BeforeStart(func() error {
			monitor.DefaultMonitor.Run()
			return nil
		}),
	)

	service.Init()

	service.Server().Subscribe(
		service.Server().NewSubscriber(monitor.HealthCheckTopic, monitor.DefaultMonitor.ProcessHealthCheck),
	)

	proto.RegisterMonitorHandler(service.Server(), new(handler.Monitor))

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
