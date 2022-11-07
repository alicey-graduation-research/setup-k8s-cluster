package main

import (
	//"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
	"strconv"
)

var port string
var api_server string
var token_server_port string
//諸々の設定は環境変数から読み込みたい

func main(){
	// logをファイル書き込み
	// logfile, err := os.OpenFile("./udpServer.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	println("cannot open logfile", err)
	// }
	// defer logfile.Close()
	// log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetOutput(io.MultiWriter(os.Stdout))

	port = "32432"
	api_server = "127.0.0.1"
	api_server = "172.10.200.14"
	token_server_port = "32765"

	cluster_status_check()
	//cluster_join(log)
	time.Sleep(time.Second * 60) 
}

func cluster_status_check(){
	cmd := "'/usr/bin/curl --cacert /etc/kubernetes/pki/ca.crt https://" + api_server + ":6443/version'"
	// r, err := exec.Command("/usr/bin/curl","--cacert","/etc/kubernetes/pki/ca.crt","https://" + api_server + ":6443/version").Output()
	r, err := exec.Command("/bin/sh","-c",cmd).CombinedOutput()
	if err != nil {
		log.Fatalln("[ERROR]exec.Command curl: " + err.Error())
	}
	log.Println("[INFO]curl exec: ", r)
}

func cluster_join() {
	// os.Setenv("TEST","testenv")
	// fmt.Print(string(os.Getenv("TEST")))

	// nodeにK8sコンポーネントのインストール
	// out, err := exec.Command("/bin/sh","./setup_component.sh").Output()
	// if err != nil {
	// 	log.Print("[ERROR]" + string(out))
	// 	log.Fatalln("[ERROR]K8s compornent install:" + err.Error())
	// }

	// masterにkubeadmにjoinするトークン要求用の定義
	conn, err := net.Dial("udp4", "255.255.255.255:" + token_server_port)
	if err != nil {
		log.Fatalln("[ERROR]net.Dial: " + err.Error())
	}
	defer conn.Close()

	// token受け取り用の定義
	p, _ := strconv.Atoi(port)
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: p,
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalln("[ERROR]net.ListenUDP: " + err.Error())
	}

	buf := make([]byte, 256)
	log.Println("[INFO]Starting UDP Server...")

	var kubeadm_command string
	for {
		// control-planeにtokenを要求
		_, err = conn.Write([]byte("please-kubeadm-token"))
		if err != nil {
			log.Fatalln("[INFO]please-kubeadm-token: " + err.Error())
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
				if v[0] != "kubeadm" {
					log.Fatalln("[Error]return data: arg0(kubeadm)")
				}
				if v[1] != "join" {
					log.Fatalln("[Error]return data: arg1(join)")
				}
				// if v[2] != strings.Contains(v[2], str(addr.IP)){
				// 	//K8sコンテナで動かす場合、対応ホストIPとapi-serverが一致せず引っかかるかも
				// 	log.Fatalln("[Error]return data: arg2(ip-addr)")
				// }
				if v[3] != "--token" {
					log.Fatalln("[Error]return data: arg3(--token)")
				}
				// if v[4] != ""{
				// 	log.Fatalln("[Error]return data: arg4(token-data)")
				// }
				if v[5] != "--discovery-token-ca-cert-hash" {
					log.Fatalln("[Error]return data: arg5(--discovery-token-ca-cert-hash)")
				}
				// if v[6] != ""{
				// 	log.Fatalln("[Error]return data: arg6(hash-data)")
				// }

				log.Println("[INFO]Reciving data: ", s)

				kubeadm_command = s
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

		if kubeadm_command != "" {
			log.Println("[INFO]token get")
			break
		}
	}

	//log.Println("[DEBUG] ", kubeadm_command)
	// kubeadm joinする
	r, err := exec.Command(kubeadm_command).Output()
	if err != nil {
		log.Fatalln("[ERROR]exec.Command kubeadm join: " + err.Error())
	}
	log.Println("[INFO]kubeadm exec: ", r)

}
