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
var operations = []Operation{opAdd, opMultiply, opDivide, opRaise}
const UseNegation = true  // don't use opSubtract in this case!

type ValueGenerator interface {
	HasValue() bool
	NextValue()
	ResetValue()
}

type Cache struct {
	value int
	err error
	valid bool
}

type Term struct {  // todo delegate most to operation
	fst, snd  SubTerm
	negationFactor *Integer  // only used for raise
	operationIndex int
	cached Cache
}

func NewTerm(fst, snd SubTerm) *Term {
	return &Term{fst, snd, NewInteger(1), 0, Cache{0, nil, false}}
}

func (term *Term) Operation() Operation {
	return operations[term.operationIndex]
}

func (term *Term) Value() (int, error) {
	if term.cached.valid {
		return term.cached.value, term.cached.err
	}
	term.cached.valid = true
	fstValue, err := term.fst.Value()
	if err != nil {term.cached.err = err; return 0, err}
	sndValue, err := term.snd.Value()
	if err != nil {term.cached.err = err; return 0, err}
	result, err := term.Operation().eval(fstValue, sndValue)
	if err != nil {term.cached.err = err; return 0, err}
	factor, _ := term.negationFactor.Value()
	term.cached.value = factor * result
	term.cached.err = nil
	return term.cached.value, nil
}

func (term *Term) HasValue() bool {
	return term.operationIndex < len(operations)
}

func (term *Term) NextValue() {  // should delegate to operation
	term.cached.valid = false
	subTermGen := []ValueGenerator{term.fst, term.snd}
	if term.Operation().symbol == '^' {
		subTermGen = append(subTermGen, term.negationFactor)
	}
	for _, subTerm := range subTermGen {
		subTerm.NextValue()
		if subTerm.HasValue() {
			return
		}
		subTerm.ResetValue()
	}
	term.operationIndex++
}

func (term *Term) ResetValue() {
	term.operationIndex = 0
}

func (term *Term) String() string {
	prefix := ""
	if term.negationFactor.negated {
		prefix = " -"
	}
	return fmt.Sprintf("%s( %s %c %s )", prefix, term.fst.String(), term.Operation().symbol, term.snd.String())
}

func (term *Term) Compact() string {
	return strings.ReplaceAll(term.String(), " ", "")
}

type SubTerm interface {
	ValueGenerator
	Value() (int, error)
	String() string
	Compact() string
}

type Integer struct {
	value int
	negated bool
	exhausted bool
}

func NewInteger(value int) *Integer {
	return &Integer{value, false, false}
}

func (int *Integer) Value() (int, error) {
	if int.negated {
		return -int.value, nil
	} else {
		return int.value, nil
	}
}

func (int *Integer) HasValue() bool {
	return !int.exhausted
}

func (int *Integer) NextValue() {
	if !int.negated && UseNegation {
		int.negated = true
	} else {
		int.exhausted = true
	}
}

func (int *Integer) ResetValue() {
	int.negated = false
	int.exhausted = false
}

func (int *Integer) String() string {
	value, _ := int.Value()
	return strconv.Itoa(value)
}

func (int *Integer) Compact() string {
	return int.String()
}

type TermBuilder struct {
	digits string
	splitIndex int
	fst, snd* TermBuilder
}

func NewTermBuilder(digits string) *TermBuilder {
	return &TermBuilder{digits, 0, nil, nil}
}

func Atoi(digits string) int {
	integer, _ := strconv.Atoi(digits)  // todo error handling
	return integer	
}

func (builder *TermBuilder) Value() SubTerm {
	if builder.splitIndex == 0 {
		return NewInteger(Atoi(builder.digits))
	}
	return NewTerm(builder.fst.Value(), builder.snd.Value())
}

func (builder *TermBuilder) HasValue() bool {
	return builder.splitIndex < len(builder.digits)
}

func (builder *TermBuilder) NextValue() {
	if builder.splitIndex == 0 {
		builder.Next()
		return
	}
	for _, subTerm := range []ValueGenerator{builder.fst, builder.snd} {
		subTerm.NextValue()
		if subTerm.HasValue() {
			return
		}
		subTerm.ResetValue()
	}
	builder.Next()
}

func (builder *TermBuilder) Next() {
	builder.splitIndex += 1
	inverseSplitIndex := len(builder.digits) - builder.splitIndex
	builder.fst = NewTermBuilder(builder.digits[:inverseSplitIndex])
	builder.snd = NewTermBuilder(builder.digits[inverseSplitIndex:])
}

func (builder *TermBuilder) ResetValue() {
	builder.splitIndex = 0
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
	termGenCounter := 0
	termEvalCounter := 0
	numbersGeneratedCounter := 0
	result := make(map[int][]string)
	builder := *NewTermBuilder(DigitString)
	maxCount := baseTerms[len(DigitString)]
	count := 0
	for ;builder.HasValue();builder.NextValue() {
		count++
		term := builder.Value()
		for ;term.HasValue();term.NextValue() {
			value, err := term.Value()
			termGenCounter++
			if err != nil {
				continue
			}
			termEvalCounter++
			terms := result[value]
			if terms == nil {
				result[value] = make([]string, 0)
				numbersGeneratedCounter++
			}
			result[value] = append(terms, term.Compact())
		}
		fmt.Printf("%d / %d\n", count, maxCount)
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
	printer.Printf("generated %d terms\n", termGenCounter)
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
