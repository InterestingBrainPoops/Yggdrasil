package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
  "github.com/BattlesnakeOfficial/rules"
  "time"
)
var strl = rules.StandardRuleset{FoodSpawnChance:25, MinimumFood: 1}

type Game struct {
	ID      string `json:"id"`
	Timeout int32  `json:"timeout"`
}

type Coord struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
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
	Height int32           `json:"height"`
	Width  int32           `json:"width"`
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
	Turn  int32         `json:"turn"`
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
  ID string
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

func newMove(direction string, vector Coord, snake Battlesnake) Move {
  return Move {
    Head: snake.Head.add(vector),
    Direction: direction,
    ID: snake.ID,
  }
}

func (b Board) toBoardState() *rules.BoardState {
  food := make([]rules.Point, len(b.Food))
  for i, v := range b.Food {
    food[i] = rules.Point(v)
  }
  
  snakes := make([]rules.Snake, len(b.Snakes))
  for i, v := range b.Snakes {
    body := make([]rules.Point, len(v.Body))
    for i, v := range v.Body {
      body[i] = rules.Point(v)
    }

    snakes[i] = rules.Snake {
      ID: v.ID,
      Body: body,
      Health: v.Health,
    }
  }
  return &rules.BoardState{
    Height: b.Height,
    Width: b.Width,
    Food: food,
    Snakes: snakes,
  }
}

func (b Board) getMoves() []rules.SnakeMove {
  ret := make([]rules.SnakeMove, len(b.Snakes))
  for i, snake := range b.Snakes {
    possibleMoves := b.getValidMoves(snake) 
    var move Move
    if(len(possibleMoves ) == 0) {
      // dying
      move = Move{Direction:"up"};
    } else {
      move = possibleMoves[rand.Intn(len(possibleMoves))]
    }
    fmt.Printf("%s is moving %s\n", snake.Name, move.Direction)
    ret[i] = rules.SnakeMove{
      ID: snake.ID, 
      Move: move.Direction,
    }
  }
  return ret
}

func (b Board) getValidMoves(snake Battlesnake) []Move {
  // Choose a random direction to move in
	possibleMoves := []Move{
    Move {
     Direction: "up",
     Head: snake.Head.add(UP),
    }, 
    Move {
     Direction: "down",
     Head: snake.Head.add(DOWN),
    }, 
    Move {
     Direction: "left",
     Head: snake.Head.add(LEFT),
    }, 
    Move {
     Direction: "right",
     Head: snake.Head.add(RIGHT),
    },
  }
  // prune out the "invalid" moves. basically moves that will defenitely kill you

  // hitting another snake
  for _, v := range b.Snakes {
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
        return b.isInBounds(move.Head)
      })

  return possibleMoves
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

/*
type SnakeMove struct {
	ID   string
	Move string
}
*/



// HandleMove is called for each turn of each game.
// Valid responses are "up", "down", "left", or "right".
// TODO: Use the information in the GameRequest object to determine your next move.
func HandleMove(w http.ResponseWriter, r *http.Request) {
  start := time.Now()
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}
  
  standardRules := rules.StandardRuleset{FoodSpawnChance : 25, MinimumFood : 1};

  boardState, err := standardRules.CreateNextBoardState(request.Board.toBoardState(), request.Board.getMoves())
	
  fmt.Printf("New boardstate: %v\n", boardState)

  possibleMoves := request.Board.getValidMoves(request.You)
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
  endtime := time.Now()
  fmt.Printf("TimeTaken: %d Microseconds\n", endtime.Sub(start).Microseconds());
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
