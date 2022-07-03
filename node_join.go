package main

import(
	"fmt"
	"net"
	"os/exec"
	"os"
	"io"
	"log"
)

func main(){

	os.Setenv("TEST","testenv")
	// fmt.Print(string(os.Getenv("TEST")))

	// nodeにK8sコンポーネントのインストール
	_, err := exec.Command("/bin/sh","./setup_component.sh").Output()
    if err != nil {
        fmt.Println("K8s compornent install error:" + err.Error())
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
	log.SetOutput(io.MultiWriter(os.Stdout))
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 32432,
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalln(err)
	}

	buf := make([]byte, 64)
	log.Println("Starting UDP Server...")

	for {
		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln(err)
		}

		go func() {
			log.Printf("From: %v Reciving data: %s", addr.String(), string(buf[:n]))
			fmt.Println(string(buf[:n]))
		}()

		localAddr := udpConn.LocalAddr().(*net.UDPAddr).String()
		fmt.Println(localAddr)
	}
	
	// kubeadm joinする　
	// _, err := exec.Command("/usr/bin/","./setup_test.sh").Output()
    // if err != nil {
    //     fmt.Print(err.Error())
    // }

}