package cmd_test

import (
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/youminxue/v2/cmd"
	"github.com/youminxue/v2/cmd/internal/svc"
	"github.com/youminxue/v2/cmd/mock"
	"testing"
)

func Test_versionCmd_Yes(t *testing.T) {
	Convey("Should not panic and succeed to upgrade when run version command", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		prompt := mock.NewMockISelect(ctrl)
		prompt.
			EXPECT().
			Run().
			AnyTimes().
			Return(0, "Yes", nil)

		cmd.Prompt = prompt

		runner := mock.NewMockRunner(ctrl)
		runner.
			EXPECT().
			Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		cmd.VersionSvc = func(dir string, opts ...svc.SvcOption) svc.ISvc {
			return svc.NewSvc("", svc.WithRunner(runner))
		}

		cmd.LatestReleaseVerFunc = func() string {
			return "v999999.0.0"
		}
		defer func() {
			cmd.LatestReleaseVerFunc = cmd.LatestReleaseVer
		}()

		So(func() {
			ExecuteCommandC(cmd.GetRootCmd(), []string{"version"}...)
		}, ShouldNotPanic)
	})
}

func Test_versionCmd_Yes_Panic(t *testing.T) {
	Convey("Should panic and fail to upgrade when run version command", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		prompt := mock.NewMockISelect(ctrl)
		prompt.
			EXPECT().
			Run().
			AnyTimes().
			Return(0, "Yes", nil)

		cmd.Prompt = prompt

		runner := mock.NewMockRunner(ctrl)
		runner.
			EXPECT().
			Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(errors.New("mock runner error"))

		cmd.VersionSvc = func(dir string, opts ...svc.SvcOption) svc.ISvc {
			return svc.NewSvc("", svc.WithRunner(runner))
		}

		cmd.LatestReleaseVerFunc = func() string {
			return "v999999.0.0"
		}
		defer func() {
			cmd.LatestReleaseVerFunc = cmd.LatestReleaseVer
		}()

		So(func() {
			ExecuteCommandC(cmd.GetRootCmd(), []string{"version"}...)
		}, ShouldPanic)
	})
}
