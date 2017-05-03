package rerand

import (
	"math"
	"math/rand"
	"regexp"
	"regexp/syntax"
	"testing"
)

func TestError(t *testing.T) {
	if _, err := New(`[a-z`, syntax.Perl, nil); err == nil {
		t.Error("want syntax error, got nil")
	}

	if _, err := New(`[a-z]*`, syntax.Perl, nil); err != ErrTooManyRepeat {
		t.Errorf("want too many repeat error, got %v", err)
	}
}

func TestGenerator(t *testing.T) {
	in := []string{
		`abc`,
		`[abc]`,
		`[a-zA-Z0-9][a-zA-Z0-9]`,
		`a{1,16}`,
		`[a-z]{1,16}`,
		`abc|def`,
		`abc|def|ghi`,
		`abc(def|ghi)`,
		`[[:alpha:]]`,
		`\pN`,
		`\p{Greek}`,
	}

	test := func(g *Generator, re *regexp.Regexp, pattern string) {
		for i := 0; i < 10000; i++ {
			s := g.Generate()
			if !re.MatchString(s) {
				t.Errorf(`generated string "%s" does not match "%s"`, s, pattern)
			}
		}
	}

	for _, pattern := range in {
		re := regexp.MustCompile(pattern)
		test(Must(New(pattern, syntax.Perl, rand.New(rand.NewSource(1)))), re, pattern)
		test(Must(NewDistinctRunes(pattern, syntax.Perl, rand.New(rand.NewSource(1)))), re, pattern)
		test(Must(NewWithProbability(pattern, syntax.Perl, rand.New(rand.NewSource(1)), math.MaxInt64/2)), re, pattern)
	}
}

func TestGeneratorDistinctRunesDistribution(t *testing.T) {
	const RuneNum = 100000
	const AllowError = 2000
	in := []struct {
		pattern string
		num     int
	}{
		{`abc`, 1},
		{`[abc]`, 3},
		{`a{1,16}`, 16},
		{`[ab]{1,3}`, 2 + 2*2 + 2*2*2},
		{`abc|def`, 2},
		{`abc|def|ghi`, 3},
		{`[abc]|def`, 4},
		{`[あいうえお]{2}`, 5 * 5},
	}

	for _, c := range in {
		g, err := NewDistinctRunes(c.pattern, syntax.Perl, rand.New(rand.NewSource(1)))
		if err != nil {
			t.Errorf("unexpected error: %v in %s", err, c.pattern)
			continue
		}
		count := map[string]int{}
		for i := 0; i < RuneNum*c.num; i++ {
			s := g.Generate()
			count[s] = count[s] + 1
		}
		if len(count) != c.num {
			t.Errorf("want %d, got %d in %s", c.num, len(count), c.pattern)
		}
		for s, cc := range count {
			if cc < RuneNum-AllowError || cc > RuneNum+AllowError {
				t.Errorf("incorrect count of '%s'(%d) in %s", s, cc, c.pattern)
			}
		}
	}
}

func TestRuneGenerator(t *testing.T) {
	const RuneNum = 100000
	const AllowError = 2000
	in := [][]rune{
		{'a'},
		{'a', 'a'},
		{'a', 'z'},
		{'a', 'z', 'A', 'A'},
		{'a', 'z', 'A', 'Z', '0', '9'},
	}

	for _, runes := range in {
		num := 0
		if len(runes) == 1 {
			num = 1
		} else {
			for i := 0; i < len(runes); i += 2 {
				num += int(runes[i+1] - runes[i] + 1)
			}
		}

		g := NewRuneGenerator(runes, rand.New(rand.NewSource(1)))
		count := map[rune]int{}
		for i := 0; i < RuneNum*num; i++ {
			r := g.Generate()
			count[r] = count[r] + 1
		}

		for r, c := range count {
			if c < RuneNum-AllowError || c > RuneNum+AllowError {
				t.Errorf("%+v: incorrect count of '%c'(%d)", runes, r, c)
			}
		}
	}
}

func BenchmarkGenerator(b *testing.B) {
	cases := []struct {
		name   string
		regexp string
	}{
		{`coffeescript`, `[カコヵか][ッー]{1,3}?[フヒふひ]{1,3}[ィェー]{1,3}[ズス][ドクグュ][リイ][プブぷぶ]{1,3}[トドォ]{1,2}`},
		{``, `[あ-お]{10}`},
		{``, `[[:alpha:]]`},
		{``, `\S`},
		{``, `\pN`},
		{``, `\p{Greek}`},
	}

	for _, c := range cases {
		g, _ := New(c.regexp, syntax.Perl, rand.New(rand.NewSource(1)))
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

func BenchmarkRuneGenerator(b *testing.B) {
	cases := []struct {
		name  string
		runes []rune
	}{
		{"[a]", []rune{'a'}},
		{"[a-z]", []rune{'a', 'z'}},
		{"[a-zA-Z0-9]", []rune{'a', 'z', 'A', 'Z', '0', '9'}},
	}

	for _, c := range cases {
		g := NewRuneGenerator(c.runes, rand.New(rand.NewSource(1)))
		b.Run(c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				g.Generate()
			}
		})
	}
}
