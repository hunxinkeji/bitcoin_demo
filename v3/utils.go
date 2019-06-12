package main

import "fmt"

func IfError(data string, err error) {
	if err != nil {
		fmt.Println(data+"err=", err)
	}
}
