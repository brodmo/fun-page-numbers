package main

import (
	"fmt"
	"strconv"
)

type Result struct {
	value int
	repr string
}

func Generate(digits string) <-chan []Result {
	out := make(chan []Result)
	go func() {
		defer close(out)
		scale, _ := opRaise.eval(10, len(digits) - 1)
		terms := make([]Result, 0, 2 * scale)
		value := Atoi(digits)
		terms = append(terms, Result{value, digits})
		terms = append(terms, Result{-value, "-" + digits})
		var fstTerms, sndTerms []Result
		for splitIdx := len(digits) - 1; splitIdx > 0; splitIdx-- {  // reverse iteration so left->right terms are generated first
			fstTerms = <- Generate(digits[:splitIdx])
			sndTerms = <- Generate(digits[splitIdx:])
			for _, fst := range fstTerms {
				for _, snd := range sndTerms {
					for _, op := range operators {
						value, valid := op.eval(fst.value, snd.value)
						if valid {
							repr := fmt.Sprintf("(%s%c%s)", fst.repr, op.symbol, snd.repr)
							terms = append(terms, Result{value, repr})
							if op.symbol == '^' {
								terms = append(terms, Result{-value, "-" + repr})
							}
						}
					}
				}
			}
		}
		out <- terms
	}()
	return out
}

func Atoi(digits string) int {
	integer, err := strconv.Atoi(digits)
	if err != nil {
		panic(err)
	}
	return integer	
}
