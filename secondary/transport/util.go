package transport

import "io"
import "encoding/binary"
import "github.com/couchbase/indexing/secondary/logging"

func Send(conn transporter, buf []byte, flags TransportFlag, payload []byte) (err error) {
	var n int

	// transport framing
	l := pktLenSize + pktFlagSize
	if maxLen := len(buf); l > maxLen {
		logging.Errorf("sending packet length %v > %v\n", l, maxLen)
		err = ErrorPacketOverflow
		return
	}

	a, b := pktLenOffset, pktLenOffset+pktLenSize
	binary.BigEndian.PutUint32(buf[a:b], uint32(len(payload)))
	a, b = pktFlagOffset, pktFlagOffset+pktFlagSize
	binary.BigEndian.PutUint16(buf[a:b], uint16(flags))
	if n, err = conn.Write(buf[:pktDataOffset]); err == nil {
		if n, err = conn.Write(payload); err == nil && n != len(payload) {
			logging.Errorf("transport wrote only %v bytes for payload\n", n)
			err = ErrorPacketWrite
		}
		laddr, raddr := conn.LocalAddr(), conn.RemoteAddr()
		logging.Tracef("wrote %v bytes on connection %v->%v", len(payload), laddr, raddr)

	} else if n != pktDataOffset {
		logging.Errorf("transport wrote only %v bytes for header\n", n)
		err = ErrorPacketWrite
	}
	return
}

func SendResponseEnd(conn transporter) error {
	buf := make([]byte, pktLenSize+pktFlagSize)
	// Special 0 byte payload and flag to indicate end of response
	return Send(conn, buf, 0, nil)
}

func Receive(conn transporter, buf []byte) (flags TransportFlag, payload []byte, err error) {
	// transport de-framing
	bufHeader := safeBufSlice(buf, pktDataOffset)
	if err = fullRead(conn, bufHeader); err != nil {
		if err == io.EOF {
			logging.Tracef("receiving packet: %v\n", err)
		} else {
			logging.Errorf("receiving packet: %v\n", err)
		}
		return
	}
	a, b := pktLenOffset, pktLenOffset+pktLenSize
	pktlen := binary.BigEndian.Uint32(bufHeader[a:b])

	a, b = pktFlagOffset, pktFlagOffset+pktFlagSize
	flags = TransportFlag(binary.BigEndian.Uint16(bufHeader[a:b]))

	bufPkt := safeBufSlice(buf, int(pktlen))
	if err = fullRead(conn, bufPkt); err != nil {
		if err == io.EOF {
			logging.Tracef("receiving packet: %v\n", err)
		} else {
			logging.Errorf("receiving packet: %v\n", err)
		}
		return
	}

	return flags, bufPkt, err
}

func safeBufSlice(b []byte, l int) []byte {
	if cap(b) >= l {
		return b[:l]
	}

	return make([]byte, l)
}
