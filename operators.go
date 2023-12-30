package main

type Operator struct {
	symbol rune
	eval func(int, int) (int, bool)
}

var (
	opAdd        = Operator{'+', func(fst, snd int) (int, bool) {
		if fst > 1_000_000 && snd > 1_000_000 || fst < -1_000_000 && snd < -1_000_000 {
			return 0, false
		}
		return fst + snd, true
	}}
	opMultiply   = Operator{'*', func(fst, snd int) (int, bool) {
		if fst > 1_000 && snd > 1_000 || fst < -1_000 && snd < -1_000 || fst > 1_000_000 || fst < -1_000_000 || snd > 1_000_000 || snd < -1_000_000 {
			return 0, false
		}
		if snd < 0 {
			return 0, false
		}
		return fst * snd, true
	}}
	opDivide     = Operator{'/', func(fst, snd int) (int, bool) {
		if snd < 0 {
			return 0, false
		}
		if snd == 0 {
			return 0, false
		}
		if fst % snd != 0 {
			return 0, false
		}
		return fst / snd, true
	}}
	opRaise      = Operator{'^', func(fst, snd int) (int, bool) {
		if fst < 0 {
			return 0, false
		}
		if snd < 0 {
			return 0, false
		}
		if snd > 100 {
			return 0, false
		}
		result := 1
		for i := 0; i < snd; i++ {
			result *= fst
			if result > 1_000_000 {
				return 0, false
			}
		}
		return result, true
	}}
)
