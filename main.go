package main

import (
    "fmt"
)

func main() {
	p, q := cal(123, 456)
  print(p+q)
}

func cal(x, y int) (sum int, z int) {
  m := x + y
  n := m*2
  print(m + n)
  return m, n
}

func print(p int) {
    fmt.Println(p)
}
