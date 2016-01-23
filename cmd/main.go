package main

import "fmt"
import "github.com/akramer/lateral/getsid"

func main() {
  sid, err := getsid.Getsid(0)
  if err != nil {
    panic(err)
  }
  fmt.Printf("%d\n", sid)
}
