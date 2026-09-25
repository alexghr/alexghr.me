package main

import (
	"fmt"
	"os"

	"alexghr.me/internal/web"
	"golang.org/x/term"
)

func main() {
	fmt.Fprint(os.Stderr, "Password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read password: %q\n", err)
		os.Exit(1)
	}

	hash, err := web.HashPassword(pw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash password: %q\n", err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stderr, "\n"+hash+"\n")
}
