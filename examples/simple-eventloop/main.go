package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ii64/gouring/examples/simple-eventloop/lib"
)

type myEchoServer struct{}

func (h myEchoServer) OnAccept(ctx lib.Context, sa syscall.Sockaddr) {
	fmt.Printf("accept: %+#v\n", sa)
	ctx.SetContext(context.Background())
	ctx.Read()
}
func (h myEchoServer) OnRead(ctx lib.Context, b []byte) {
	sctx := ctx.Context()
	fmt.Printf("read ctx %+#v %+#v\n", sctx, b)
	ctx.Write(b)
}
func (h myEchoServer) OnWrite(ctx lib.Context, nb int) {
	ctx.Read()
}
func (h myEchoServer) OnClose(ctx lib.Context) {
}

type myHTTP11Server struct{}

func (h myHTTP11Server) OnAccept(ctx lib.Context, sa syscall.Sockaddr) {
	ctx.Read()
}
func (h myHTTP11Server) OnRead(ctx lib.Context, b []byte) {
	statusCode := http.StatusOK

	if !bytes.HasPrefix(b, []byte("GET /")) {
		statusCode = 400
	}

	statusText := http.StatusText(statusCode)
	header := []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\nServer: gouring-simple-evloop\r\nConnection: closed\r\nContent-Length: %d\r\n\r\n",
		statusCode, statusText,
		len(b)))
	buf := make([]byte, len(header)+len(b))
	copy(buf[0:], header)
	copy(buf[len(header):], b)

	ctx.Write(buf)
}
func (h myHTTP11Server) OnWrite(ctx lib.Context, nb int) {
	ctx.Close()
}
func (h myHTTP11Server) OnClose(ctx lib.Context) {
}

func runServer(addr string, handler lib.EventHandler) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	file, err := lis.(*net.TCPListener).File()
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fd := file.Fd()
	evloop := lib.New(64, int(fd), handler)
	evloop.Run()
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go runServer("0.0.0.0:11338", myEchoServer{})
	go runServer("0.0.0.0:11339", myHTTP11Server{})

	<-sig
}
