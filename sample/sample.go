package main

import(
	"fmt"
	"time"
	"strconv"
)
func main(){
fmt.Println(strconv.FormatInt(time.Now().Unix(),10))
}