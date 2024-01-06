package main

import (
	"fmt"
	"strconv"
	"sync"
)


var poolMap = make(map[int]*sync.Pool)

func GetSlice(digitsLen int) *[]Result {
	pool := poolMap[digitsLen]
	if pool == nil {
		pool = &sync.Pool{
			New: func() interface{} {
				return MakeResultSlice(digitsLen)
			},
		}
		poolMap[digitsLen] = pool
	}
	return pool.Get().(*[]Result)  // type assertion
}

func PutSlice(digitsLen int, slice *[]Result) {
	// https://blog.mike.norgate.xyz/unlocking-go-slice-performance-navigating-sync-pool-for-enhanced-efficiency-7cb63b0b453e
	sl := *slice
	sl = sl[:0]
	*slice = sl
	pool := poolMap[digitsLen]
	if pool == nil {
		panic("Put before Get")
	}
	pool.Put(slice)
}

type Result struct {
	value int
	repr string
}

func MakeResultSlice(digitsLen int) *[]Result {
	scale, _ := opRaise.eval(10, digitsLen - 1)
	resultSlice := make([]Result, 0, 2 * scale)
	return &resultSlice
}

var poolThreshold = -1

func Generate(digits string) <-chan *[]Result {
	out := make(chan *[]Result)
	go func() {
		defer close(out)
		var termsPointer *[]Result
		if len(digits) > poolThreshold {
			termsPointer = GetSlice(len(digits))
		} else {
			termsPointer = MakeResultSlice(len(digits))
		}
		terms := *termsPointer
		value := Atoi(digits)
		terms = append(terms, Result{value, digits})
		terms = append(terms, Result{-value, "-" + digits})
		var fstTerms, sndTerms *[]Result
		for splitIdx := len(digits) - 1; splitIdx > 0; splitIdx-- {  // reverse iteration so left->right terms are generated first
			fstDigits := digits[:splitIdx]
			sndDigits := digits[splitIdx:]
			fstTerms = <- Generate(fstDigits)
			sndTerms = <- Generate(sndDigits)
			for _, fst := range *fstTerms {
				for _, snd := range *sndTerms {
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
			if len(fstDigits) > poolThreshold {
				PutSlice(len(fstDigits), fstTerms)
			}
			if len(sndDigits) > poolThreshold {
				PutSlice(len(sndDigits), sndTerms)
			}
		}
		*termsPointer = terms
		out <- termsPointer
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
