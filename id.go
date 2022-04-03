package trace

import (
	"fmt"
	"net"
)

const (
	PrefixTEID         = "TEID_0x"
	PrefixSEID         = "SEID_0x"
	PrefixUEIP         = "UEIP_"
	PrefixDLTeid       = "DLTEID_0x"
)

type TEID  uint32

func (t TEID) String() string {
	return fmt.Sprintf("%s%X", PrefixTEID, t)
}

type DLTeid uint32
func (t DLTeid) String() string {
	return fmt.Sprintf("%s%X", PrefixDLTeid, t)
}

type  UEIP  net.IP
func (u UEIP) String() string {
	return fmt.Sprintf("%s%s", PrefixUEIP, net.IP(u).String())
}

type ID string
func (i ID) String() string {
	return string(i)
}

type SEID uint64
func (s SEID) String() string {
	return fmt.Sprintf("%s%X", PrefixSEID, s)
}
