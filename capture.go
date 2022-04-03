package trace

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

var cap = (&capture{}).init()

type capture struct {
	users
	file
	pcap
	m         sync.Mutex
	isCapture bool
}

func NewCapture() *capture {
	return cap
}

func (c *capture) init() *capture {
	c.file.SetFilePath("./pcap/" + time.Now().Format("2006-01-02"))
	c.file.SetFileMaxNum(3)
	c.file.SetFileMaxSize(50 * 1024 * 1024)
	c.file.fType = "pcap"
	c.file.bufSize = 8 * 1024
	c.file.header = c.pcap.defaultFileHeader()
	c.users.creator = c
	c.pcap.fileHeader = DefaultPcapHeader().Marshal()
	return c
}

func (c *capture) Start() error {
	c.m.Lock()
	defer c.m.Unlock()
	if c.isCapture {
		return fmt.Errorf("capture is already started")
	}
	c.isCapture = true
	return nil
}

func (c *capture) Stop() error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.isCapture {
		return fmt.Errorf("capture is already stopped")
	}
	c.isCapture = false
	return c.users.Close()
}

func (c *capture) IsStop() bool {
	if !c.isCapture {
		return true
	}
	return false
}

func (c *capture) NewWriteCloser(stringer fmt.Stringer) io.WriteCloser {
	return &file{
		maxSize: c.file.maxSize,
		maxNum:  c.file.maxNum,
		path:    c.file.path,
		fType:   c.file.fType,
		header:  c.pcap.defaultFileHeader(),
		bufSize: c.file.bufSize,
		user:    stringer,
	}
}

func (c *capture) WritePcap(data []byte, id fmt.Stringer) error {
	if !c.isCapture {
		return nil
	}
	return c.users.Write([]io.Reader{
		c.pcap.newPacketHeader(len(data)),
		bytes.NewReader(data)},
		id,
	)
}
