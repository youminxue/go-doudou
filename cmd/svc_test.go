package cmd_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/youminxue/odin/cmd"
	"testing"
)

func Test_svcCmd(t *testing.T) {
	Convey("Should not panic when run svc command", t, func() {
		So(func() {
			ExecuteCommandC(cmd.GetRootCmd(), []string{"svc"}...)
		}, ShouldNotPanic)
	})
}
