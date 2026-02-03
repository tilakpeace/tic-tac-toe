package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Board [3][3]string

type Message struct {
	Type   string
	Row    int
	Col    int
	Board  Board
	Symbol string
	Result string
}

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)
	input := bufio.NewReader(os.Stdin)

	for {
		var msg Message
		if err := dec.Decode(&msg); err != nil {
			return
		}

		switch msg.Type {

		case "start":
			fmt.Println("Game started. You are:", msg.Symbol)

		case "turn":
			printBoard(msg.Board)
			fmt.Print("Your move (row col): ")
			fmt.Fscan(input, &msg.Row, &msg.Col)
			enc.Encode(Message{Type: "move", Row: msg.Row, Col: msg.Col})

		case "state":
			printBoard(msg.Board)

		case "end":
			fmt.Println("Game ended:", msg.Result)
			return
		}
	}
}

func printBoard(b Board) {
	fmt.Println()
	for _, row := range b {
		for _, cell := range row {
			if cell == "" {
				fmt.Print(". ")
			} else {
				fmt.Print(cell, " ")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}
