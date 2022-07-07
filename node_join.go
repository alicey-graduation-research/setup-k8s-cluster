package main

import (
	//"fmt"
	"net"
	//"os/exec"
	"io"
	"log"
	"os"
	"time"
)

func main() {
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

	// masterにkubeadmにjoinするトークン要求用の定義
	conn, err := net.Dial("udp4", "255.255.255.255:43210")
	if err != nil {
		log.Fatalln("[ERROR]net.Dial: " + err.Error())
	}
	defer conn.Close()

	// token受け取り用の定義
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
	for {
		// control-planeにtokenを要求
		_, err = conn.Write([]byte("please-kubeadm-token"))
		if err != nil {
			log.Println("[INFO]please-kubeadm-token: " + err.Error())
		}

		// UDPパケットを待機する
		err = udpConn.SetReadDeadline(time.Now().Add(10 * time.Second)) // timeout 10 second
		if err != nil {
			log.Fatalln("[ERROR]udpConn: timeout setting error")
		}

		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Println("[ERROR]udpConn.ReadFromUDP: " + err.Error())
		}

		ch := make(chan bool, 1)
		go func() {
			if addr != nil {
				log.Println("resIP: " + addr.String())
				// 内容の判定
				s := string(buf[:n])
				log.Println("From: %v Reciving data: %s", addr.String(), s)

				token_get_flag = true
				ch <- true
			}
		}()

		// データ受け取りORタイムアウト時の処理
		select {
		case <-ch:
			log.Println("[INFO]goroutin done")
		case <-time.After(50 * time.Millisecond):
			log.Println("[ERROR]goroutine time out")
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
