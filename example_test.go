package rerand_test

import (
	"fmt"
	"math/rand"
	"regexp/syntax"

	rerand "github.com/shogo82148/go-rerand"
)

func ExampleGenerator_Generate() {
	r := rand.New(rand.NewSource(1))
	g := rerand.Must(rerand.New(`\d{2,3}-\d{3,4}-\d{3,4}`, syntax.Perl, r))
	fmt.Println(g.Generate())
	fmt.Println(g.Generate())
	// Output:
	// 17-9180-604
	// 29-4156-568
}
