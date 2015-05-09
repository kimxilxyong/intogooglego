package main

import (
	"fmt"
	"github.com/kimxilxyong/rpcbotinterfaceobjects"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"runtime"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	InstallCtrlCPanic()
	InstallKillPanic()

	bot := new(rpcbotinterfaceobjects.Bot)
	rpc.Register(bot)

	listener, e := net.Listen("tcp", ":9876")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	fmt.Println("Server listening")

	rpc.Accept(listener)
}

// InstallCtrlCPanic installs a Ctrl-C signal handler that panics
func InstallCtrlCPanic() {
	go func() {
		defer SavePanicTrace()
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		for _ = range ch {
			panic("ctrl-c")
		}
	}()
}

// InstallKillPanic installs a kill signal handler that panics
// From the command-line, this signal is agitated with kill -ABRT
func InstallKillPanic() {
	go func() {
		//defer SavePanicTrace()
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Kill)
		for _ = range ch {
			panic("sigkill")
		}
	}()
}

func SavePanicTrace() {
	r := recover()
	if r == nil {
		return
	}
	// Redirect stderr
	file, err := os.Create("panic")
	if err != nil {
		panic("dumper (no file) " + r.(fmt.Stringer).String())
	}

	//syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd()))
	defer func() { file.Close() }()
	panic("dumper " + r.(string))
}
