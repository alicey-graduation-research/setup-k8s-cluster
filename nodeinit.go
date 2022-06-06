package main

import(
	"fmt"
	//"net"
	"os/exec"
)

func main(){

	res, err := exec.Command("/bin/sh","./test.sh").Output()
    if err != nil {
        fmt.Print(err.Error())
    }
    fmt.Print(string(res))

	


}