// In-memory framing demonstration. No listening socket or public connection.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const maxBody = 1024

type Frame struct {
	Kind    byte
	Payload []byte
}

func writeAll(w io.Writer, p []byte) error {
	for len(p) > 0 {
		n, err := w.Write(p)
		if n < 0 || n > len(p) {
			return fmt.Errorf("invalid writer count")
		}
		p = p[n:]
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}

func writeFrame(w io.Writer, f Frame) error {
	if (f.Kind != 1 && f.Kind != 2) || len(f.Payload) > maxBody-2 {
		return fmt.Errorf("unsupported kind or oversized payload")
	}
	var prefix [4]byte
	binary.BigEndian.PutUint32(prefix[:], uint32(2+len(f.Payload)))
	if err := writeAll(w, prefix[:]); err != nil {
		return err
	}
	if err := writeAll(w, []byte{1, f.Kind}); err != nil {
		return err
	}
	return writeAll(w, f.Payload)
}

func readFrame(r io.Reader) (Frame, error) {
	var prefix [4]byte
	if _, err := io.ReadFull(r, prefix[:]); err != nil {
		return Frame{}, err
	}
	n := binary.BigEndian.Uint32(prefix[:])
	// Check the untrusted length before allocation.
	if n < 2 || n > maxBody {
		return Frame{}, fmt.Errorf("invalid body length %d", n)
	}
	body := make([]byte, int(n))
	if _, err := io.ReadFull(r, body); err != nil {
		return Frame{}, err
	}
	if body[0] != 1 || (body[1] != 1 && body[1] != 2) {
		return Frame{}, fmt.Errorf("unsupported version or message kind")
	}
	return Frame{Kind: body[1], Payload: body[2:]}, nil
}

// Makes each read arbitrarily short, like a byte-stream transport may do.
type chunkReader struct{ r io.Reader }

func (c chunkReader) Read(p []byte) (int, error) {
	if len(p) > 3 {
		p = p[:3]
	}
	return c.r.Read(p)
}

func main() {
	var wire bytes.Buffer
	for _, f := range []Frame{{1, []byte("block-88")}, {2, []byte("vote-A")}} {
		if err := writeFrame(&wire, f); err != nil {
			panic(err)
		}
	}
	r := chunkReader{bytes.NewReader(wire.Bytes())}
	count := 0
	for {
		f, err := readFrame(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		count++
		fmt.Printf("frame %d: kind=%d payload=%s\n", count, f.Kind, f.Payload)
	}
	fmt.Println("decoded frames:", count)
	var oversized [4]byte
	binary.BigEndian.PutUint32(oversized[:], maxBody+1)
	_, err := readFrame(bytes.NewReader(oversized[:]))
	fmt.Println("oversized frame rejected:", err != nil)
	truncated := []byte{0, 0, 0, 5, 1, 1, 'a'}
	_, err = readFrame(bytes.NewReader(truncated))
	fmt.Println("truncated body rejected:", err == io.ErrUnexpectedEOF)
}
