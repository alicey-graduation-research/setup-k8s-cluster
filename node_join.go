package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var port string
var api_server string
var token_server_port string
var cluster_destruction bool
var new_comer bool

func main() {
	// logをファイル書き込み
	// logfile, err := os.OpenFile("./udpServer.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	println("cannot open logfile", err)
	// }
	// defer logfile.Close()
	// log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetOutput(io.MultiWriter(os.Stdout))

	//諸々の設定は環境変数から読み込みたい
	port = "32432"
	api_server = "127.0.0.1"
	api_server = "172.10.200.14"
	token_server_port = "32765"
	cluster_destruction = true
	new_comer = false

	fail_counter := 0

	for {
		cluster_check_res, err := cluster_status_check()
		if err != nil {
			//未構築で証明書がない場合などもここに入る
			log.Println("[INFO] cluster_check:", err.Error())
		}
		// 新規参加時の挙動
		if new_comer == true {
			err = cluster_join()
			if err != nil {
				log.Println("[ERROR] cluster join:", err.Error())
				fail_counter++
				continue
			}
			new_comer = false
			fail_counter = 0
			continue
		}

		// クラスタ再構築時の挙動
		if cluster_check_res == true {
			//クラスタ壊したり再構築したり
			if cluster_destruction {
				err = cluster_destroy()
				if err != nil {
					log.Println("[ERROR]Cluster Destroy: " + err.Error())
					fail_counter++
					continue
				}
				err := cluster_join()
				if err != nil {
					log.Println("[ERROR] cluster join:", err.Error())
					fail_counter++
					continue
				}
				fail_counter = 0
				continue
			} else {
				// 疎通取れないが再構築許されてないので何もできない
				log.Println("[ERROR] Unable to connect to cluster. Try to rebuild or check the connection.")
			}
		}
		if fail_counter > 0 {
			log.Println("[ERROR_INFO] process fail count:", fail_counter)
		} else if fail_counter > 100 {
			//Fatallnにするか迷う
			log.Println("[ERROR] Critical error, Please Check the connection to the cluster.", fail_counter)
		}

		//何もなかったとき
		fail_counter = 0
		time.Sleep(time.Second * 120)
	}
}

// チェック問題なしfalse, 問題ありTrue
func cluster_status_check() (bool, error) {
	for count := 0; count < 5; count++ {
		cmd := "curl -i --cacert /etc/kubernetes/pki/ca.crt https://" + api_server + ":6443/version"
		// r, err := exec.Command("/usr/bin/curl","--cacert","/etc/kubernetes/pki/ca.crt","https://" + api_server + ":6443/version").Output()
		r, err := exec.Command("sh", "-c", cmd).Output()
		if err != nil {
			log.Println("[ERROR]exec.Command curl: " + err.Error())
			return true, err
		}
		//log.Println("[INFO]curl exec:\n", string(r))

		res := true
		for _, s := range []string{"200", "kubernetes", "platform"} {
			res = res && strings.Contains(string(r), s)
			//fmt.Println(strings.Contains(string(r), s))
		}
		if res == true {
			//fmt.Println("validate res:" + strconv.FormatBool(res))
			return false, nil
		}
		fmt.Println("status check fallen: count " + strconv.Itoa(count))
		time.Sleep(time.Second * 5)
	}
	return true, errors.New("cluster status fail")
}

func cluster_destroy() error {
	cmd := "echo Y | kubeadm reset"
	r, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		//log.Println("[ERROR]exec.Command kubeadm reset: " + err.Error())
		return err
	}
	log.Println("[INFO] kubeadm reset:", r)
	return nil
}

func cluster_join() error {
	// os.Setenv("TEST","testenv")
	// fmt.Print(string(os.Getenv("TEST")))

	// nodeにK8sコンポーネントのインストール
	// out, err := exec.Command("/bin/sh","./setup_component.sh").Output()
	// if err != nil {
	// 	log.Print("[ERROR]" + string(out))
	// 	log.Fatalln("[ERROR]K8s compornent install:" + err.Error())
	// }

	// masterにkubeadmにjoinするトークン要求用の定義
	conn, err := net.Dial("udp4", "255.255.255.255:"+token_server_port)
	if err != nil {
		log.Println("[ERROR]net.Dial: " + err.Error())
		return err
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
		log.Println("[ERROR]net.ListenUDP: " + err.Error())
		return err
	}

	buf := make([]byte, 256)
	log.Println("[INFO]Starting UDP Server...")

	var kubeadm_command string
	var response_validate_flag bool
	for {
		// control-planeにtokenを要求
		_, err = conn.Write([]byte("please-kubeadm-token"))
		if err != nil {
			log.Println("[INFO]please-kubeadm-token: " + err.Error())
			return err
		}

		// UDPパケットを待機する
		err = udpConn.SetReadDeadline(time.Now().Add(10 * time.Second)) // timeout 10 second
		if err != nil {
			log.Println("[ERROR]udpConn: timeout setting error")
			return err
		}

		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Println("[ERROR]udpConn.ReadFromUDP: " + err.Error())
			return err
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

				response_validate_flag = false

				if len(v) != 7 {
					log.Println("[Error]return data: Different number of arguments")
					response_validate_flag = true
				}
				if v[0] != "kubeadm" {
					log.Println("[Error]return data: arg0(kubeadm)")
					response_validate_flag = true
				}
				if v[1] != "join" {
					log.Println("[Error]return data: arg1(join)")
					response_validate_flag = true
				}
				if !check_regexp(`^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])$`, v[2]) {
					log.Println("[Error]return data: arg2(ip-addr)")
					response_validate_flag = true
				}
				if v[3] != "--token" {
					log.Println("[Error]return data: arg3(--token)")
					response_validate_flag = true
				}
				if !check_regexp(`^[a-zA-Z.]{23}&`, v[4]) {
					log.Println("[Error]return data: arg4(token-data)")
					response_validate_flag = true
				}
				if v[5] != "--discovery-token-ca-cert-hash" {
					log.Println("[Error]return data: arg5(--discovery-token-ca-cert-hash)")
					response_validate_flag = true
				}
				if !check_regexp(`[a-zA-Z]{64}`, v[6]) {
					log.Println("[Error]return data: arg6(hash-data)")
					response_validate_flag = true
				}
				log.Println("[INFO]Reciving data: ", s)

				kubeadm_command = s
				ch <- true
			}
		}()
		if response_validate_flag != false {
			return errors.New("response validate error")
		}

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
	r, err := exec.Command("sh", "-c", kubeadm_command).Output()
	if err != nil {
		log.Println("[ERROR]exec.Command kubeadm join: " + err.Error())
		return err
	}
	log.Println("[INFO]kubeadm exec: ", r)

	time.Sleep(time.Second * 120)
	return nil
}

func check_regexp(reg string, str string) bool {
	return regexp.MustCompile(reg).Match([]byte(str))
}
