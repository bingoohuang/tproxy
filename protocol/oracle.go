package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

type oracleInterop struct{}

func (o *oracleInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	_ = o.Handler(r, source)
}

const (
	TNS_HEAD_LEN = 8

	OFFSET_TNS_LEN        = 0
	OFFSET_TNS_CKSUM      = 2
	OFFSET_TNS_TYPE       = 4
	OFFSET_TNS_RESV       = 5
	OFFSET_TNS_HEAD_CKSUM = 6
	OFFSET_TNS_DATA       = TNS_HEAD_LEN

	OFFSET_TNS_CONNECT_VERSION            = 0
	OFFSET_TNS_CONNECT_VERSION_COMPATIBLE = 2
	OFFSET_TNS_CONNECT_DATA_LEN           = 16
	OFFSET_TNS_CONNECT_DATA_OFFSET        = 18

	OFFSET_TNS_ACCEPT_VERSION     = 0
	OFFSET_TNS_ACCEPT_DATA_LEN    = 30
	OFFSET_TNS_ACCEPT_DATA_OFFSET = 32

	OFFSET_TNS_REFUSE_PROCESS_CAUSE = 0
	OFFSET_TNS_REFUSE_SYS_CAUSE     = 1
	OFFSET_TNS_REFUSE_DATA_LEN      = 2
	OFFSET_TNS_REFUSE_DATA_OFFSET   = 4

	OFFSET_TNS_DATA_FALG   = 0
	OFFSET_TNS_DATA_ID     = 2
	OFFSET_TNS_DATA_ID_SUB = 3
	OFFSET_TNS_DATA_OFFSET = 4
)

func PKT_LEN(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_LEN:])
}

func PKT_CKSUM(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_CKSUM:])
}

func PKT_TYPE(pkt []byte) PktType {
	return PktType(pkt[OFFSET_TNS_TYPE])
}

func PKT_RESV(pkt []byte) byte {
	return pkt[OFFSET_TNS_RESV]
}

func PKT_HEAD_CKSUM(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_HEAD_CKSUM:])
}

func PKT_DATA(pkt []byte) []byte {
	return pkt[OFFSET_TNS_DATA:]
}

func CONNECT_VER(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_CONNECT_VERSION:])
}

func CONNECT_VER_COMPATIBLE(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_CONNECT_VERSION_COMPATIBLE:])
}

func CONNECT_DATA_LEN(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_CONNECT_DATA_LEN:])
}

func CONNECT_DATA_OFFSET(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_CONNECT_DATA_OFFSET:])
}

func CONNECT_DATA(pkt []byte) []byte {
	offset := CONNECT_DATA_OFFSET(pkt)
	return pkt[offset:]
}

func ACCEPT_VER(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_ACCEPT_VERSION:])
}

func ACCEPT_DATA_LEN(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_ACCEPT_DATA_LEN:])
}

func ACCEPT_DATA_OFFSET(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_ACCEPT_DATA_OFFSET:])
}

func ACCEPT_DATA(pkt []byte) []byte {
	offset := ACCEPT_DATA_OFFSET(pkt)
	return pkt[offset:]
}

func REFUSE_PRO_CAUSE(pkt []byte) byte {
	return pkt[OFFSET_TNS_REFUSE_PROCESS_CAUSE]
}

func REFUSE_SYS_CAUSE(pkt []byte) byte {
	return pkt[OFFSET_TNS_REFUSE_SYS_CAUSE]
}

func REFUSE_DATA_LEN(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_REFUSE_DATA_LEN:])
}

func REFUSE_DATA(pkt []byte) []byte {
	return pkt[OFFSET_TNS_REFUSE_DATA_OFFSET:]
}

func DATA_FLAG(pkt []byte) uint16 {
	return binary.BigEndian.Uint16(pkt[OFFSET_TNS_DATA_FALG:])
}

func DATA_ID(pkt []byte) byte {
	return pkt[OFFSET_TNS_DATA_ID]
}

func DATA_ID_SUB(pkt []byte) byte {
	return pkt[OFFSET_TNS_DATA_ID_SUB]
}

func DATA_DATA(pkt []byte) []byte {
	return pkt[OFFSET_TNS_DATA_OFFSET:]
}

type PktType byte

const (
	TNS_TYPE_CONNECT PktType = iota + 1
	TNS_TYPE_ACCEPT
	TNS_TYPE_ACK
	TNS_TYPE_REFUSE
	TNS_TYPE_REDIRECT
	TNS_TYPE_DATA
	TNS_TYPE_NULL
	TNS_TYPE_UNKNOWN1
	TNS_TYPE_ABORT
	TNS_TYPE_UNKNOWN2
	TNS_TYPE_RESEND
	TNS_TYPE_MARKER
	TNS_TYPE_UNKNOWN3
	TNS_TYPE_UNKNOWN4 = 14
)

const (
	TNS313_DATA_ID_SNS PktType = 0xde // secure network service
	TNS313_DATA_ID_SP  PktType = 0x01 // set protocol
	TNS313_DATA_ID_OCI PktType = 0x03 // OCI function
	TNS313_DATA_ID_RS  PktType = 0x04 // return status
)

type DataIDVer314 byte

const (
	TNS314_DATA_ID_SNS DataIDVer314 = 0xde // secure network service
	TNS314_DATA_ID_SP  DataIDVer314 = 0x01 // set protocol
	TNS314_DATA_ID_OCI DataIDVer314 = 0x03 // OCI function
	TNS314_DATA_ID_RS  DataIDVer314 = 0x04 // return status
)

const (
	TNS_313 uint16 = 313
	TNS_314 uint16 = 314
)

type ErrorCode byte

const (
	CODE_SUCCESS  ErrorCode = 0x00
	CODE_PKT_LEN  ErrorCode = 0x01
	CODE_PKT_TYPE ErrorCode = 0x02
	CODE_DATA_ID  ErrorCode = 0x03
)

type PktHandle struct {
	Version uint16
	Type    PktType
	Func    func(data []byte, sess *tnsSession) error
}

type tnsSession struct {
	Version uint16
}

var msgTypeHandles = []PktHandle{
	{Type: TNS_TYPE_CONNECT, Func: func(pkt []byte, sess *tnsSession) error {
		sess.Version = CONNECT_VER(pkt)
		log.Printf("sess.Version: %d", sess.Version)
		return nil
	}},
	{Type: TNS_TYPE_DATA, Func: func(pkt []byte, sess *tnsSession) error {
		dataID := DATA_ID(pkt)
		pktData := PKT_DATA(pkt)

		for _, handle := range msgDataIdHandles {
			if handle.Version == sess.Version && handle.Type == PktType(dataID) && handle.Func != nil {
				_ = handle.Func(pktData, sess)
			}
		}
		return nil
	}},
}

var msgDataIdHandles = []PktHandle{
	{Version: TNS_313, Type: TNS313_DATA_ID_SNS, Func: func(pkt []byte, sess *tnsSession) error {
		clientVer := pkt[8:12]
		log.Printf("    Client Version: %d.%d.%d.%d\n", clientVer[0], clientVer[1], clientVer[2], clientVer[3])
		return nil
	}},
}

// Handler gets incoming requests
func (o *oracleInterop) Handler(r io.Reader, source string) error {
	if source != ClientSide {
		_, err := io.Copy(io.Discard, r)
		return err
	}

	sess := &tnsSession{}

	for {
		pkt, err := o.readPacket(r)
		if err != nil {
			return err
		}

		pktType := PKT_TYPE(pkt)

		for _, handle := range msgTypeHandles {
			if handle.Type == pktType && handle.Func != nil {
				handle.Func(pkt, sess)
			}
		}
	}
}

// wrapper around our classic readPacket to handle segmented packets
func (o *oracleInterop) readPacket(c io.Reader) (buf []byte, err error) {
	buf, err = ReadPacket(c)
	if err != nil {
		return
	}
	packetLen := int(PKT_LEN(buf))
	for {
		if len(buf) == packetLen {
			break
		}
		var b []byte
		b, err = ReadPacket(c)
		if err != nil {
			return
		}
		buf = append(buf, b...)
	}
	return
}

const (
	chunkSize = 4096
)

// ReadPacket all available data from socket
func ReadPacket(conn io.Reader) ([]byte, error) {
	data := make([]byte, chunkSize)
	buf := bytes.Buffer{}
	for {
		n, err := conn.Read(data)
		if err != nil {
			return nil, err
		}
		buf.Write(data[:n])
		if n != chunkSize {
			break
		}
	}
	return buf.Bytes(), nil
}
