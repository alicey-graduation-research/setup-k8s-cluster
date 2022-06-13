package main

import(
	"fmt"
	"net"
	"os/exec"
	"os"
)

func main(){

	os.Setenv("TEST","testenv")
	// fmt.Print(string(os.Getenv("TEST")))

	// nodeにK8sコンポーネントのインストール
	_, err := exec.Command("/bin/sh","./setup_component.sh").Output()
    if err != nil {
        fmt.Print(err.Error())
    }

	// masterのIPアドレスを調べる


	// masterにkubeadmにjoinするトークン要求
	
	// kubeadm joinする　
	_, err := exec.Command("/usr/bin/","./setup_test.sh").Output()
    if err != nil {
        fmt.Print(err.Error())
    }

}