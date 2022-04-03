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
	users     users
	files     file
	pcap      pcap
	status    bool
	whiteList sync.Map
	isUseWL   bool
	m         sync.Mutex
	isCapture bool
}

func NewCapture() *capture {
	return cap
}

func (c *capture)SetUserIDs(ids ...fmt.Stringer) {
	c.users.SetNames(ids...)
}

func (c *capture) init() *capture {
	c.files.SetFilePath("./pcap/" + time.Now().Format("2006-01-02"))
	c.files.SetFileMaxNum(3)
	c.files.SetFileMaxSize(50 * 1024 * 1024)
	c.files.fType = "pcap"
	c.files.bufSize = 4 * 1024 * 1024
	c.files.header = c.pcap.defaultFileHeader()
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
	c.status = true
	c.isCapture = true
	return nil
}

func (c *capture) Stop() error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.isCapture {
		return fmt.Errorf("capture is already stopped")
	}
	c.status = false
	c.isCapture = false
	return c.users.Close()
}

func (c *capture) IsStop() bool {
	if !c.isCapture || !c.status {
		return true
	}
	return false
}

func (c *capture) NewWriteCloser(stringer fmt.Stringer) io.WriteCloser {
	return &file{
		maxSize: c.files.maxSize,
		maxNum:  c.files.maxNum,
		path:    c.files.path,
		fType:   c.files.fType,
		header:  c.pcap.defaultFileHeader(),
		user:    stringer,
	}
}


func (c *capture) WritePcap(data []byte, id fmt.Stringer) error {
	if !c.status {
		return nil
	}
	if c.isUseWL {
		if _, ok := c.whiteList.Load(id); !ok {
			return nil
		}
	}
	if len(data) == 0 {
		return nil
	}
	return c.users.Write([]io.Reader{
		c.pcap.newPacketHeader(len(data)),
		bytes.NewReader(data)},
		id,
	)
}
