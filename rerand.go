package rerand

import (
	"errors"
	"log"
	"math"
	"math/big"
	"math/rand"
	"regexp/syntax"
	"sync"
	"time"
)

// ErrTooManyRepeat the error used for New.
var ErrTooManyRepeat = errors.New("rerand: counted too many repeat")

// Generator is random string generator
type Generator struct {
	pattern  string
	prog     *syntax.Prog
	inst     []myinst
	min, max int
	rand     *rand.Rand
	runes    *sync.Pool
}

type myinst struct {
	syntax.Inst
	runeGenerator *RuneGenerator
	x, y          *big.Int
}

// Must is a helper that wraps a call to a function returning (*Generator, error) and panics if the error is non-nil.
func Must(g *Generator, err error) *Generator {
	if err != nil {
		panic(err)
	}
	return g
}

// New returns new Generator.
func New(pattern string, flags syntax.Flags, r *rand.Rand) (*Generator, error) {
	return newGenerator(pattern, flags, r, false, 0)
}

// NewDistinctRunes returns new Generator.
func NewDistinctRunes(pattern string, flags syntax.Flags, r *rand.Rand) (*Generator, error) {
	return newGenerator(pattern, flags, r, true, 0)
}

// NewWithProbability returns new Generator.
func NewWithProbability(pattern string, flags syntax.Flags, r *rand.Rand, prob int64) (*Generator, error) {
	return newGenerator(pattern, flags, r, false, prob)
}

func newGenerator(pattern string, flags syntax.Flags, r *rand.Rand, distinctRunes bool, prob int64) (g *Generator, err error) {
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	re, err := syntax.Parse(pattern, flags)
	if err != nil {
		return nil, err
	}
	min := re.Min
	max := re.Max
	re = re.Simplify()
	prog, err := syntax.Compile(re)
	if err != nil {
		return nil, err
	}

	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if e == ErrTooManyRepeat {
			err = ErrTooManyRepeat
			return
		}
		panic(err)
	}()

	cache := make([]*big.Int, len(prog.Inst))
	visitied := make([]bool, len(prog.Inst))
	var count func(i uint32) *big.Int
	count = func(i uint32) *big.Int {
		if visitied[i] {
			panic(ErrTooManyRepeat)
		}
		if cache[i] != nil {
			return cache[i]
		}

		visitied[i] = true
		var ret *big.Int
		switch prog.Inst[i].Op {
		default:
			ret = big.NewInt(0)
		case syntax.InstRune:
			if !distinctRunes {
				ret = count(prog.Inst[i].Out)
				break
			}
			var sum int64
			runes := prog.Inst[i].Rune
			if len(runes) == 1 {
				sum = 1
			} else {
				for i := 0; i < len(runes); i += 2 {
					sum += int64(runes[i+1] - runes[i] + 1)
				}
			}
			ret = big.NewInt(sum)
			ret.Mul(ret, count(prog.Inst[i].Out))
		case syntax.InstRune1:
			ret = count(prog.Inst[i].Out)
		case syntax.InstAlt:
			ret = big.NewInt(0)
			ret.Add(count(prog.Inst[i].Arg), count(prog.Inst[i].Out))
		case syntax.InstCapture:
			ret = count(prog.Inst[i].Out)
		case syntax.InstMatch:
			ret = big.NewInt(1)
		}
		cache[i] = ret
		visitied[i] = false
		return ret
	}
	inst := make([]myinst, len(prog.Inst))
	for i, in := range prog.Inst {
		in2 := myinst{Inst: in}
		switch in.Op {
		case syntax.InstRune:
			in2.runeGenerator = NewRuneGenerator(in.Rune, r)
		case syntax.InstAlt:
			if prob == 0 {
				in2.x = count(in.Out)
				in2.y = count(uint32(i))
			} else {
				in2.x = big.NewInt(prob)
				in2.y = big.NewInt(math.MaxInt64)
			}
		}
		inst[i] = in2
	}

	gen := &Generator{
		pattern: pattern,
		prog:    prog,
		inst:    inst,
		min:     min,
		max:     max,
		rand:    r,
		runes: &sync.Pool{
			New: func() interface{} { return []rune{} },
		},
	}
	return gen, nil
}

func (g *Generator) String() string {
	return g.pattern
}

// Generate generates a random string
func (g *Generator) Generate() string {
	inst := g.inst
	pc := uint32(g.prog.Start)
	i := inst[pc]
	result := g.runes.Get().([]rune)[:0]

	for {
		switch i.Op {
		default:
			log.Fatalf("%v: %v", i.Op, "bad operation")
		case syntax.InstFail:
			// nothing
		case syntax.InstRune:
			result = append(result, i.runeGenerator.Generate())
			pc = i.Out
			i = inst[pc]
		case syntax.InstRune1:
			result = append(result, i.Rune[0])
			pc = i.Out
			i = inst[pc]
		case syntax.InstAlt:
			a := big.NewInt(0)
			a.Rand(g.rand, i.y)
			if a.Cmp(i.x) < 0 {
				pc = i.Out
			} else {
				pc = i.Arg
			}
			i = inst[pc]
		case syntax.InstCapture:
			pc = i.Out
			i = inst[pc]
		case syntax.InstMatch:
			strresult := string(result)
			g.runes.Put(result)
			return strresult
		}
	}
}

// RuneGenerator is random rune generator.
type RuneGenerator struct {
	aliases []int
	probs   []int64
	sum     int64
	runes   []rune
	rand    *rand.Rand
}

// NewRuneGenerator returns new RuneGenerator.
func NewRuneGenerator(runes []rune, r *rand.Rand) *RuneGenerator {
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	if len(runes) <= 2 {
		return &RuneGenerator{
			runes: runes,
			rand:  r,
		}
	}

	pairs := len(runes) / 2
	aliases := make([]int, pairs)
	probs := make([]int64, pairs)

	// calculate weights and normalize them
	var sum int64
	for i := 0; i < pairs; i++ {
		aliases[i] = i
		w := int64(runes[i*2+1] - runes[i*2] + 1)
		probs[i] = w * int64(pairs)
		sum += w
	}

	// Walker’s alias method
	hl := make([]int, pairs)
	h := 0
	l := pairs - 1
	for i, p := range probs {
		if p > sum {
			hl[h] = i
			h++
		} else {
			hl[l] = i
			l--
		}
	}
	h--
	l++
	for h >= 0 && l < pairs {
		j := hl[l]
		k := hl[h]
		aliases[j] = k
		probs[k] += probs[j] - sum
		l++
		if probs[k] < sum {
			l--
			h--
			hl[l] = k
		}
	}

	return &RuneGenerator{
		aliases: aliases,
		probs:   probs,
		sum:     sum,
		runes:   runes,
		rand:    r,
	}
}

// Generate generates random rune.
func (g *RuneGenerator) Generate() rune {
	if len(g.runes) == 1 {
		return g.runes[0]
	}

	i := 0
	if len(g.runes) > 2 {
		i = g.rand.Intn(len(g.probs))
		v := g.rand.Int63n(g.sum)
		if g.probs[i] <= v {
			i = g.aliases[i]
		}
	}

	min := int(g.runes[2*i])
	max := int(g.runes[2*i+1])
	if min == max {
		return rune(min)
	}
	randi := min
	randi += g.rand.Intn(max - min + 1)
	return rune(randi)
}
