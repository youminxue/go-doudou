package codegen

import (
	"github.com/youminxue/odin/toolkit/astutils"
	"os"
	"path/filepath"
	"testing"
)

func TestGenGoClientProxy1(t *testing.T) {
	dir := testDir + "clientproxy"
	InitSvc(dir)
	defer os.RemoveAll(dir)
	svcfile := filepath.Join(testDir, "svc.go")
	ic := astutils.BuildInterfaceCollector(svcfile, astutils.ExprString)

	type args struct {
		dir string
		ic  astutils.InterfaceCollector
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				dir: dir,
				ic:  ic,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GenGoClientProxy(tt.args.dir, tt.args.ic)
		})
	}
}
