package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// Reproduces the packet commitment bytes of ibc-go v8.7.0 (IBC Classic).
// It does not verify Merkle proofs, a light client, or the packet lifecycle.
func packetCommitment(data []byte, timestamp, revision, height uint64) [32]byte {
	b := make([]byte, 24, 56)
	binary.BigEndian.PutUint64(b[0:8], timestamp)
	binary.BigEndian.PutUint64(b[8:16], revision)
	binary.BigEndian.PutUint64(b[16:24], height)
	h := sha256.Sum256(data)
	b = append(b, h[:]...)
	return sha256.Sum256(b)
}

func key(port, channel string, sequence uint64) string {
	// Fixed valid identifiers below; production identifiers need ICS-24 validation.
	return fmt.Sprintf("commitments/ports/%s/channels/%s/sequences/%d", port, channel, sequence)
}

func check(name string, ok bool) {
	fmt.Printf("%s: %v\n", name, ok)
	if !ok {
		panic(name)
	}
}

func main() {
	data := []byte(`{"amount":"100","denom":"uatom","receiver":"bob","sender":"alice"}`)
	c := packetCommitment(data, 0, 2, 1000)
	fmt.Printf("commitment: %x\n", c)
	fmt.Println("store key:", key("transfer", "channel-7", 42))
	check("same bytes agree", c == packetCommitment(data, 0, 2, 1000))
	check("different timeout differs", c != packetCommitment(data, 0, 2, 1001))
	changed := []byte(`{"amount":"101","denom":"uatom","receiver":"bob","sender":"alice"}`)
	check("different data differs", c != packetCommitment(changed, 0, 2, 1000))
	// Routing context is authenticated by its store path and channel/client state,
	// not by silently adding fields to the actual ICS-04 commitment function.
	check("sequence changes key", key("transfer", "channel-7", 42) != key("transfer", "channel-7", 43))
	check("channel changes key", key("transfer", "channel-7", 42) != key("transfer", "channel-8", 42))
	ack := []byte(`{"result":"AQ=="}`)
	fmt.Printf("acknowledgement commitment: %x\n", sha256.Sum256(ack))
}
