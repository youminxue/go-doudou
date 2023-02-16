package main

import (
	"github.com/youminxue/odin/framework/rest"
	service "testdata"
	"testdata/config"
	"testdata/transport/httpsrv"
)

func main() {
	conf := config.LoadFromEnv()
	svc := service.NewUsersvc(conf)
	handler := httpsrv.NewUsersvcHandler(svc)
	srv := rest.NewRestServer()
	srv.AddRoute(httpsrv.Routes(handler)...)
	srv.Run()
}
