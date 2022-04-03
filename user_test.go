package trace

import (
	"fmt"
	"runtime/debug"
	"testing"
)

func ErrLog() {
	if err := recover(); err != nil {
		fmt.Printf("%+v\n", err)
		debug.PrintStack()
	}
}
func E(err error) {
	if err != nil {
		panic(err)
	}
}

func TestCapture(t *testing.T) {
	c := NewCapture()

	defer ErrLog()

	E(c.Start())

	defer func() {
		if er := recover(); er != nil {
			fmt.Println(er)
			debug.PrintStack()
		}
		E(c.Stop())
	}()

	c.AddUser(ID("TanGao"), ID("LiHua"))
	c.UseTrace(true).Trace(ID("TanGao"), true)
	for i := 0; i < 10000000; i++ {
		//time.Sleep(time.Second)
		E(c.WritePcap(PacketGenerator([]byte("hello,TanGao"), 19), ID("TanGao")))
		E(c.WritePcap(nil, ID("TanGao")))
		E(c.WritePcap(PacketGenerator([]byte("hello,LiHua"), 19), ID("LiHua")))
		E(c.WritePcap(PacketGenerator([]byte("hello,JiangYan"), 19), ID("JiangYan")))
		E(c.WritePcap(PacketGenerator([]byte("hello,ZiZhu"), 19), ID("ZiZhu")))
	}
	fmt.Println(c.GetUsrInfo(ID("TanGao")))
	fmt.Println(c.GetUsrInfo(ID("ZiZhu")))
	fmt.Println(c.ListUsers())
}

func TestRestart(t *testing.T) {
	for i := 0; i < 1; i++ {
		TestCapture(t)
	}
}
