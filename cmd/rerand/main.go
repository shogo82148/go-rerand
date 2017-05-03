package main

import (
	"flag"
	"fmt"
	"log"
	"regexp/syntax"

	"math"

	rerand "github.com/shogo82148/go-rerand"
)

func main() {
	var n int
	var distinctRunes bool
	var prob float64
	var help bool
	flag.IntVar(&n, "n", 1, "the number of random strings")
	flag.IntVar(&n, "number", 1, "the number of random strings")
	flag.BoolVar(&distinctRunes, "d", false, "distinct runes")
	flag.BoolVar(&distinctRunes, "distinct-runes", false, "distinct runes")
	flag.Float64Var(&prob, "p", 0, "the probability for AltInst")
	flag.Float64Var(&prob, "prob", 0, "the probability for AltInst")
	flag.BoolVar(&help, "h", false, "show help message")
	flag.BoolVar(&help, "help", false, "show help message")
	flag.Parse()

	var g *rerand.Generator
	var err error
	if distinctRunes {
		g, err = rerand.NewDistinctRunes(flag.Arg(0), syntax.Perl, nil)
	} else if prob > 0 {
		if prob >= 1 {
			log.Fatal("prob must be less than 1")
		}
		g, err = rerand.NewWithProbability(flag.Arg(0), syntax.Perl, nil, int64(math.MaxInt64*prob))
	} else {
		g, err = rerand.New(flag.Arg(0), syntax.Perl, nil)
	}
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < n; i++ {
		fmt.Println(g.Generate())
	}
}
