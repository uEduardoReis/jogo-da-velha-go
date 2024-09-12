package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

var (
	board = [3][3]string{}
	turn  = "X"
	lock  sync.Mutex
)

func main() {
	http.HandleFunc("/api/board", getBoard)
	http.HandleFunc("/api/play", playMove)
	http.HandleFunc("/api/reset", resetGame)
	http.Handle("/", http.FileServer(http.Dir("."))) // Serve static files from the current directory
	http.ListenAndServe(":8080", nil)
}

// Struct for the board response
type BoardResponse struct {
	Board  [3][3]string `json:"board"`
	Winner string       `json:"winner"`
}

// Returns the current state of the board and the winner
func getBoard(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	response := BoardResponse{
		Board:  board,
		Winner: checkWinner(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Processes a move and returns the updated board
func playMove(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	rowStr := r.FormValue("row")
	colStr := r.FormValue("col")
	row, err1 := strconv.Atoi(rowStr)
	col, err2 := strconv.Atoi(colStr)
	if err1 != nil || err2 != nil || row < 0 || row > 2 || col < 0 || col > 2 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	lock.Lock()
	if board[row][col] == "" && checkWinner() == "" {
		board[row][col] = turn
		turn = map[string]string{"X": "O", "O": "X"}[turn]
	}
	lock.Unlock()

	getBoard(w, r) // Return the updated board and winner
}

// Resets the game board
func resetGame(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	board = [3][3]string{}
	turn = "X"
	lock.Unlock()

	getBoard(w, r) // Return the initial empty board and no winner
}

// Check if there is a winner or if it's a draw
func checkWinner() string {
	lines := [][3][2]int{
		{{0, 0}, {0, 1}, {0, 2}}, // Rows
		{{1, 0}, {1, 1}, {1, 2}},
		{{2, 0}, {2, 1}, {2, 2}},
		{{0, 0}, {1, 0}, {2, 0}}, // Columns
		{{0, 1}, {1, 1}, {2, 1}},
		{{0, 2}, {1, 2}, {2, 2}},
		{{0, 0}, {1, 1}, {2, 2}}, // Diagonals
		{{0, 2}, {1, 1}, {2, 0}},
	}
	for _, line := range lines {
		if board[line[0][0]][line[0][1]] != "" &&
			board[line[0][0]][line[0][1]] == board[line[1][0]][line[1][1]] &&
			board[line[1][0]][line[1][1]] == board[line[2][0]][line[2][1]] {
			return board[line[0][0]][line[0][1]]
		}
	}
	for _, row := range board {
		for _, cell := range row {
			if cell == "" {
				return "" // The game is still ongoing
			}
		}
	}
	return "Draw" // All cells are filled and no winner
}
