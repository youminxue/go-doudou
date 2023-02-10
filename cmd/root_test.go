package cmd_test

import (
	"github.com/youminxue/v2/cmd"
	"testing"
	"time"
)

func TestRootCmd(t *testing.T) {
	go cmd.GetRootCmd().Run(nil, nil)
	time.Sleep(2 * time.Second)
}
