package main

import "fmt"

func main() {
	for i := 0; i < 1000000000000000000; i++ {
		for j := 0; j < 1000000000000000000; j++ {
			fmt.Printf("it is - i = %s and j = %s\n", i, j)
		}
	}
}
