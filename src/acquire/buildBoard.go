package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type BoardPiece struct {
	Text     string `json:"text"`
	Occupied bool   `json:"occupied"`
}

func main() {

	var boardPiece BoardPiece
	var BoardPieces []BoardPiece

	for row := 1; row < 10; row++ {
		for col := 1; col < 13; col++ {
			boardPiece.Text = string(row+64) + strconv.Itoa(col)
			boardPiece.Occupied = false
			BoardPieces = append(BoardPieces, boardPiece)
		}
	}

	jsondata, err := json.Marshal(BoardPieces) // convert to JSON

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// sanity check
	// NOTE : You can stream the JSON data to http service as well instead of saving to file
	fmt.Println(string(jsondata))

	// now write to JSON file

	jsonFile, err := os.Create("./board.json")

	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	jsonFile.Write(jsondata)
	jsonFile.Close()
}
