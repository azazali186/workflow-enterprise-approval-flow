package websocket

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"
)

// Frame opcodes (RFC 6455 §5.2).
const (
	opContinuation = 0x0
	opText         = 0x1
	opBinary       = 0x2
	opClose        = 0x8
	opPing         = 0x9
	opPong         = 0xA
)

const (
	// maxControlPayload is the largest allowed payload for control frames.
	maxControlPayload = 125
	// pongWait is how long the server waits for a pong before dropping the
	// connection. Must exceed the hub's ping interval (default 30s).
	pongWait = 90 * time.Second
	// writeWait is how long a single frame write may take.
	writeWait = 10 * time.Second
)

// rawConn is the minimal socket surface FrameConn needs. Both Hertz's
// network.Conn (used by the server) and a plain net.Conn (tests) satisfy it.
type rawConn interface {
	io.Reader
	io.Writer
	io.Closer
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

// FrameConn implements the WebSocket wire protocol (RFC 6455) over a raw
// connection: frame encoding/decoding, client masking, ping/pong keepalive,
// and the close handshake. Server→client frames are unmasked per the spec.
//
// Read transparently answers ping frames with pongs, observes pongs as
// liveness proof, and performs the close handshake when a close frame arrives
// (returning io.EOF). It enforces a read deadline of pongWait so a dead peer
// (e.g. behind a load balancer that stops forwarding) is eventually evicted
// instead of holding a connection open forever.
type FrameConn struct {
	conn rawConn

	wmu sync.Mutex
	rmu sync.Mutex

	closed bool
}

// NewFrameConn wraps a raw connection in a WebSocket frame connection.
func NewFrameConn(conn rawConn) *FrameConn {
	return &FrameConn{conn: conn}
}

// Read reads one complete data frame, transparently handling control frames.
func (c *FrameConn) Read(p []byte) (int, error) {
	c.rmu.Lock()
	defer c.rmu.Unlock()

	// A peer that stops responding (no pongs) fails the read deadline.
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return 0, err
	}

	var buf []byte
	for {
		opcode, fin, payload, err := c.readFrame()
		if err != nil {
			return 0, err
		}

		switch opcode {
		case opPing:
			// Reply with a pong carrying the same payload (RFC 6455 §5.5.3).
			if err := c.writeFrame(opPong, payload); err != nil {
				return 0, err
			}
			continue
		case opPong:
			// Liveness proven; nothing to do — reading continues.
			continue
		case opClose:
			// Echo the close frame to complete the handshake, then drop.
			_ = c.writeFrame(opClose, payload)
			return 0, io.EOF
		case opText, opBinary, opContinuation:
			buf = append(buf, payload...)
			if fin {
				n := copy(p, buf)
				return n, nil
			}
			continue
		default:
			return 0, fmt.Errorf("websocket: unsupported opcode %d", opcode)
		}
	}
}

// Write sends a binary data frame (server→client, unmasked).
func (c *FrameConn) Write(p []byte) (int, error) {
	if err := c.writeFrame(opBinary, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Ping sends a keepalive ping frame.
func (c *FrameConn) Ping(ctx context.Context) error {
	return c.writeFrame(opPing, nil)
}

// Pong sends an unsolicited pong frame (accepted by peers, rarely needed).
func (c *FrameConn) Pong(ctx context.Context) error {
	return c.writeFrame(opPong, nil)
}

// Close performs the close handshake (best effort) and closes the connection.
func (c *FrameConn) Close(code int, reason string) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	payload := make([]byte, 2+len(reason))
	binary.BigEndian.PutUint16(payload, uint16(code))
	copy(payload[2:], reason)
	_ = c.writeFrameLocked(opClose, payload)

	return c.conn.Close()
}

// readFrame reads and decodes a single frame, unmasking client payloads.
func (c *FrameConn) readFrame() (opcode byte, fin bool, payload []byte, err error) {
	hdr := make([]byte, 2)
	if _, err = io.ReadFull(c.conn, hdr); err != nil {
		return 0, false, nil, err
	}

	fin = hdr[0]&0x80 != 0
	opcode = hdr[0] & 0x0F
	masked := hdr[1]&0x80 != 0
	length := uint64(hdr[1] & 0x7F)

	switch length {
	case 126:
		ext := make([]byte, 2)
		if _, err = io.ReadFull(c.conn, ext); err != nil {
			return 0, false, nil, err
		}
		length = uint64(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err = io.ReadFull(c.conn, ext); err != nil {
			return 0, false, nil, err
		}
		length = binary.BigEndian.Uint64(ext)
	}

	var maskKey [4]byte
	if masked {
		if _, err = io.ReadFull(c.conn, maskKey[:]); err != nil {
			return 0, false, nil, err
		}
	}

	payload = make([]byte, length)
	if _, err = io.ReadFull(c.conn, payload); err != nil {
		return 0, false, nil, err
	}
	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return opcode, fin, payload, nil
}

// writeFrame serializes and writes a frame under the write lock.
func (c *FrameConn) writeFrame(opcode byte, payload []byte) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	return c.writeFrameLocked(opcode, payload)
}

// writeFrameLocked serializes and writes a frame. Callers must hold c.wmu.
func (c *FrameConn) writeFrameLocked(opcode byte, payload []byte) error {
	if c.closed {
		return io.ErrClosedPipe
	}

	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}

	header := make([]byte, 2, 10)
	header[0] = 0x80 | opcode // FIN + opcode (server frames are unmasked)
	ln := len(payload)
	switch {
	case ln <= maxControlPayload:
		header[1] = byte(ln)
	case ln <= 65535:
		header[1] = 126
		header = append(header, 0, 0)
		binary.BigEndian.PutUint16(header[2:4], uint16(ln))
	default:
		header[1] = 127
		header = append(header, make([]byte, 8)...)
		binary.BigEndian.PutUint64(header[2:10], uint64(ln))
	}

	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	if len(payload) > 0 {
		if _, err := c.conn.Write(payload); err != nil {
			return err
		}
	}
	return nil
}
