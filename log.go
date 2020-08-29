package main

import (
	"fmt"
	"github.com/TwinProduction/go-color"
)

func logInfo(reference, data string) {
	fmt.Println(color.Ize(color.Green, "["+reference+"] --INF-- "+data))
}

func logError(reference, data string) {
	fmt.Println(color.Ize(color.Red, "["+reference+"] --INF-- "+data))
}
