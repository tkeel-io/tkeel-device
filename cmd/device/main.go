package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/kit/transport"
	"device/pkg/server"
	"device/pkg/service"
)

import (//User import
	helloworld "device/api/helloworld/v1"
	openapi "device/api/openapi/v1"
	Device_v1 "device/api/device/v1"
	Group_v1 "device/api/group/v1"
	Measure_v1 "device/api/measure/v1"
)

var (
	// app name
	Name string
	// http addr
	HTTPAddr string
	// grpc addr
	GRPCAddr string
)

func init() {
	flag.StringVar(&Name, "name", "hello", "app name.")
	flag.StringVar(&HTTPAddr, "http_addr", ":31234", "http listen address.")
	flag.StringVar(&GRPCAddr, "grpc_addr", ":31233", "grpc listen address.")
}

func main() {
	flag.Parse()

	httpSrv := server.NewHTTPServer(HTTPAddr)
	grpcSrv := server.NewGRPCServer(GRPCAddr)
	serverList := []transport.Server{httpSrv, grpcSrv}

	app := app.New(Name,
		&log.Conf{
			App:   Name,
			Level: "debug",
			Dev:   true,
		},
		serverList...,
	)


	{//User service
		GreeterSrv := service.NewGreeterService()
		helloworld.RegisterGreeterHTTPServer(httpSrv.Container, GreeterSrv)
		helloworld.RegisterGreeterServer(grpcSrv.GetServe(), GreeterSrv)

		OpenapiSrv := service.NewOpenapiService()
		openapi.RegisterOpenapiHTTPServer(httpSrv.Container, OpenapiSrv)
		openapi.RegisterOpenapiServer(grpcSrv.GetServe(), OpenapiSrv)

		DeviceSrv := service.NewDeviceService()
		Device_v1.RegisterDeviceHTTPServer(httpSrv.Container, DeviceSrv)
		Device_v1.RegisterDeviceServer(grpcSrv.GetServe(), DeviceSrv)

		GroupSrv := service.NewGroupService()
		Group_v1.RegisterGroupHTTPServer(httpSrv.Container, GroupSrv)
		Group_v1.RegisterGroupServer(grpcSrv.GetServe(), GroupSrv)

		MeasureSrv := service.NewMeasureService()
		Measure_v1.RegisterMeasureHTTPServer(httpSrv.Container, MeasureSrv)
		Measure_v1.RegisterMeasureServer(grpcSrv.GetServe(), MeasureSrv)

	}


	if err := app.Run(context.TODO()); err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop

	if err := app.Stop(context.TODO()); err != nil {
		panic(err)
	}
}
