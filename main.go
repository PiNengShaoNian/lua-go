package main

import (
	"io/ioutil"
	"lua_go/state"
	"os"
)

func main() {
	os.Args = append(os.Args, "a.out")
	if len(os.Args) > 1 {
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}

		ls := state.New()
		ls.Load(data, os.Args[1], "b")
		ls.Call(0, 0)
	}
}
