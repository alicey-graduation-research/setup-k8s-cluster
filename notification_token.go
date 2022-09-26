package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
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

	// UDP Server
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 43210,
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
			log.Println(err)
		}

		go func() {
			// log.Printf("From: %v Reciving data: %s", addr.String(), string(buf[:n]))
			// log.Println("[INFO]receive: " + addr.String())
			s := string(buf[:n])
			// log.Println("[DEBUG]",s)

			if s == "please-kubeadm-token" {
				log.Println("[INFO] get token request: ", addr.IP)
				kubeadm_command, err := exec.Command("kubeadm token create --print-join-command").Output()
				if err != nil {
					log.Println("[ERROR]exec.Command kubeadm token create: " + err.Error())
					return
				}
				// kubeadm_command := "echo aaa"

				// log.Println("[DEBUG]IP-base",  fmt.Sprintf("%s", addr.IP) )
				// log.Println("[DEBUG]IP-str", fmt.Sprintf("%s", addr.IP)+":32432")
				sendConn, err := net.Dial("udp4", fmt.Sprintf("%s", addr.IP)+":32432")
				if err != nil {
					log.Println("[ERROR]net.Dial: " + err.Error())
					return
				}
				defer sendConn.Close()

				log.Println("[INFO]Sending token")
				_, err = sendConn.Write([]byte(kubeadm_command))
				if err != nil {
					log.Println("[ERROR]send token: " + err.Error())
					return
				}

			}
		}()
	}
}
