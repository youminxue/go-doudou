package caller_test

import (
	"fmt"
	"github.com/youminxue/v2/toolkit/caller"
	"testing"
)

func TestCaller_String(t *testing.T) {
	c := caller.NewCaller()
	fmt.Println(c.String())
}
