package main

import "fmt"

func IfError(data string, err error) {
	if err != nil {
		fmt.Println(data+"err=", err)
	}
}

// 反转切片
func Reverse(data []byte) {
	for i, j := 0, len(data) - 1; i < j; i, j = i +1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
