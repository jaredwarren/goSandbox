package main

import (
	"fmt"
	"litmosauthor.com/unison/hello/fuzzy"
)

func main() {
	model := fuzzy.NewModel()
	// For testing only, this is not advisable on production
	model.SetThreshold(1)

	// This expands the distance searched, but costs more resources (memory and time).
	// For spell checking, "2" is typically enough, for query suggestions this can be higher
	model.SetDepth(5)

	words := []string{"bob", "your", "uncle", "dynamite", "delicate", "biggest", "big", "bigger", "aunty", "you're"}
	model.Train(words)
	fmt.Println("   Deletion test (yor) : ", model.SpellCheck("yor"))
}
