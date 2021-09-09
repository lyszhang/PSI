/**
 * @Author: lyszhang
 * @Email: ericlyszhang@gmail.com
 * @Date: 2021/8/17 4:22 PM
 */

package main

import (
	"fmt"
	tls "github.com/tjfoc/gmtls"
	"log"
	"net"
)

// HandleRequestSSL tcp 转发 gmssl协议
func HandleRequestSSL(conn net.Conn, forwardAddr string) {
	cers, err := loadCerts()
	if err != nil {
		log.Println("server_echo : loadCerts err->", err)
		return
	}
	conf := &tls.Config{
		InsecureSkipVerify: true, //为true 接收任何服务端的证书不做校验
		Certificates: cers,
	}
	proxy, err := tls.Dial("tcp", forwardAddr, conf)
	if err != nil {
		log.Printf("try connect %s -> %s failed: %s\n", conn.RemoteAddr(), forwardAddr, err.Error())
		conn.Close()
		return
	}

	log.Printf("connected: %s -> %s\n", conn.RemoteAddr(), forwardAddr)

	Pipe(conn, proxy)
}

// ListenAndServeSSL 监听ssl tcp
func ListenAndServeSSL(listenAddr string, forwardAddr string, handler func(conn net.Conn, forwardAddr string)) {
	cers, err := loadCerts()
	if err != nil {
		log.Println("server_echo : loadCerts err->", err)
		return
	}
	config := &tls.Config{Certificates: cers, ClientAuth: tls.RequireAnyClientCert}
	ln, err := tls.Listen("tcp", listenAddr, config)
	if err != nil {
		log.Println("server_echo : Listen err->", err)
		return
	}
	defer ln.Close()
	log.Printf("accept %s to %s\n", listenAddr, forwardAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept %s error: %s\n", listenAddr, err.Error())
		}

		go handler(conn, forwardAddr)
	}
}

func ListenAndServe(listenAddr string, forwardAddr string, handler func(conn net.Conn, forwardAddr string)) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("listen addr %s failed: %s", listenAddr, err.Error())
	}

	log.Printf("accept %s to %s\n", listenAddr, forwardAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept %s error: %s\n", listenAddr, err.Error())
		}

		go handler(conn, forwardAddr)
	}
}

func loadCerts() ([]tls.Certificate, error) {
	pemdir := "./sm2Certs"
	cerfiles := []string{"SS", "CA", "SE"}
	certs := make([]tls.Certificate, 0)
	for _, n := range cerfiles {
		certname := fmt.Sprintf("%s/%s.cert.pem", pemdir, n)
		certkey := fmt.Sprintf("%s/%s.key.pem", pemdir, n)
		cer, err := tls.LoadX509KeyPair(certname, certkey)
		if err != nil {
			log.Println("tls.LoadX509KeyPair err->", err, " name=", certname, " key=", certkey)
			return nil, err
		}
		certs = append(certs, cer)
	}
	return certs, nil

}
