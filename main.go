package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

var (
	board  = [3][3]string{}
	turn   = "X"
	winner = ""
	lock   sync.Mutex
)

type BoardResponse struct {
	Board  [3][3]string `json:"board"`
	Winner string       `json:"winner"`
}

func main() {
	http.HandleFunc("/api/board", getBoard)
	http.HandleFunc("/api/play", playMove)
	http.HandleFunc("/api/reset", resetGame)
	http.Handle("/", http.FileServer(http.Dir("."))) // Serve arquivos estáticos
	http.ListenAndServe(":8080", nil)
}

// Retorna o estado atual do tabuleiro e o vencedor
func getBoard(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	response := BoardResponse{
		Board:  board,
		Winner: winner,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Processa uma jogada e retorna o tabuleiro atualizado
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
	if board[row][col] == "" && winner == "" {
		board[row][col] = turn
		if checkWinner() != "" {
			winner = turn
		} else if checkDraw() {
			winner = "Draw"
		} else {
			turn = map[string]string{"X": "O", "O": "X"}[turn]
		}
	}
	lock.Unlock()

	getBoard(w, r) // Retorna o tabuleiro e vencedor atualizados
}

// Reseta o tabuleiro do jogo
func resetGame(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	board = [3][3]string{}
	turn = "X"
	winner = ""
	lock.Unlock()

	getBoard(w, r) // Retorna o tabuleiro vazio e sem vencedor
}

// Verifica se há um vencedor ou empate
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
				return "" // Jogo ainda em andamento
			}
		}
	}
	return "Draw" // Empate
}

func checkDraw() bool {
	for _, row := range board {
		for _, cell := range row {
			if cell == "" {
				return false
			}
		}
	}
	return true
}
