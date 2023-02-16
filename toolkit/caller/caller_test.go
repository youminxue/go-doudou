package caller_test

import (
	"fmt"
	"github.com/youminxue/odin/toolkit/caller"
	"testing"
)

func TestCaller_String(t *testing.T) {
	c := caller.NewCaller()
	fmt.Println(c.String())
}
