package utils

import (
	"github.com/hashicorp/go-sockaddr"
	"github.com/youminxue/odin/framework/internal/config"
	"github.com/youminxue/odin/toolkit/stringutils"
	"github.com/youminxue/odin/toolkit/zlogger"
)

var GetPrivateIP = sockaddr.GetPrivateIP

func GetRegisterHost() string {
	registerHost := config.DefaultGddRegisterHost
	if stringutils.IsNotEmpty(config.GddRegisterHost.Load()) {
		registerHost = config.GddRegisterHost.Load()
	}
	if stringutils.IsEmpty(registerHost) {
		var err error
		registerHost, err = GetPrivateIP()
		if err != nil {
			zlogger.Panic().Err(err).Msg("[odin] failed to get interface addresses")
		}
		if stringutils.IsEmpty(registerHost) {
			zlogger.Panic().Msg("[odin] no private IP address found, and explicit IP not provided")
		}
	}
	return registerHost
}
