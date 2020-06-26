package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	Move  string `json:"move"`
	Shout string `json:"shout,omitempty"`
}

type LinkedCoordinateList struct {
  X     int
  Y     int
  Next  *LinkedCoordinateList
  Depth int
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	response := BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "Frostflight",
		Color:      "#80d9ff",
		Head:       "evil",
		Tail:       "bolt",
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func HandleStart(w http.ResponseWriter, r *http.Request) {
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("START\n")
}

func HandleMove(w http.ResponseWriter, r *http.Request) {
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		log.Fatal(err)
	}

  move := chooseBehaviour(&request)

	response := MoveResponse{
		Move: move,
	}

	fmt.Printf("MOVE: %s\n", response.Move)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Fatal(err)
	}
}

func chooseBehaviour(data *GameRequest) string {
  //figure out which strategy to use.
  var board [11][11]int = constructBoard(data)
  return foodFocus(data,board)
}

func constructBoard(data *GameRequest) [11][11]int {
  var board [11][11]int

  //Mark food
  for _, food := range data.Board.Food {
    board[food.Y][food.X] = 1
  }

  //Mark snakes
  for _, snake := range data.Board.Snakes {
    if len(snake.Body) >= len(data.You.Body) && snake.Name != "Gosnake" {
      if (snake.Head.Y > 0) {
        board[snake.Head.Y-1][snake.Head.X] = 999
      }
      if (snake.Head.X > 0) {
        board[snake.Head.Y][snake.Head.X-1] = 999
      }
      if (snake.Head.Y < data.Board.Height-1) {
        board[snake.Head.Y+1][snake.Head.X] = 999
      }
      if (snake.Head.X < data.Board.Width-1) {
        board[snake.Head.Y][snake.Head.X+1] = 999
      }
    }
    for i, body := range snake.Body {
      board[body.Y][body.X] = len(snake.Body)+1-i
    }  
  }

  return board
}

func avoidFocus (data *GameRequest, board [11][11]int) string {
  var max int = countOpenSpaces(createBoardCopy(&board),data.You.Head.X,data.You.Head.Y+1,false)
  var maxDir string = "up"
  var temp int

  temp = countOpenSpaces(createBoardCopy(&board),data.You.Head.X,data.You.Head.Y-1,false)
  if (temp > max) {
    max = temp
    maxDir = "down"
  }

  temp = countOpenSpaces(createBoardCopy(&board),data.You.Head.X+1,data.You.Head.Y,false)
  if (temp > max) {
    max = temp
    maxDir = "right"
  }

  temp = countOpenSpaces(createBoardCopy(&board),data.You.Head.X-1,data.You.Head.Y,false)
  if (temp > max) {
    maxDir = "left"
  }

  if max == 0 {
    temp = countOpenSpaces(createBoardCopy(&board),data.You.Head.X,data.You.Head.Y-1,true)
    if (temp > max) {
      max = temp
      maxDir = "down"
    }

    temp = countOpenSpaces(createBoardCopy(&board),data.You.Head.X+1,data.You.Head.Y,true)
    if (temp > max) {
      max = temp
      maxDir = "right"
    }

    temp = countOpenSpaces(createBoardCopy(&board),data.You.Head.X-1,data.You.Head.Y,true)
    if (temp > max) {
      maxDir = "left"
    }
  }

  return maxDir
}

func createBoardCopy(board *[11][11]int) *[11][11]int {
  var copyBoard [11][11]int
  for y := 0; y < 11; y++ {
    for x := 0; x < 11; x++ {
      copyBoard[y][x] = board[y][x]
    }
  }
  return &copyBoard
}

/*func countOpenSpaces (board *[11][11]int, x int, y int) int {
  if (y > len(board)-1 || y < 0 || x > len(board[0])-1 || x < 0 || board[y][x] > 1) {
    return 0
  }
  board[y][x] = 1000
  var total int = 1
  if y < len(board)-1 && board[y+1][x] < 2 {
    total += countOpenSpaces(board, x, y+1)
  }
  if y > 0 && board[y-1][x] < 2 {
    total += countOpenSpaces(board, x, y-1)
  }
  if x < len(board[0])-1 && board[y][x+1] < 2 {
    total += countOpenSpaces(board, x+1, y)
  }
  if x > 0 && board[y][x-1] < 2 {
    total += countOpenSpaces(board, x-1, y)
  }
  return total
}*/

func countOpenSpaces (board *[11][11]int, startX int, startY int, flag bool) int {
  var head *LinkedCoordinateList = &LinkedCoordinateList{startX, startY, nil, 1}
  var tail *LinkedCoordinateList = head

  var maxDepth int = 0

  for head != nil {
    if (head.X < 0 || head.Y < 0 || head.X >= len(board[0]) || head.Y >= len(board) || board[head.Y][head.X] > head.Depth+1) {
      if (!(head.X < 0 && head.Y < 0 && head.X >= len(board[0]) && head.Y >= len(board) && flag && board[head.Y][head.X] == 999)) {
        head = head.Next
        continue;
      }
    }
    board[head.Y][head.X] = 1000
    if head.Depth > maxDepth {
      maxDepth = head.Depth
    }

    tail.Next = &LinkedCoordinateList{head.X,head.Y+1,nil,head.Depth+1}
    tail = tail.Next
    
    tail.Next = &LinkedCoordinateList{head.X,head.Y-1,nil,head.Depth+1}
    tail = tail.Next
    
    tail.Next = &LinkedCoordinateList{head.X+1,head.Y,nil,head.Depth+1}
    tail = tail.Next
    
    tail.Next = &LinkedCoordinateList{head.X-1,head.Y,nil,head.Depth+1}
    tail = tail.Next

    head = head.Next
  }

  return maxDepth
}

func foodFocus (data *GameRequest, board [11][11]int) string {
  fmt.Println("Food Focus")
  fmt.Println("Snake Length:",len(data.You.Body))
  if (len(data.Board.Food) != 0) {

    var min int = 1000
    var minDir string = ""
    var temp int
    var tempBoardSpace int

    temp = foodBFS(board,data.You.Head.X,data.You.Head.Y+1)
    tempBoardSpace = countOpenSpaces(createBoardCopy(&board),data.You.Head.X,data.You.Head.Y+1,false)
    fmt.Println("Up: ",tempBoardSpace)
    if (temp < min && tempBoardSpace > len(data.You.Body)) {
      min = temp
      minDir = "up"
    }

    temp = foodBFS(board,data.You.Head.X,data.You.Head.Y-1)
    fmt.Println("Up: ",tempBoardSpace)
    tempBoardSpace = countOpenSpaces(createBoardCopy(&board),data.You.Head.X,data.You.Head.Y-1,false)
    if (temp < min && tempBoardSpace > len(data.You.Body)) {
      min = temp
      minDir = "down"
    }

    temp = foodBFS(board,data.You.Head.X+1,data.You.Head.Y)
    fmt.Println("Up: ",tempBoardSpace)
    tempBoardSpace = countOpenSpaces(createBoardCopy(&board),data.You.Head.X+1,data.You.Head.Y,false)
    if (temp < min  && tempBoardSpace > len(data.You.Body)) {
      min = temp
      minDir = "right"
    }

    temp = foodBFS(board,data.You.Head.X-1,data.You.Head.Y)
    fmt.Println("Up: ",tempBoardSpace)
    tempBoardSpace = countOpenSpaces(createBoardCopy(&board),data.You.Head.X-1,data.You.Head.Y,false)
    if (temp < min  && tempBoardSpace > len(data.You.Body)) {
      minDir = "left"
    }
    if (min < 1000) {
      return minDir
    } else {
      return avoidFocus(data, board)
    }
  }
  fmt.Println("No visible food, Avoiding")
  return avoidFocus(data, board)
}

func foodBFS(board [11][11]int, startX int, startY int) int {
  var head *LinkedCoordinateList = &LinkedCoordinateList{startX, startY, nil, 1}
  var tail *LinkedCoordinateList = head

  for head != nil {
    if (head.X < 0 || head.Y < 0 || head.X >= len(board[0]) || head.Y >= len(board) || board[head.Y][head.X] > head.Depth+1) {
      head = head.Next
      continue;
    }
    if (board[head.Y][head.X] == 1) {
      return head.Depth
    }

    board[head.Y][head.X] = 1000

    tail.Next = &LinkedCoordinateList{head.X,head.Y+1,nil,head.Depth+1}
    tail = tail.Next
    
    tail.Next = &LinkedCoordinateList{head.X,head.Y-1,nil,head.Depth+1}
    tail = tail.Next
    
    tail.Next = &LinkedCoordinateList{head.X+1,head.Y,nil,head.Depth+1}
    tail = tail.Next
    
    tail.Next = &LinkedCoordinateList{head.X-1,head.Y,nil,head.Depth+1}
    tail = tail.Next

    head = head.Next
  }

  return 10000
}

func HandleEnd(w http.ResponseWriter, r *http.Request) {
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}
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
