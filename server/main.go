package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

type Board [3][3]string

type Message struct {
	Type   string `json:"type"`
	Row    int    `json:"row,omitempty"`
	Col    int    `json:"col,omitempty"`
	Board  Board  `json:"board,omitempty"`
	Symbol string `json:"symbol,omitempty"`
	Result string `json:"result,omitempty"`
}

type Client struct {
	conn   net.Conn
	enc    *json.Encoder
	dec    *json.Decoder
	symbol string
}

func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server listening on :9000")

	waiting := make(chan *Client)
	go matchmaker(waiting)


	for {
		conn, err := listener.Accept()

		if err != nil {
			continue
		}

		client := &Client{
			conn: conn,
			enc:  json.NewEncoder(conn),
			dec:  json.NewDecoder(bufio.NewReader(conn)),
		}

		waiting <- client
	}
}

func matchmaker(waiting chan *Client) {

	// fmt.Println("test:",&waiting)

	for {
		p1 := <-waiting
		p2 := <-waiting
		go runGame(p1, p2)
	}
}

func runGame(p1, p2 *Client) {
	defer p1.conn.Close()
	defer p2.conn.Close()

	p1.symbol = "X"
	p2.symbol = "O"

	p1.enc.Encode(Message{Type: "start", Symbol: "X"})
	p2.enc.Encode(Message{Type: "start", Symbol: "O"})

	var board Board
	current := p1

	for {
		current.enc.Encode(Message{Type: "turn", Board: board})

		var move Message
		if err := current.dec.Decode(&move); err != nil {
			return
		}

		if !makeMove(&board, move.Row, move.Col, current.symbol) {
			continue
		}

		p1.enc.Encode(Message{Type: "state", Board: board})
		p2.enc.Encode(Message{Type: "state", Board: board})

		if hasWon(board, current.symbol) {
			current.enc.Encode(Message{Type: "end", Result: "win"})
			other(current, p1, p2).enc.Encode(Message{Type: "end", Result: "lose"})
			return
		}

		if isDraw(board) {
			p1.enc.Encode(Message{Type: "end", Result: "draw"})
			p2.enc.Encode(Message{Type: "end", Result: "draw"})
			return
		}

		current = other(current, p1, p2)
	}
}

func other(c, p1, p2 *Client) *Client {
	if c == p1 {
		return p2
	}
	return p1
}

func makeMove(b *Board, row, col int, s string) bool {
	if row < 0 || row > 2 || col < 0 || col > 2 {
		return false
	}
	if b[row][col] != "" {
		return false
	}
	b[row][col] = s
	return true
}

func hasWon(b Board, s string) bool {
	lines := [][][2]int{
		{{0, 0}, {0, 1}, {0, 2}},
		{{1, 0}, {1, 1}, {1, 2}},
		{{2, 0}, {2, 1}, {2, 2}},
		{{0, 0}, {1, 0}, {2, 0}},
		{{0, 1}, {1, 1}, {2, 1}},
		{{0, 2}, {1, 2}, {2, 2}},
		{{0, 0}, {1, 1}, {2, 2}},
		{{0, 2}, {1, 1}, {2, 0}},
	}

	for _, line := range lines {
		if b[line[0][0]][line[0][1]] == s &&
			b[line[1][0]][line[1][1]] == s &&
			b[line[2][0]][line[2][1]] == s {
			return true
		}
	}
	return false
}

func isDraw(b Board) bool {
	for _, row := range b {
		for _, cell := range row {
			if cell == "" {
				return false
			}
		}
	}
	return true
}
