package main

import (
	"log"
	"os"
)

func main() {
	if _, err := os.Stat(pkgName); os.IsNotExist(err) {
		os.Mkdir(pkgName, os.ModePerm)
	}
	err := NewGenerator().Generate()
	if err != nil {
		log.Fatal(err)
	}
}
