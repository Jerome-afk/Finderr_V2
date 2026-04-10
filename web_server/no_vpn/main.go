package main

import (
	utility "finderr_v2/utils"
	"fmt"
)

// TODO: Main server implementation
func main() {
	// 1. Initialize config
	cfg, err := utility.InitConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}
}
