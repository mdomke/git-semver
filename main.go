package main

import (
	"fmt"
	"log"
)

func main() {
	v, err := GitVersion()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(v)
}
