package memberlist_test

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/youminxue/odin/toolkit/memberlist"
	memmock "github.com/youminxue/odin/toolkit/memberlist/mock"
	"testing"
)

func Test_weightBroadcast_Invalidates(t *testing.T) {
	msg1 := memberlist.NewWeightBroadcast("testNode", []byte("test weight message1"))
	ok := msg1.Invalidates(memberlist.NewWeightBroadcast("testNode", []byte("test weight message2")))
	require.True(t, ok)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	b := memmock.NewMockBroadcast(ctrl)
	ok = msg1.Invalidates(b)
	require.False(t, ok)

	require.Equal(t, "test weight message1", string(msg1.Message()))
	msg1.Finished()
}
