package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sort"
)

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
