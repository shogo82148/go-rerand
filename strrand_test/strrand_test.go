package main

import (
	"testing"

	"github.com/Songmu/strrand"
)

func BenchmarkStrRand(b *testing.B) {
	cases := []struct {
		name   string
		regexp string
	}{
		{`coffeescript`, `[カコヵか][ッー]{1,3}?[フヒふひ]{1,3}[ィェー]{1,3}[ズス][ドクグュ][リイ][プブぷぶ]{1,3}[トドォ]{1,2}`},
		{``, `[あ-お]{10}`},
		{``, `\S`},
		{``, `\S{10}`},
		{`telephone`, `\d{2,3}-\d{3,4}-\d{3,4}`},
	}

	for _, c := range cases {
		g, _ := strrand.New().CreateGenerator(c.regexp)
		name := c.name
		if name == "" {
			name = c.regexp
		}
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				g.Generate()
			}
		})
	}
}
