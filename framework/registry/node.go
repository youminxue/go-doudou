package registry

import (
	"github.com/youminxue/odin/framework/internal/config"
	"github.com/youminxue/odin/framework/registry/constants"
	"github.com/youminxue/odin/framework/registry/etcd"
	"github.com/youminxue/odin/framework/registry/memberlist"
	"github.com/youminxue/odin/framework/registry/nacos"
	logger "github.com/youminxue/odin/toolkit/zlogger"
)

type IServiceProvider interface {
	SelectServer() string
}

func NewRest(data ...map[string]interface{}) {
	for mode, _ := range config.ServiceDiscoveryMap() {
		switch mode {
		case constants.SD_NACOS:
			nacos.NewRest(data...)
		case constants.SD_ETCD:
			etcd.NewRest(data...)
		case constants.SD_MEMBERLIST:
			memberlist.NewRest(data...)
		default:
			logger.Warn().Msgf("[odin] unknown service discovery mode: %s", mode)
		}
	}
}

func NewGrpc(data ...map[string]interface{}) {
	for mode, _ := range config.ServiceDiscoveryMap() {
		switch mode {
		case constants.SD_NACOS:
			nacos.NewGrpc(data...)
		case constants.SD_ETCD:
			etcd.NewGrpc(data...)
		case constants.SD_MEMBERLIST:
			memberlist.NewGrpc(data...)
		default:
			logger.Warn().Msgf("[odin] unknown service discovery mode: %s", mode)
		}
	}
}

func ShutdownRest() {
	for mode, _ := range config.ServiceDiscoveryMap() {
		switch mode {
		case constants.SD_NACOS:
			nacos.ShutdownRest()
		case constants.SD_ETCD:
			etcd.ShutdownRest()
		case constants.SD_MEMBERLIST:
			memberlist.Shutdown()
		default:
			logger.Warn().Msgf("[odin] unknown service discovery mode: %s", mode)
		}
	}
}

func ShutdownGrpc() {
	for mode, _ := range config.ServiceDiscoveryMap() {
		switch mode {
		case constants.SD_NACOS:
			nacos.ShutdownGrpc()
		case constants.SD_ETCD:
			etcd.ShutdownGrpc()
		case constants.SD_MEMBERLIST:
			memberlist.Shutdown()
		default:
			logger.Warn().Msgf("[odin] unknown service discovery mode: %s", mode)
		}
	}
}
