# go-rerand

[![GoDoc](https://godoc.org/github.com/shogo82148/go-rerand?status.svg)](http://godoc.org/github.com/shogo82148/go-rerand)

## NAME

rerand - Generate random strings based on regular expressions.

## DESCRIPTION

rerand makes it trivial to generate random strings.
You can use the same regular expression as [regexp/syntax](https://golang.org/pkg/regexp/syntax/).

## SYNOPSIS

### Programable Interface

``` go
import (
    "fmt"
    "log"
    "regexp/syntax"

    "github.com/shogo82148/go-rerand"
)

func main() {
    g, err := rerand.New(`\d{2,3}-\d{3,4}-\d{3,4}`, syntax.Perl, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(g.Generate())
    // 403-165-405

    fmt.Println(g.Generate())
    // 093-0033-3349
}
```

### Command Line Interface

``` plain
Usage of rerand:
  -d	distinct runes
  -distinct-runes
    	distinct runes
  -h	show help message
  -help
    	show help message
  -n int
    	the number of random strings (default 1)
  -number int
    	the number of random strings (default 1)
  -p float
    	the probability for AltInst
  -prob float
    	the probability for AltInst
```

## INSTALLATION

``` bash
go get github.com/shogo82148/go-rerand/cmd/rerand
```

## AUTHOR

Ichinose Shogo<shogo82148@gmail.com>

## LICENSE

MIT license. See LICENSE.md for more detail.

## SEE ALSO

- [Songmu/strrand](https://github.com/Songmu/strrand)
- [t-mrt/gocha](https://github.com/t-mrt/gocha)
