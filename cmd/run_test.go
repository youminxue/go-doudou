package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newSvc(t *testing.T) {
	s := newSvc()
	assert.NotNil(t, s)
}