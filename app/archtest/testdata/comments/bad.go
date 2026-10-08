// Package bad explains itself in a comment.
package bad

import "fmt"

//go:generate echo allowed

func Run() {
	fmt.Println("http://not-a-comment") // trailing comment
	/* block comment */
	_ = 1 //nolint:gosec
	// slices:imports
}
