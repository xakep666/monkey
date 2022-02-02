//go:build integration

package monkey_test

import (
	"github.com/xakep666/monkey"
	"testing"
	"time"
)

type X struct{}

//go:noinline
func (X) Int() int { return 42 }

type Y interface {
	Z() string
}

func init() {
	monkey.NewPatcher().
		Apply(func(patcher *monkey.Patcher) {
			// works if not inlined
			monkey.RegisterReplacement(patcher, time.Now, func() time.Time {
				return time.Date(1980, 1, 2, 3, 4, 5, 6, time.UTC)
			})
			monkey.RegisterReplacement(patcher, X.Int, func(X) int {
				return 100500
			})
			// works only for "Y.Z(<impl>)" form
			monkey.RegisterReplacement(patcher, Y.Z, func(Y) string {
				return "xxx"
			})
		}).MustPatchAndExec()
}

func TestMonkey_Integration(t *testing.T) {
	if now := time.Now(); !now.Equal(time.Date(1980, 1, 2, 3, 4, 5, 6, time.UTC)) {
		t.Errorf("Time not patched, returned: %s", now)
	}

	if ret := Y.Z(nil); ret != "xxx" {
		t.Errorf("Nil interface call not patched, returned: %s", ret)
	}

	if ret := (X{}).Int(); ret != 100500 {
		t.Errorf("Method call not patched, returned: %d", ret)
	}
}
