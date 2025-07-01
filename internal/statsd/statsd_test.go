package statsd

import (
	"log/slog"
	"net"
	"testing"

	"github.com/matryer/is"
)

var dataOut = make(chan []byte)

func testServer(t *testing.T) error {
	t.Helper()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 65004})
	if err != nil {
		return err
	}

	go func() {
		buf := make([]byte, 0, BUFFER_SIZE)
		oob := make([]byte, 0, BUFFER_SIZE_JUMBO)
		n, _, _, _, err := conn.ReadMsgUDPAddrPort(buf, oob)
		if err != nil {
			slog.Error("err", "er", err)
		}

		if n > int(BUFFER_SIZE) {

		}
	}()

	return err
}

func TestBufferPool(t *testing.T) {
	is := is.New(t)
	is.NoErr(nil)

	testServer(t)
}
