package main

import (
	"fmt"

	"github.com/tbocek/qotp"
)

func main() {
	fmt.Println("Connection ID Size:", qotp.ConnIdSize)
}
