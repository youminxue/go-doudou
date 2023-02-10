package codegen

import (
	"fmt"
	"github.com/youminxue/v2/toolkit/astutils"
	"github.com/youminxue/v2/toolkit/copier"
	"path/filepath"
	"testing"
)

func Test_unimplementedMethods(t *testing.T) {
	ic := astutils.BuildInterfaceCollector(filepath.Join(testDir, "svc.go"), astutils.ExprString)
	var meta astutils.InterfaceMeta
	_ = copier.DeepCopy(ic.Interfaces[0], &meta)
	unimplementedMethods(&meta, filepath.Join(testDir, "transport/httpsrv"))
	fmt.Println(len(meta.Methods))
}
