package main

import (
	"fmt"
	"time"

	"github.com/xakep666/monkey"
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
				return time.Date(2022, 1, 2, 3, 4, 5, 6, time.UTC)
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

func main() {
	fmt.Println(time.Now())
	fmt.Println(X{}.Int())
	fmt.Println(Y.Z(nil))
}
