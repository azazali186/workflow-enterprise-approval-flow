package websocket

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTCPPair returns a FrameConn (server side) and the raw client conn.
func newTCPPair(t *testing.T) (*FrameConn, net.Conn) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	serverCh := make(chan net.Conn, 1)
	go func() {
		c, err := ln.Accept()
		if err == nil {
			serverCh <- c
		}
	}()

	client, err := net.Dial("tcp", ln.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	var raw net.Conn
	select {
	case raw = <-serverCh:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for accept")
	}
	t.Cleanup(func() { _ = raw.Close() })

	return NewFrameConn(raw), client
}

// clientFrame builds a masked frame as a browser would send it (RFC 6455 §5.3).
func clientFrame(t *testing.T, opcode byte, payload []byte) []byte {
	t.Helper()

	frame := []byte{0x80 | opcode}
	ln := len(payload)
	switch {
	case ln <= maxControlPayload:
		frame = append(frame, 0x80|byte(ln))
	case ln <= 65535:
		frame = append(frame, 0x80|126, 0, 0)
		binary.BigEndian.PutUint16(frame[len(frame)-2:], uint16(ln))
	default:
		frame = append(frame, 0x80|127)
		frame = append(frame, make([]byte, 8)...)
		binary.BigEndian.PutUint64(frame[len(frame)-8:], uint64(ln))
	}

	mask := []byte{0x12, 0x34, 0x56, 0x78}
	frame = append(frame, mask...)
	for i, b := range payload {
		frame = append(frame, b^mask[i%4])
	}
	return frame
}

// readClientFrame parses an unmasked server frame from the client's perspective.
func readClientFrame(t *testing.T, c net.Conn) (opcode byte, payload []byte) {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))

	hdr := make([]byte, 2)
	_, err := io.ReadFull(c, hdr)
	require.NoError(t, err)

	opcode = hdr[0] & 0x0F
	ln := int(hdr[1] & 0x7F)
	switch ln {
	case 126:
		ext := make([]byte, 2)
		_, err := io.ReadFull(c, ext)
		require.NoError(t, err)
		ln = int(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		_, err := io.ReadFull(c, ext)
		require.NoError(t, err)
		ln = int(binary.BigEndian.Uint64(ext))
	}

	payload = make([]byte, ln)
	_, err = io.ReadFull(c, payload)
	require.NoError(t, err)
	return opcode, payload
}

func TestFrameConnServerWrite(t *testing.T) {
	server, client := newTCPPair(t)

	n, err := server.Write([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)

	opcode, payload := readClientFrame(t, client)
	assert.Equal(t, byte(opBinary), opcode, "server data frames must be binary")
	assert.Equal(t, "hello", string(payload))
}

func TestFrameConnClientRead(t *testing.T) {
	server, client := newTCPPair(t)

	_, err := client.Write(clientFrame(t, opText, []byte("ping!")))
	require.NoError(t, err)

	buf := make([]byte, 128)
	n, err := server.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "ping!", string(buf[:n]), "masked client payload must be unmasked")
}

func TestFrameConnPingPong(t *testing.T) {
	server, client := newTCPPair(t)

	// Server blocks on Read; the ping must be answered transparently.
	readCh := make(chan struct {
		data []byte
		err  error
	}, 1)
	go func() {
		buf := make([]byte, 128)
		n, err := server.Read(buf)
		readCh <- struct {
			data []byte
			err  error
		}{data: append([]byte(nil), buf[:n]...), err: err}
	}()

	// Client sends a ping; the server must reply with a pong carrying the payload.
	_, err := client.Write(clientFrame(t, opPing, []byte("hb")))
	require.NoError(t, err)

	opcode, payload := readClientFrame(t, client)
	assert.Equal(t, byte(opPong), opcode, "ping must be answered with a pong")
	assert.Equal(t, "hb", string(payload), "pong must echo the ping payload")

	// Then a real data frame completes the read.
	_, err = client.Write(clientFrame(t, opBinary, []byte("data")))
	require.NoError(t, err)

	res := <-readCh
	require.NoError(t, res.err)
	assert.Equal(t, "data", string(res.data), "read must resume with the data frame")
}

func TestFrameConnCloseHandshake(t *testing.T) {
	server, client := newTCPPair(t)

	readCh := make(chan error, 1)
	go func() {
		buf := make([]byte, 128)
		_, err := server.Read(buf)
		readCh <- err
	}()

	closePayload := make([]byte, 2)
	binary.BigEndian.PutUint16(closePayload, 1000)
	_, err := client.Write(clientFrame(t, opClose, closePayload))
	require.NoError(t, err)

	// Server echoes the close frame then returns io.EOF.
	opcode, payload := readClientFrame(t, client)
	assert.Equal(t, byte(opClose), opcode, "server must echo the close frame")
	assert.Len(t, payload, 2)

	select {
	case err := <-readCh:
		assert.Equal(t, io.EOF, err)
	case <-time.After(3 * time.Second):
		t.Fatal("read did not return after close handshake")
	}
}

func TestFrameConnLargePayload(t *testing.T) {
	server, client := newTCPPair(t)

	// > 65535 bytes exercises the 64-bit extended length path.
	payload := make([]byte, 70000)
	for i := range payload {
		payload[i] = byte(i)
	}

	_, err := client.Write(clientFrame(t, opBinary, payload))
	require.NoError(t, err)

	buf := make([]byte, 80000)
	n, err := server.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, len(payload), n)
	assert.Equal(t, payload, buf[:n])
}
