package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/BattlesnakeOfficial/rules"
)

var strl = rules.StandardRuleset{FoodSpawnChance: 25, MinimumFood: 1}

type Game struct {
	ID      string `json:"id"`
	Timeout int32  `json:"timeout"`
}

var UP = rules.Point{X: 0, Y: 1}
var DOWN = rules.Point{X: 0, Y: -1}
var LEFT = rules.Point{X: -1, Y: 0}
var RIGHT = rules.Point{X: 1, Y: 0}

type BattlesnakeInfoResponse struct {
	APIVersion string `json:"apiversion"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
}

type GameRequest struct {
	Game  Game             `json:"game"`
	Turn  int32            `json:"turn"`
	Board rules.BoardState `json:"board"`
	You   rules.Snake      `json:"you"`
}

type MoveResponse struct {
	Direction string `json:"move"`
	Shout     string `json:"shout,omitempty"`
}

type Move struct {
	Head      rules.Point
	Direction string
	ID        string
}

// HandleIndex is called when your Battlesnake is created and refreshed
// by play.battlesnake.com. BattlesnakeInfoResponse contains information about
// your Battlesnake, including what it should look like on the game board.
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	response := BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "Odin",
		Color:      "#3ebf37",
		Head:       "default",
		Tail:       "freckled",
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

// HandleStart is called at the start of each game your Battlesnake is playing.
// The GameRequest object contains information about the game that's about to start.
// TODO: Use this function to decide how your Battlesnake is going to look on the board.
func HandleStart(w http.ResponseWriter, r *http.Request) {
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	// Nothing to respond with here
	fmt.Print("START\n")
}

func add(c rules.Point, vector rules.Point) rules.Point {
	return rules.Point{
		X: c.X + vector.X,
		Y: c.Y + vector.Y,
	}
}

func isInBounds(b rules.BoardState, pos rules.Point) bool {
	return pos.X >= 0 && pos.X < b.Width && pos.Y >= 0 && pos.Y < b.Height
}

func newMove(direction string, vector rules.Point, snake rules.Snake) Move {
	return Move{
		Head:      add(snake.Body[0], vector),
		Direction: direction,
		ID:        snake.ID,
	}
}

func getMoves(b rules.BoardState) []rules.SnakeMove {
	ret := make([]rules.SnakeMove, len(b.Snakes))
	for i, snake := range b.Snakes {
		possibleMoves := getValidMoves(b, snake)
		var move Move
		if len(possibleMoves) == 0 {
			// dying
			move = Move{Direction: "up"}
		} else {
			move = possibleMoves[rand.Intn(len(possibleMoves))]
		}

		fmt.Printf("%s is moving %s\n", snake.ID, move.Direction)
		ret[i] = rules.SnakeMove{
			ID:   snake.ID,
			Move: move.Direction,
		}
	}
	return ret
}

func getValidMoves(b rules.BoardState, snake rules.Snake) []Move {
	// Choose a random direction to move in
	possibleMoves := []Move{
		Move{
			Direction: "up",
			Head:      add(snake.Body[0], UP),
		},
		Move{
			Direction: "down",
			Head:      add(snake.Body[0], DOWN),
		},
		Move{
			Direction: "left",
			Head:      add(snake.Body[0], LEFT),
		},
		Move{
			Direction: "right",
			Head:      add(snake.Body[0], RIGHT),
		},
	}
	// prune out the "invalid" moves. basically moves that will defenitely kill you

	// hitting another snake
	for _, v := range b.Snakes {
		bodyMinusTail := v.Body[:len(v.Body)-1]
		for _, segment := range bodyMinusTail {
			possibleMoves = Filter(possibleMoves,
				func(move Move) bool {
					return move.Head != segment
				})
		}
	}

	// out of bounds
	possibleMoves = Filter(possibleMoves,
		func(move Move) bool {
			return isInBounds(b, move.Head)
		})

	return possibleMoves
}

func Filter(vs []Move, f func(Move) bool) []Move {
	vsf := make([]Move, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// HandleMove is called for each turn of each game.
// Valid responses are "up", "down", "left", or "right".
// TODO: Use the information in the GameRequest object to determine your next move.
func HandleMove(w http.ResponseWriter, r *http.Request) {
	start := time.Now()      // create times
	request := GameRequest{} // jsonifies the request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	standardRules := rules.StandardRuleset{FoodSpawnChance: 25, MinimumFood: 1} // defines standard ruleset for game playouts

	moves := getMoves(request.Board)

	boardState, err := standardRules.CreateNextBoardState(&request.Board, moves) // creates boardstate based on random playout

	fmt.Printf("New boardstate: %v\n", boardState)

	possibleMoves := getValidMoves(request.Board, request.You) // generates valid moves for snake
	var move Move
	if len(possibleMoves) == 0 {
		fmt.Println("Dying now :(") // if there is no more valid moves , itll just die.
		move = Move{Direction: "up"}
	} else {
		move = possibleMoves[rand.Intn(len(possibleMoves))] // otherwise it will pick at random what move to go to.
	}
	response := MoveResponse{
		Direction: move.Direction,
	}
	fmt.Printf("TURN: %s\n", request.Turn)
	fmt.Printf("MOVE: %s\n", response.Direction)
	//time.Sleep(100 * time.Millisecond)
	endtime := time.Now()
	fmt.Printf("TimeTaken: %d Microseconds\n", endtime.Sub(start).Microseconds())
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response) // sends the thing off.
	if err != nil {
		log.Fatal(err)
	}

}

// HandleEnd is called when a game your Battlesnake was playing has ended.
// It's purely for informational purposes, no response required.
func HandleEnd(w http.ResponseWriter, r *http.Request) {
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	// Nothing to respond with here
	fmt.Print("END\n")
}

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8000"
	}

	http.HandleFunc("/ygg/", HandleIndex)
	http.HandleFunc("/ygg/start", HandleStart)
	http.HandleFunc("/ygg/move", HandleMove)
	http.HandleFunc("/ygg/end", HandleEnd)

	fmt.Printf("Starting Battlesnake Server at http://0.0.0.0:%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
