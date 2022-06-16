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

	// masterにkubeadmにjoinするトークン要求
	conn, err := net.Dial("udp4", "255.255.255.255:43210")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("please-kubeadm-token"))
	if err != nil {
		panic(err)
	}

	// token受け取り
	
	// kubeadm joinする　
	_, err := exec.Command("/usr/bin/","./setup_test.sh").Output()
    if err != nil {
        fmt.Print(err.Error())
    }

}