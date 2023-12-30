package main

import "fmt"

func genAdd(fst, snd int) <-chan Result {
	out := make(chan Result)
	go func() {
		if !(fst > 1_000_000 && snd > 1_000_000 || fst < -1_000_000 && snd < -1_000_000) {
			out <- Result{fst + snd, fmt.Sprintf("%d+%d", fst, snd)}
		}
		close(out)
	}()
	return out
}

func genMultiply(fst, snd int) <-chan Result {
	out := make(chan Result)
	go func() {
		if !(snd < 0 || fst > 1_000 && snd > 1_000 || fst < -1_000 && snd < -1_000 || fst > 1_000_000 || fst < -1_000_000 || snd > 1_000_000 || snd < -1_000_000) {
			out <- Result{fst * snd, fmt.Sprintf("%d*%d", fst, snd)}
		}
		close(out)
	}()
	return out
}
func genDivide(fst, snd int) <-chan Result {
	out := make(chan Result)
	go func() {
		if snd > 0 && fst % snd == 0 {
			out <- Result{fst / snd, fmt.Sprintf("%d/%d", fst, snd)}
		}
		close(out)
	}()
	return out
}
func genRaise(fst, snd int) <-chan Result {
	out := make(chan Result)
	go func() {
		if fst >= 0 && snd >= 0 && snd < 100 {
			result := 1
			for i := 0; i < snd; i++ {
				result *= fst
				if result > 1_000_000 {
					close(out)
					return
				}
			}
			repr := fmt.Sprintf("%d^%d", fst, snd)
			out <- Result{result, repr}
			out <- Result{-result, "-" + repr}
		}
		close(out)
	}()
	return out
}
