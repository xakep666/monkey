package monkey

import (
	"errors"
	"runtime"
	"testing"
	"time"
)

func TestCyclicReplacementDetection(t *testing.T) {
	testNow := func() time.Time {
		return time.Date(2022, 1, 2, 3, 4, 5, 6, time.UTC)
	}
	testNowOther := func() time.Time {
		return time.Date(2021, 1, 2, 3, 4, 5, 6, time.UTC)
	}

	err := NewPatcher().
		Apply(func(patcher *Patcher) {
			RegisterReplacement(patcher, time.Now, testNow)
			RegisterReplacement(patcher, testNow, testNowOther)
			RegisterReplacement(patcher, runtime.GOMAXPROCS, func(int) int {
				return -10
			})
			RegisterReplacement(patcher, testNowOther, time.Now)
		}).
		detectCyclicReplacements()

	if !errors.Is(err, ErrCyclicReplacement) {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestRegisterNotFunction(t *testing.T) {
	err := NewPatcher().
		Apply(func(patcher *Patcher) {
			RegisterReplacement(patcher, runtime.GOARCH, "aaa")
		}).
		PatchAndExec()

	if !errors.Is(err, ErrFunctionNotFound) {
		t.Errorf("Unexpected error: %s", err)
	}
}
