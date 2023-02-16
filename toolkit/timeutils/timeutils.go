package timeutils

import (
	"context"
	"github.com/hyperjumptech/jiffy"
	"github.com/pkg/errors"
	"github.com/youminxue/odin/toolkit/stringutils"
	"time"
)

// Parse parses string to time.Duration
func Parse(t string, defaultDur time.Duration) (time.Duration, error) {
	var (
		dur time.Duration
		err error
	)
	if stringutils.IsNotEmpty(t) {
		if dur, err = jiffy.DurationOf(t); err != nil {
			err = errors.Wrapf(err, "parse %s from config file fail, use default 15s instead", t)
		}
	}
	if dur <= 0 {
		dur = defaultDur
	}
	return dur, err
}

func CallWithCtx(ctx context.Context, fn func() struct{}) error {
	result := make(chan struct{}, 1)
	go func() {
		result <- fn()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-result:
		return nil
	}
}
