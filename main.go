package main

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// todo vast majority of generated terms are no good, fix by delegating to operations for next value
// mostly due to negations -> eliminate duplicate negatives
// todo parallelize

const DigitString = "4812018"
var baseTerms = []int {0, 1, 2, 5, 15, 51, 188, 731, 2950, 12235, 51822, 223191, 974427, 4302645, 19181100, 86211885, 390248055, 1777495635, 8140539950, 37463689775, 173164232965, 803539474345, 3741930523740, 17481709707825, 81912506777200, 384847173838501, 1812610804416698}
var operations = []*Operation{&opAdd, &opMultiply, &opDivide, &opRaise}
const UseNegation = true  // don't use opSubtract in this case!

type Term struct {  // todo delegate most to operation
	fst, snd  SubTerm
	operation *Operation
	negated bool
}

func NewTerm(fst, snd SubTerm) *Term {
	return &Term{fst, snd, nil, false}
}

func (term *Term) Values() <-chan int {
	out := make(chan int)
	go func() {
		for _, op := range operations {
			term.operation = op
			for fstValue := range term.fst.Values() {
				for sndValue := range term.snd.Values() {
					result, err := op.eval(fstValue, sndValue)
					if err == nil {
						out <- result
						if op.symbol == '^' {
							term.negated = true
							out <- -result
							term.negated = false
						}
					}
				}
			}
		}
		close(out)
	}()
	return out
}

func (term *Term) String() string {
	prefix := ""
	if term.negated {
		prefix = " -"
	}
	return fmt.Sprintf("%s( %s %c %s )", prefix, term.fst.String(), term.operation.symbol, term.snd.String())
}

func (term *Term) Compact() string {
	return strings.ReplaceAll(term.String(), " ", "")
}

type SubTerm interface {
	Values() <-chan int
	String() string
	Compact() string
}

type Integer struct {
	value int
}

func (number *Integer) Values() <-chan int {
	out := make(chan int)
	go func() {
		out <- number.value
		out <- -number.value
		close(out)
	}()
	return out
}

func (number *Integer) String() string {
	return strconv.Itoa(number.value)
}

func (number *Integer) Compact() string {
	return number.String()
}

type TermBuilder struct {
	digits string
}

func (builder *TermBuilder) Terms() <-chan SubTerm {
	out := make(chan SubTerm)
	go func() {
		for splitIdx := range builder.digits {
			if splitIdx == 0 {
				out <- &Integer{Atoi(builder.digits)}
				continue
			}
			invSplitIdx := len(builder.digits) - splitIdx
			fstBuilder := TermBuilder{builder.digits[:invSplitIdx]}
			sndBuilder := TermBuilder{builder.digits[invSplitIdx:]}
			for fstTerm := range fstBuilder.Terms() {
				for sndTerm := range sndBuilder.Terms() {
					out <- NewTerm(fstTerm, sndTerm)
				}
			}
		}
		close(out)
	}()
	return out
}

func Atoi(digits string) int {
	integer, _ := strconv.Atoi(digits)  // todo error handling
	return integer	
}


// multiple between? 1 * - 2? probably not: implicit 0

type Operation struct {
	symbol rune
	eval func(int, int) (int, error)
}

var (
	opAdd        = Operation{'+', func(fst, snd int) (int, error) {
		if fst > 1_000_000 && snd > 1_000_000 || fst < -1_000_000 && snd < -1_000_000 {
			return 0, errors.New("result too big")
		}
		return fst + snd, nil
	}}
	opSubtract   = Operation{'-', func(fst, snd int) (int, error) {
		return fst - snd, nil
	}}
	opMultiply   = Operation{'*', func(fst, snd int) (int, error) {
		if fst > 1_000 && snd > 1_000 || fst < -1_000 && snd < -1_000 || fst > 1_000_000 || fst < -1_000_000 || snd > 1_000_000 || snd < -1_000_000 {
			return 0, errors.New("result too big")
		}
		if snd < 0 {
			return 0, errors.New("duplicate negative")
		}
		return fst * snd, nil
	}}
	opDivide     = Operation{'/', func(fst, snd int) (int, error) {
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
	opRaise      = Operation{'^', func(fst, snd int) (int, error) {
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

func main() {
	start := time.Now()
	termEvalCounter := 0
	numbersGeneratedCounter := 0
	result := make(map[int][]string)
	builder := TermBuilder{DigitString}
	// maxCount := baseTerms[len(DigitString)]
	count := 0
	for term := range builder.Terms() {
		count++
		for value := range term.Values() {
			termEvalCounter++
			terms := result[value]
			if terms == nil {
				result[value] = make([]string, 0)
				numbersGeneratedCounter++
			}
			result[value] = append(terms, term.Compact())
		}
		// fmt.Printf("%d / %d\n", count, maxCount)
	}
	fmt.Printf("took %v\n", time.Since(start))
	fname := fmt.Sprintf("%s-complete-%d.txt", DigitString, time.Now().Unix())
	completeWriter := NewMyWriter(fname)
	compactWriter := NewMyWriter(strings.ReplaceAll(fname, "complete", "compact"))
	smallestNumber, gotSmallest := -1, false
	for key := 1; key <= 1000; key++ {
		terms := result[key]
		if terms == nil {
			if !gotSmallest {
				smallestNumber, gotSmallest = key, true
			}
			terms = []string{}
		}
		WriteResult(completeWriter, key, terms, -1)
		WriteResult(compactWriter, key, terms, 3)
	}
	printer := message.NewPrinter(language.English)
	printer.Printf("evaluated %d terms\n", termEvalCounter)
	printer.Printf("generated %d different numbers\n", numbersGeneratedCounter)
	printer.Printf("smallest number not generated: %d\n", smallestNumber)
}

func WriteResult(writer *FileWriter, key int, terms []string, limit int) {
	writer.Write(fmt.Sprintf("[%7d] %d", len(terms), key))
	selectedTerms := SelectRandom(terms, limit)
	for _, term := range selectedTerms {
		writer.Write(" = " + term)
	}
	if len(selectedTerms) < len(terms) {
		writer.Write(" = ...")
	}
	writer.Write("\n")
	writer.Flush()
}

func SelectRandom(slice []string, n int) []string {
	if n == -1 {
		return slice
	}
	indexSet := make(map[int]bool)
	for ;len(indexSet) < n && len(indexSet) < len(slice); {
		indexSet[rand.Intn(len(slice))] = true
	}
	var indexSlice []int
	for index, _ := range indexSet {
		indexSlice = append(indexSlice, index)
	}
	sort.Ints(indexSlice)
	var result []string
	for _, index := range indexSlice {
		result = append(result, slice[index])
	}
	return result
} 

type FileWriter struct {
	writer *bufio.Writer
}

func NewMyWriter(fileName string) *FileWriter {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	return &FileWriter{bufio.NewWriter(file)}
}

func (writer FileWriter) Write(str string) {
	_, err := writer.writer.WriteString(str)
	if err != nil {
		panic(err)
	}
}

func (writer FileWriter) Flush() {
	err := writer.writer.Flush()
	if err != nil {
		panic(err)
	}
}
