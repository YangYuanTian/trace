package trace

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"time"
)

type pcap struct {
	fileHeader   []byte
	packetHeader PacketHeader
}

func (p *pcap) defaultFileHeader() []byte {
	return p.fileHeader
}
func (p *pcap) newPacketHeader(l int) io.Reader {
	return p.packetHeader.NewPacketHeader(l)
}

type PcapGlobalHear struct {
	Magic    uint32 //：4Byte：标记文件开始，并用来识别文件自己和字节顺序。0xa1b2c3d4用来表示按照原来的顺序读取，0xd4c3b2a1表示下面的字节都要交换顺序读取。考虑到计算机内存的存储结构，一般会采用0xd4c3b2a1，即所有字节都需要交换顺序读取。
	Major    uint16 //：2Byte： 当前文件主要的版本号，一般为 0x0200【实际上因为需要交换读取顺序，所以计算机看到的应该是 0x0002】
	Minor    uint16 //：2Byte： 当前文件次要的版本号，一般为 0x0400【计算机看到的应该是 0x0004】
	ThisZone uint32 //：4Byte：当地的标准时间，如果用的是GMT则全零，一般都直接写
	SigFigs  uint32 //：4Byte：时间戳的精度，设置为 全零 即可
	SnapLen  uint32 //：4Byte：最大的存储长度，如果想把整个包抓下来，设置为 ，但一般来说 ff7f 0000就足够了【计算机看到的应该是 0000 ff7f 】
	LinkType uint32 //：4Byte：链路类型  以太网或者环路类型为
}

func DefaultPcapHeader() *PcapGlobalHear {
	h := &PcapGlobalHear{}
	h.Magic = 0xa1b2c3d4
	h.Major = 0x0002
	h.Minor = 0x0004
	h.SnapLen = 0x0000ff7f
	h.LinkType = 1
	return h
}

func (h *PcapGlobalHear) Marshal() []byte {
	buf := make([]byte, 24)
	binary.LittleEndian.PutUint32(buf[0:4], h.Magic)
	binary.LittleEndian.PutUint16(buf[4:6], h.Major)
	binary.LittleEndian.PutUint16(buf[6:8], h.Minor)
	binary.LittleEndian.PutUint32(buf[8:12], h.ThisZone)
	binary.LittleEndian.PutUint32(buf[12:16], h.SigFigs)
	binary.LittleEndian.PutUint32(buf[16:20], h.SnapLen)
	binary.LittleEndian.PutUint32(buf[20:24], h.LinkType)
	return buf
}

type PacketHeader struct {
	TimestampH uint32 //：被捕获时间的高位，单位是seconds
	TimestampL uint32 //：被捕获时间的低位，单位是microseconds
	CapLen     uint32 //：当前数据区的长度，即抓取到的数据帧长度，不包括Packet Header本身的长度，单位是 Byte ，由此可以得到下一个数据帧的位置。
	Len        uint32 //：离线数据长度：网络中实际数据帧的长度，一般不大于caplen，多数情况下和Caplen数值相等。
}

func (p *PacketHeader) Marshal() []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint32(buf[0:4], p.TimestampH)
	binary.LittleEndian.PutUint32(buf[4:8], p.TimestampL)
	binary.LittleEndian.PutUint32(buf[8:12], p.CapLen)
	binary.LittleEndian.PutUint32(buf[12:16], p.Len)
	return buf
}

func (p *PacketHeader) NewPacketHeader(l int) io.Reader {
	p.Len = uint32(l)
	p.CapLen = p.Len
	t2 := time.Now().UnixNano()
	p.TimestampH = uint32(t2 / 1e9)       //单位是s
	p.TimestampL = uint32(t2 % 1e9 / 1e3) //单位是ms
	return bytes.NewReader(p.Marshal())
}

//for test

func PacketGenerator(Data []byte, Port uint16) []byte {
	var port = make([]byte, 2, 2)
	binary.BigEndian.PutUint16(port, Port)
	//ip                                          ps            pd
	head := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xa4, 0x16, 0xe7, 0x7c, 0xf5, 0x1e, 0x08, 0x00, 0x45, 0xc0, 0x01, 0x71, 0x44, 0xfd, 0x00, 0x00, 0xff, 0x11, 0x74, 0xbf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, port[0], port[1], port[0], port[1], 0x01, 0x5d, 0x00, 0x00}
	ipLen := 16
	udpLen := 38
	data := Data
	dataLen := uint16(len(data) + 8)
	buf := make([]byte, 2, 2)
	binary.BigEndian.PutUint16(buf[0:2], dataLen)
	head[udpLen] = buf[0]
	head[udpLen+1] = buf[1]
	binary.BigEndian.PutUint16(buf[0:2], dataLen+20)
	head[ipLen] = buf[0]
	head[ipLen+1] = buf[1]
	return append(head, data...)
}

func isDirExist(dir string) bool {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
