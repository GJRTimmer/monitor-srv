package main

import (
	log "github.com/golang/glog"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/server"
	"github.com/micro/monitoring-srv/handler"
	"github.com/micro/monitoring-srv/monitor"
	proto "github.com/micro/monitoring-srv/proto/monitor"
)

func main() {
	cmd.Init()

	server.Init(
		server.Name("go.micro.srv.monitoring"),
	)

	proto.RegisterMonitorHandler(server.DefaultServer, new(handler.Monitor))

	server.Subscribe(
		server.NewSubscriber(monitor.HealthCheckTopic, monitor.DefaultMonitor.ProcessHealthCheck),
	)

	monitor.DefaultMonitor.Run()

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
