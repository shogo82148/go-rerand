package main

import (
	"testing"

	"github.com/t-mrt/gocha"
)

func BenchmarkGocha(b *testing.B) {
	cases := []struct {
		name   string
		regexp string
	}{
		{`coffeescript`, `[カコヵか][ッー]{1,3}?[フヒふひ]{1,3}[ィェー]{1,3}[ズス][ドクグュ][リイ][プブぷぶ]{1,3}[トドォ]{1,2}`},
		{``, `[あ-お]{10}`},
	}

	for _, c := range cases {
		_, g := gocha.New(c.regexp)
		name := c.name
		if name == "" {
			name = c.regexp
		}
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				g.Gen()
			}
		})
	}
}
