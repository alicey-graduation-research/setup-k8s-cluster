package main

import(
	"fmt"
	"net"
	//"os/exec"
	"os"
	"io"
	"log"
	"time"
)

func main(){
	// logをファイル書き込み
	// logfile, err := os.OpenFile("./udpServer.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	println("cannot open logfile", err)
	// }
	// defer logfile.Close()
	// log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetOutput(io.MultiWriter(os.Stdout))

	// os.Setenv("TEST","testenv")
	// fmt.Print(string(os.Getenv("TEST")))

	// nodeにK8sコンポーネントのインストール
	// out, err := exec.Command("/bin/sh","./setup_component.sh").Output()
    // if err != nil {
	// 	log.Print("[ERROR]" + string(out))
	// 	log.Fatalln("[ERROR]K8s compornent install:" + err.Error())
    // }

	// masterにkubeadmにjoinするトークン要求
	conn, err := net.Dial("udp4", "255.255.255.255:43210")
	if err != nil {
		log.Fatalln("[ERROR]net.Dial: " + err.Error())
	}
	defer conn.Close()

	// token受け取り
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 32432,
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalln("[ERROR]net.ListenUDP: " + err.Error())
	}

	buf := make([]byte, 64)
	log.Println("[INFO]Starting UDP Server...")

	token_get_flag := false
	for{
		// control-planeにtokenを要求
		_, err = conn.Write([]byte("please-kubeadm-token"))
		if err != nil {
			log.Println("[INFO]please-kubeadm-token: " + err.Error())
		}

		start_time := time.Now()
		for {
			// UDPパケットを待機するループ
			n, addr, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				log.Println("[ERROR]udpConn.ReadFromUDP: " + err.Error())
			}

			go func() {
				log.Println("From: %v Reciving data: %s", addr.String(), string(buf[:n]))
				fmt.Println(string(buf[:n]))
				token_get_flag = true
			}()

			//localAddr := udpConn.LocalAddr().(*net.UDPAddr).String()
			//fmt.Println(localAddr)

			if token_get_flag{
				break
			}

			if time.Since(start_time).Milliseconds() > 1000{
				log.Println("[ERROR]UDP packet wait: time out")
				break
			}
		}

		if token_get_flag {
			log.Println("[INFO]token get")
			break
		}
	}
	
	// kubeadm joinする　
	// _, err := exec.Command("/usr/bin/","./setup_test.sh").Output()
    // if err != nil {
    //     log.Fatalln("[ERROR]exec.Command kubeadm join: " + err.Error())
    // }

}