package rest

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// Oas store OpenAPI3.0 description json string
var Oas string
var DocRoutes = docRoutes

func docRoutes() []Route {
	return []Route{
		{
			Name:    "GetDoc",
			Method:  "GET",
			Pattern: "/odin/doc",
			HandlerFunc: func(_writer http.ResponseWriter, _req *http.Request) {
				var (
					tpl    *template.Template
					err    error
					buf    bytes.Buffer
					scheme string
				)
				if tpl, err = template.New("onlinedoc.tmpl").Parse(onlinedocTmpl); err != nil {
					panic(err)
				}
				if _req.TLS == nil {
					scheme = "http"
				} else {
					scheme = "https"
				}
				doc := Oas
				docUrl := fmt.Sprintf("%s://%s/odin/openapi.json", scheme, _req.Host)
				if err = tpl.Execute(&buf, struct {
					Doc    string
					DocUrl string
				}{
					Doc:    doc,
					DocUrl: docUrl,
				}); err != nil {
					panic(err)
				}
				_writer.Header().Set("Content-Type", "text/html; charset=utf-8")
				_writer.Write(buf.Bytes())
			},
		},
		{
			Name:    "GetOpenAPI",
			Method:  "GET",
			Pattern: "/odin/openapi.json",
			HandlerFunc: func(_writer http.ResponseWriter, _req *http.Request) {
				_writer.Write([]byte(Oas))
			},
		},
	}
}
