type Node struct {
  MoveList []rules.SnakeMove
  Children []Node
  Visits int32
  RolloutScore int32
}

func getUCB(Parent Node, Nodeid int32) float64 {
  if(Parent[Nodeid].Visits == 0){
    return math.Inf(1) // n/0 == infinity in this usecase. in reality this is not the case.
  }
  avg := 0
  if(len(Parent[Nodeid].Children) == 0){
    avg = 0
  }else{
    sum := 0
    for _ , child range Parent[Nodeid].Children {
      sum += child.
    }
  }
}