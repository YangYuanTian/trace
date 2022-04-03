package trace

import (
	"fmt"
	"runtime/debug"
	"testing"
	"time"
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

type id uint64

func (i id) String() string {
	return fmt.Sprintf("SEID_%d", i)
}

func TestCapture(t *testing.T) {
	cap := NewCapture()

	defer ErrLog()

	E(cap.Start())

	defer func() {
		if er:=recover();er!=nil{
			fmt.Println(er)
			debug.PrintStack()
		}
		E(cap.Stop())
	}()

	cap.SetUserIDs(id(1), id(2),id(3),id(4))

	for i := 0; i < 100000; i++ {
		E(cap.WritePcap(PacketGenerator([]byte("hello,world"), 19), id(1)))
		E(cap.WritePcap(PacketGenerator([]byte("hello,world"), 19), id(2)))
		E(cap.WritePcap(PacketGenerator([]byte("hello,world"), 19), id(3)))
		E(cap.WritePcap(PacketGenerator([]byte("hello,world"), 19), id(4)))
	}

	time.Sleep(time.Second)
}
