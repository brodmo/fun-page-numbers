package main

import "errors"

type Operator struct {
	symbol rune
	eval func(int, int) (int, error)
}

var (
	opAdd        = Operator{'+', func(fst, snd int) (int, error) {
		if fst > 1_000_000 && snd > 1_000_000 || fst < -1_000_000 && snd < -1_000_000 {
			return 0, errors.New("result too big")
		}
		return fst + snd, nil
	}}
	opSubtract   = Operator{'-', func(fst, snd int) (int, error) {
		return fst - snd, nil
	}}
	opMultiply   = Operator{'*', func(fst, snd int) (int, error) {
		if fst > 1_000 && snd > 1_000 || fst < -1_000 && snd < -1_000 || fst > 1_000_000 || fst < -1_000_000 || snd > 1_000_000 || snd < -1_000_000 {
			return 0, errors.New("result too big")
		}
		if snd < 0 {
			return 0, errors.New("duplicate negative")
		}
		return fst * snd, nil
	}}
	opDivide     = Operator{'/', func(fst, snd int) (int, error) {
		if snd < 0 {
			return 0, errors.New("duplicate negative")
		}
		if snd == 0 {
			return 0, errors.New("divide by zero")
		}
		if fst % snd != 0 {
			return 0, errors.New("non-zero remainder")
		}
		return fst / snd, nil
	}}
	opRaise      = Operator{'^', func(fst, snd int) (int, error) {
		if fst < 0 {
			return 0, errors.New("duplicate negative")
		}
		if snd < 0 {
			return 0, errors.New("negative exponent")
		}
		if snd > 100 {
			return 0, errors.New("exponent too big")
		}
		result := 1
		for i := 0; i < snd; i++ {
			result *= fst
			if result > 1_000_000 {return 0, errors.New("result too big")}
		}
		return result, nil
	}}
)
