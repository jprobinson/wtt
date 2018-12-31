package main

import (
	"github.com/NYTimes/gizmo/server/kit"
	"github.com/jprobinson/wtt"
)

func main() {
	err := kit.Run(wtt.NewService())
	if err != nil {
		panic(err)
	}
}
