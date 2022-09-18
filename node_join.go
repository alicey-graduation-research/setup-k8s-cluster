package main

import (
	//"fmt"
	"net"
	//"os/exec"
	"io"
	"log"
	"os"
	"time"
	"strings"
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

	buf := make([]byte, 256)
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
				// 内容の検証
				s := string(buf[:n])
				v := strings.Split(s, " ")
				//log.Println("From: %v Reciving data: %s", addr.String(), s)
	
				// log.Println("[DEBUG] len(v):",len(v))
				// log.Println("[DEBUG] v:", v)
				// log.Println("[DEBUG] len(s):",len(s))
				// log.Println("[DEBUG] s:", s)

				if len(v) != 7 {
					log.Fatalln("[Error]return data: Different number of arguments")
				}
				if v[0] != "kubeadm"{
					log.Fatalln("[Error]return data: arg0(kubeadm)")
				}
				if v[1] != "join"{
					log.Fatalln("[Error]return data: arg1(join)")
				}
				// if v[2] != strings.Contains(v[2], str(addr.IP)){
				// 	//K8sコンテナで動かす場合、対応ホストIPとapi-serverが一致せず引っかかるかも
				// 	log.Fatalln("[Error]return data: arg２(ip-addr)")
				// }
				if v[3] != "--token"{
					log.Fatalln("[Error]return data: arg3(--token)")
				}
				// if v[4] != ""{
				// 	log.Fatalln("[Error]return data: arg4(token-data)")
				// }
				if v[5] != "--discovery-token-ca-cert-hash"{
					log.Fatalln("[Error]return data: arg5(--discovery-token-ca-cert-hash)")
				}
				// if v[6] != ""{
				// 	log.Fatalln("[Error]return data: arg6(hash-data)")
				// }

				log.Println("Reciving data: ", s)

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
