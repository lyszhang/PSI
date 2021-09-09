/**
 * @Author: lyszhang
 * @Email: ericlyszhang@gmail.com
 * @Date: 2021/8/17 2:40 PM
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	//_ "net/http/pprof"
)

var (
	version string
)

const (
	tcpModeRaw   = "raw"
	tcpModeGmssl = "gmssl"
)

func ParseArgs() (string, string, string, string) {
	listenAddr := flag.String("l", ":8080", "listen address")
	listenMode := flag.String("lm", tcpModeRaw, "listen mode, \"raw\" or \"gmssl\", raw socket or gmssl encrypted")
	forwardAddr := flag.String("f", "", "forwarding address")
	forwardMode := flag.String("fm", "", "forwarding mode, \"raw\" or \"gmssl\", raw socket or gmssl encrypted")
	flagVersion := flag.Bool("v", false, "print version")

	flag.Parse()

	if *flagVersion {
		fmt.Println("version:", version)
		os.Exit(0)
	}

	if *forwardAddr == "" {
		flag.Usage()
		os.Exit(0)
	}

	return *listenAddr, *listenMode, *forwardAddr, *forwardMode
}

func HandleRequest(conn net.Conn, forwardAddr string) {
	d := net.Dialer{Timeout: time.Second * 10}

	proxy, err := d.Dial("tcp", forwardAddr)
	if err != nil {
		log.Printf("try connect %s -> %s failed: %s\n", conn.RemoteAddr(), forwardAddr, err.Error())
		conn.Close()
		return
	}
	log.Printf("connected: %s -> %s\n", conn.RemoteAddr(), forwardAddr)

	Pipe(conn, proxy)
}

func Pipe(src net.Conn, dest net.Conn) {
	var (
		readBytes  int64
		writeBytes int64
	)
	ts := time.Now()

	wg := sync.WaitGroup{}
	wg.Add(1)

	closeFun := func(err error) {
		dest.Close()
		src.Close()
	}

	log.Println("io copy start1")
	go func() {
		defer wg.Done()
		//buffer:= make([]byte, 1024*1024*200)
		//n, err := io.CopyBuffer(dest, src, buffer)
		//payload, err := ioutil.ReadAll(src)
		//n, err := dest.Write(payload)
		n, err := Copy(dest,src)
		readBytes += n
		closeFun(err)
	}()

	//buffer:= make([]byte, 1024*1024*200)
	//n, err := io.CopyBuffer(src, dest, buffer)
	//payload, err := ioutil.ReadAll(dest)
	//n, err := src.Write(payload)
	n, err := Copy(src,dest)
	writeBytes += n
	closeFun(err)
	log.Println("io copy end")

	wg.Wait()
	log.Printf("connection %s -> %s closed: readBytes %d, writeBytes %d, duration %s", src.RemoteAddr(), dest.RemoteAddr(), readBytes, writeBytes, time.Now().Sub(ts))
}

func Proxy(listenAddr, listenMode, forwardAddr, forwardMode string) {
	// forwarding mode
	var handler func(conn net.Conn, forwardAddr string)
	if forwardMode == tcpModeGmssl {
		handler = HandleRequestSSL
	} else {
		handler = HandleRequest
	}

	// listening mode
	switch listenMode {
	case tcpModeGmssl:
		{
			ListenAndServeSSL(listenAddr, forwardAddr, handler)
		}
	default:
		{
			ListenAndServe(listenAddr, forwardAddr, handler)
		}

	}

}

var f *os.File

// 优雅退出
func waitExit(c chan os.Signal) {
	for i := range c {
		switch i {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL:
			log.Println("receive exit signal ", i.String(), ",exit...")

			// CPU 性能分析
			f.Close()
			pprof.StopCPUProfile()

			os.Exit(0)
		}
	}
}

func main() {
	listenAddr, listenMode, forwardAddr, forwardMode := ParseArgs()

	f, err := os.OpenFile("cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	pprof.StartCPUProfile(f)
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,   syscall.SIGQUIT, syscall.SIGKILL)

	go func() {
		waitExit(c)
	}()

	Proxy(listenAddr, listenMode, forwardAddr, forwardMode)
}
