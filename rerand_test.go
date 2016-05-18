package rerand

import (
	"math/rand"
	"regexp"
	"regexp/syntax"
	"testing"
)

func TestGenerator(t *testing.T) {
	rand.Seed(1)
	in := []string{
		`abc`,
		`[abc]`,
		`[a-zA-Z0-9][a-zA-Z0-9]`,
		`a{1,16}`,
		`[a-z]{1,16}`,
		`abc|def`,
		`abc|def|ghi`,
		`abc(def|ghi)`,
	}

	for _, pattern := range in {
		re := regexp.MustCompile(pattern)
		g, err := New(pattern, syntax.Perl, rand.New(rand.NewSource(1)))
		if err != nil {
			t.Errorf("unexpected error: %v in %s", err, pattern)
			continue
		}
		for i := 0; i < 10000; i++ {
			s, err := g.Generate()
			if err != nil {
				t.Errorf("unexpected error: %v in %s", err, pattern)
			}
			if !re.MatchString(s) {
				t.Errorf(`generated string "%s" does not match "%s"`, s, pattern)
			}
		}
	}
}

func TestGeneratorDistribution(t *testing.T) {
	const RuneNum = 100000
	const AllowError = 2000
	rand.Seed(1)
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
		g, err := New(c.pattern, syntax.Perl, rand.New(rand.NewSource(1)))
		if err != nil {
			t.Errorf("unexpected error: %v in %s", err, c.pattern)
			continue
		}
		count := map[string]int{}
		for i := 0; i < RuneNum*c.num; i++ {
			s, err := g.Generate()
			if err != nil {
				t.Errorf("unexpected error: %v in %s", err, c.pattern)
			}
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
	rand.Seed(1)
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

		g := NewRuneGenerator(runes, nil)
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
