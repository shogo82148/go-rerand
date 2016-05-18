package rerand

import (
	"math/rand"
	"regexp/syntax"
	"testing"
	"time"
)

func TestGenerator(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	g, err := New("a{1,16}", syntax.Perl, nil)
	if err != nil {
		t.Error(err)
	}
	t.Log(g.Generate())
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
