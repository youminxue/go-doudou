package v3

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/youminxue/v2/toolkit/astutils"
	"github.com/youminxue/v2/toolkit/constants"
	"github.com/youminxue/v2/version"
	"reflect"
	"strings"
	"time"
)

type ProtoGenerator struct {
	fieldNamingFunc func(string) string
}

type ProtoGeneratorOption func(*ProtoGenerator)

func WithFieldNamingFunc(fn func(string) string) ProtoGeneratorOption {
	return func(p *ProtoGenerator) {
		p.fieldNamingFunc = fn
	}
}

func NewProtoGenerator(options ...ProtoGeneratorOption) ProtoGenerator {
	var p ProtoGenerator
	for _, opt := range options {
		opt(&p)
	}
	return p
}

const Syntax = "proto3"

type Service struct {
	Name      string
	Package   string
	GoPackage string
	Syntax    string
	// go-doudou version
	Version  string
	ProtoVer string
	Rpcs     []Rpc
	Messages []Message
	Enums    []Enum
	Comments []string
	Imports  []string
}

func (receiver ProtoGenerator) NewService(name, goPackage string) Service {
	return Service{
		Name:      strcase.ToCamel(name) + "Service",
		Package:   strcase.ToSnake(name),
		GoPackage: goPackage,
		Syntax:    Syntax,
		Version:   version.Release,
		ProtoVer:  fmt.Sprintf("v%s", time.Now().Local().Format(constants.FORMAT10)),
	}
}

type StreamType int

const (
	biStream = iota + 1
	clientStream
	serverStream
)

type Rpc struct {
	Name       string
	Request    Message
	Response   Message
	Comments   []string
	StreamType StreamType
}

func (receiver ProtoGenerator) NewRpc(method astutils.MethodMeta) Rpc {
	rpcName := strcase.ToCamel(method.Name) + "Rpc"
	rpcRequest := receiver.newRequest(rpcName, method.Params)
	if reflect.DeepEqual(rpcRequest, Empty) {
		ImportStore["google/protobuf/empty.proto"] = struct{}{}
	}
	if !strings.HasPrefix(rpcRequest.Name, "stream ") && !rpcRequest.IsImported {
		if _, ok := MessageStore[rpcRequest.Name]; !ok {
			MessageStore[rpcRequest.Name] = rpcRequest
		}
	}
	rpcResponse := receiver.newResponse(rpcName, method.Results)
	if reflect.DeepEqual(rpcResponse, Empty) {
		ImportStore["google/protobuf/empty.proto"] = struct{}{}
	}
	if !strings.HasPrefix(rpcResponse.Name, "stream ") && !rpcResponse.IsImported {
		if _, ok := MessageStore[rpcResponse.Name]; !ok {
			MessageStore[rpcResponse.Name] = rpcResponse
		}
	}
	var st StreamType
	if strings.HasPrefix(rpcRequest.Name, "stream ") && strings.HasPrefix(rpcResponse.Name, "stream ") {
		st = biStream
	} else if strings.HasPrefix(rpcRequest.Name, "stream ") {
		st = clientStream
	} else if strings.HasPrefix(rpcResponse.Name, "stream ") {
		st = serverStream
	}
	return Rpc{
		Name:       rpcName,
		Request:    rpcRequest,
		Response:   rpcResponse,
		Comments:   method.Comments,
		StreamType: st,
	}
}

func (receiver ProtoGenerator) newRequest(rpcName string, params []astutils.FieldMeta) Message {
	if len(params) == 0 {
		return Empty
	}
	if len(params) == 1 && params[0].Type == "context.Context" {
		return Empty
	}
	if params[0].Type == "context.Context" {
		params = params[1:]
	}
	if len(params) == 1 {
		if m, ok := receiver.MessageOf(params[0].Type).(Message); ok && m.IsTopLevel {
			if strings.HasPrefix(params[0].Name, "stream") {
				m.Name = "stream " + m.Name
			}
			return m
		}
	}
	var fields []Field
	for i, field := range params {
		fields = append(fields, receiver.newField(field, i+1))
	}
	return Message{
		Name:       strcase.ToCamel(rpcName + "Request"),
		Fields:     fields,
		IsTopLevel: true,
	}
}

func (receiver ProtoGenerator) newResponse(rpcName string, params []astutils.FieldMeta) Message {
	if len(params) == 0 {
		return Empty
	}
	if len(params) == 1 && params[0].Type == "error" {
		return Empty
	}
	if params[len(params)-1].Type == "error" {
		params = params[:len(params)-1]
	}
	if len(params) == 1 {
		if m, ok := receiver.MessageOf(params[0].Type).(Message); ok && m.IsTopLevel {
			if strings.HasPrefix(params[0].Name, "stream") {
				m.Name = "stream " + m.Name
			}
			return m
		}
	}
	var fields []Field
	for i, field := range params {
		fields = append(fields, receiver.newField(field, i+1))
	}
	return Message{
		Name:       strcase.ToCamel(rpcName + "Response"),
		Fields:     fields,
		IsTopLevel: true,
	}
}
