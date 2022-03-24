package common_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/internal/common"
)

func TestSleep(test *testing.T) {
	const dt = 50 * time.Millisecond
	t := func(name string, ctx context.Context, expected error) {
		test.Run(name, func(test *testing.T) {
			err := common.Sleep(ctx, dt)
			if !errors.Is(err, expected) {
				test.Errorf("expected %v, got %v", expected, err)
			}
		})
	}

	ctx := context.Background()
	canceled, cancel := context.WithCancel(ctx)
	cancel()

	t("background context", ctx, nil)
	t("canceled context", canceled, context.Canceled)
}
