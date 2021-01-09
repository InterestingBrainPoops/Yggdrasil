package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
)

type Game struct {
	ID      string `json:"id"`
	Timeout int32  `json:"timeout"`
}

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var UP = Coord{ X: 0, Y: 1}
var DOWN = Coord{ X: 0, Y: -1}
var LEFT = Coord{ X: -1, Y: 0}
var RIGHT = Coord{ X: 1, Y: 0}

type Battlesnake struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Health int32   `json:"health"`
	Body   []Coord `json:"body"`
	Head   Coord   `json:"head"`
	Length int32   `json:"length"`
	Shout  string  `json:"shout"`
}

type Board struct {
	Height int           `json:"height"`
	Width  int           `json:"width"`
	Food   []Coord       `json:"food"`
	Snakes []Battlesnake `json:"snakes"`
}

type BattlesnakeInfoResponse struct {
	APIVersion string `json:"apiversion"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
}

type GameRequest struct {
	Game  Game        `json:"game"`
	Turn  int         `json:"turn"`
	Board Board       `json:"board"`
	You   Battlesnake `json:"you"`
}

type MoveResponse struct {
	Direction  string `json:"move"`
	Shout string `json:"shout,omitempty"`
}

type Move struct {
  Head Coord
  Direction string
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

func (c Coord) add(vector Coord) Coord {
  return Coord {
    X: c.X + vector.X,
    Y: c.Y + vector.Y,
  }
}

func (b Board) isInBounds(pos Coord) bool {
  return pos.X >= 0 && pos.X < b.Width && pos.Y >= 0 && pos.Y < b.Height
}

// https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
func remove(s []Move, i int) []Move {
  s[i] = s[len(s)-1]
  // We do not need to put s[i] at the end, as it will be discarded anyway
  return s[:len(s)-1]

  /*if i+1 < len(array) {
    return append(array[:i], array[i+1:]...)
  }else if i == 0{
  return array[:i]
  }
  return array[:i-1]*/
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
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}
  
	// Choose a random direction to move in
	possibleMoves := []Move{
    Move {
     Direction: "up",
     Head: request.You.Head.add(UP),
    }, 
    Move {
     Direction: "down",
     Head: request.You.Head.add(DOWN),
    }, 
    Move {
     Direction: "left",
     Head: request.You.Head.add(LEFT),
    }, 
    Move {
     Direction: "right",
     Head: request.You.Head.add(RIGHT),
    },
  }
  // prune out the "invalid" moves. basically moves that will defenitely kill you.

  // hitting another snake
  for _, v := range request.Board.Snakes { 
    bodyMinusTail := v.Body[:len(v.Body)-1]
    for _, segment := range bodyMinusTail {
      possibleMoves = Filter(possibleMoves, 
      func (move Move) bool {
        return move.Head != segment 
      })
    }
  }

  // out of bounds
  possibleMoves = Filter(possibleMoves, 
      func (move Move) bool {
        return request.Board.isInBounds(move.Head)
      })
  var move Move
  if(len(possibleMoves ) == 0) {
    fmt.Println("Dying now :(")
    move = Move{Direction:"up"};
  }else{
	  move = possibleMoves[rand.Intn(len(possibleMoves))]
  }
	response := MoveResponse{
		Direction: move.Direction,
	}

	fmt.Printf("MOVE: %s\n", response.Direction)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
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
		port = "8080"
	}

	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/start", HandleStart)
	http.HandleFunc("/move", HandleMove)
	http.HandleFunc("/end", HandleEnd)

	fmt.Printf("Starting Battlesnake Server at http://0.0.0.0:%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
