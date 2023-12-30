package main

import (
	"fmt"
	"strconv"
)


type Result struct {
	value int
	repr string
}

func Generate(digits string) <-chan Result {
	out := make(chan Result)
	go func() {
		for splitIdx := range digits {
			if splitIdx == 0 {
				value := Atoi(digits)
				out <- Result{value, digits}
				out <- Result{-value, "-" + digits}
				continue
			}
			invSplitIdx := len(digits) - splitIdx
			var sndResultCache []Result  // speedup is marginal but real
			for fstResult := range Generate(digits[:invSplitIdx]) {
				if sndResultCache == nil {
					for sndResult := range Generate(digits[invSplitIdx:]) {
						sndResultCache = append(sndResultCache, sndResult)
					}
				}
				for _, sndResult := range sndResultCache {
					for _, op := range operators {
						value, err := op.eval(fstResult.value, sndResult.value)
						if err == nil {
							repr := fmt.Sprintf("(%s%c%s)", fstResult.repr, op.symbol, sndResult.repr)
							out <- Result{value, repr}
							if op.symbol == '^' {
								out <- Result{-value, "-" + repr}
							}
						}
					}
				}
			}
		}
		close(out)
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
