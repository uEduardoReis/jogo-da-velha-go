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
	http.Handle("/", http.FileServer(http.Dir("."))) // Serve arquivos estáticos do diretório atual
	http.ListenAndServe(":8080", nil)
}

type BoardResponse struct {
	Board  [3][3]string `json:"board"`
	Winner string       `json:"winner"`
}

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

func playMove(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	rowStr := r.FormValue("row")
	colStr := r.FormValue("col")
	row, err1 := strconv.Atoi(rowStr)
	col, err2 := strconv.Atoi(colStr)
	if err1 != nil || err2 != nil || row < 0 || row > 2 || col < 0 || col > 2 {
		http.Error(w, "Requisição inválida", http.StatusBadRequest)
		return
	}

	lock.Lock()
	if board[row][col] == "" && checkWinner() == "" {
		board[row][col] = turn
		turn = map[string]string{"X": "O", "O": "X"}[turn]
	}
	lock.Unlock()

	getBoard(w, r) // Retorna o tabuleiro atualizado e o vencedor
}

func resetGame(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	board = [3][3]string{}
	turn = "X"
	lock.Unlock()

	getBoard(w, r) // Retorna o tabuleiro vazio e sem vencedor
}

func checkWinner() string {
	lines := [][3][2]int{
		{{0, 0}, {0, 1}, {0, 2}}, // Linhas
		{{1, 0}, {1, 1}, {1, 2}},
		{{2, 0}, {2, 1}, {2, 2}},
		{{0, 0}, {1, 0}, {2, 0}}, // Colunas
		{{0, 1}, {1, 1}, {2, 1}},
		{{0, 2}, {1, 2}, {2, 2}},
		{{0, 0}, {1, 1}, {2, 2}}, // Diagonais
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
				return "" // O jogo ainda está em andamento
			}
		}
	}
	return "Draw" // Todas as células estão preenchidas e não há vencedor
}
