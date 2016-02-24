package main

import (
	"fmt"
	"gopkg.in/src-d/go-git.v2"
	"io"
)

func main() {
	r, err := git.NewRepository("https://github.com/src-d/go-git", nil)
	if err != nil {
		panic(err)
	}

	if err := r.Pull("origin", "refs/heads/master"); err != nil {
		panic(err)
	}

	iter := r.Commits()
	defer iter.Close()

	for {
		commit, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			panic(err)
		}

		fmt.Println(commit)
	}
}
