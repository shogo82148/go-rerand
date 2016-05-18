package rerand

import (
	"errors"
	"log"
	"math/rand"
	"regexp/syntax"
)

var (
	ErrTooFewRepeat  = errors.New("Counted too few repeat.")
	ErrTooManyRepeat = errors.New("Counted too many repeat.")
)

type Generator struct {
	pattern  string
	prog     *syntax.Prog
	inst     []myinst
	min, max int
	rand     *rand.Rand
}

type myinst struct {
	syntax.Inst
	runeGenerator *RuneGenerator
}

func New(pattern string, flags syntax.Flags, r *rand.Rand) (*Generator, error) {
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

	inst := make([]myinst, len(prog.Inst))
	for i, in := range prog.Inst {
		in2 := myinst{Inst: in}
		if in.Op == syntax.InstRune {
			in2.runeGenerator = NewRuneGenerator(in.Rune, r)
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
	}
	return gen, nil
}

func (g *Generator) String() string {
	return g.pattern
}

func (g *Generator) Generate() (string, error) {
	inst := g.inst
	pc := uint32(g.prog.Start)
	i := inst[pc]
	result := []rune{}
	cap := []uint32{}

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
			pc = g.randPath(i.Out, i.Arg, cap)
			i = inst[pc]
		case syntax.InstCapture:
			cap = append(cap, pc)
			if len(cap) > (g.max+1)*2 {
				return string(result), ErrTooManyRepeat
			}
			pc = g.randPath(i.Out, i.Arg, cap)
			i = inst[pc]
		case syntax.InstMatch:
			if g.prog.NumCap > 2 && len(cap) < g.min*2 {
				return string(result), ErrTooFewRepeat
			}
			return string(result), nil
		}
	}
}

func (g *Generator) randPath(out, arg uint32, cap []uint32) uint32 {
	if rand.Intn(356)%2 == 0 {
		if len(cap) > 0 && out > cap[len(cap)-1] {
			return out
		}
		return arg
	}
	if len(cap) > 0 && arg > cap[len(cap)-1] {
		return arg
	}
	return out
}

type RuneGenerator struct {
	aliases []int
	probs   []int64
	sum     int64
	runes   []rune
	rand    *rand.Rand
}

func NewRuneGenerator(runes []rune, r *rand.Rand) *RuneGenerator {
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

func (g *RuneGenerator) Generate() rune {
	if len(g.runes) == 1 {
		return g.runes[0]
	}

	i := 0
	if len(g.runes) > 2 {
		var v int64
		if g.rand != nil {
			i = g.rand.Intn(len(g.probs))
			v = g.rand.Int63n(g.sum)
		} else {
			i = rand.Intn(len(g.probs))
			v = rand.Int63n(g.sum)
		}
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
	if g.rand != nil {
		randi += g.rand.Intn(max - min + 1)
	} else {
		randi += rand.Intn(max - min + 1)
	}
	return rune(randi)
}
