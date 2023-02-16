package gorilla

import (
	"github.com/youminxue/odin/framework/rest"
	"testing"
)

func TestDefaultHttpSrv_printRoutes(t *testing.T) {
	srv := NewRestServer()
	srv.gddRoutes = append(srv.gddRoutes, []rest.Route{
		{
			Name:    "GetStatsvizWs",
			Method:  "GET",
			Pattern: gddPathPrefix + "statsviz/ws",
		},
		{
			Name:    "GetStatsviz",
			Method:  "GET",
			Pattern: gddPathPrefix + "statsviz/",
		},
	}...)
	srv.printRoutes()
}
