package main

import (
	"fmt"
	"strings"
)

func main() {
	original := "root/users/myserver"
	result := strings.TrimPrefix(original, "root")
	fmt.Printf("Original: %q\n", original)
	fmt.Printf("After TrimPrefix('root'): %q\n", result)
	
	if !strings.HasPrefix(result, "/") {
		result = "/" + result
	}
	fmt.Printf("Final result: %q\n", result)
}
