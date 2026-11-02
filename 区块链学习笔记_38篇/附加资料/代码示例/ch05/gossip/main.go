// Deterministic, finite, in-memory gossip model. It is not GossipSub.
package main

import (
	"crypto/sha256"
	"fmt"
	"sort"
)

const maxTTL = 8

type Message struct {
	ID      [32]byte
	Payload string
	TTL     uint8
}

type Seen struct {
	MaxTTL uint8
}

type Node struct {
	Neighbors []string
	Seen      map[[32]byte]Seen
	Delivered int
}

type Delivery struct {
	From, To string
	Message  Message
}

func messageID(payload string) [32]byte {
	return sha256.Sum256(append([]byte("CH05-GOSSIP/v1/block\x00"), payload...))
}

func receive(n *Node, m Message) (deliver, forward bool, err error) {
	if len(m.Payload) > 1024 || m.TTL > maxTTL || m.ID != messageID(m.Payload) {
		return false, false, fmt.Errorf("invalid message")
	}
	old, exists := n.Seen[m.ID]
	if !exists {
		n.Delivered++
		deliver = true
	} else if m.TTL <= old.MaxTTL {
		return false, false, nil
	}
	// A later arrival with a larger TTL may forward again but never redelivers.
	n.Seen[m.ID] = Seen{MaxTTL: m.TTL}
	return deliver, m.TTL > 0, nil
}

func main() {
	adjacency := map[string][]string{
		"A": {"B", "C"}, "B": {"A", "C", "D"},
		"C": {"A", "B", "E"}, "D": {"B", "E"}, "E": {"C", "D"},
	}
	nodes := make(map[string]*Node)
	var ids []string
	for id, neighbors := range adjacency {
		nodes[id] = &Node{Neighbors: neighbors, Seen: make(map[[32]byte]Seen)}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	msg := Message{Payload: "block-88", TTL: 3}
	msg.ID = messageID(msg.Payload)
	queue := []Delivery{{To: "A", Message: msg}}
	processed := 0
	for len(queue) > 0 {
		d := queue[0]
		queue = queue[1:]
		processed++
		if processed > 1000 {
			panic("simulation event limit")
		}
		n := nodes[d.To]
		_, forward, err := receive(n, d.Message)
		if err != nil {
			panic(err)
		}
		if !forward {
			continue
		}
		next := d.Message
		next.TTL--
		for _, neighbor := range n.Neighbors {
			if neighbor != d.From {
				queue = append(queue, Delivery{d.To, neighbor, next})
			}
		}
	}
	for _, id := range ids {
		fmt.Printf("%s delivered=%d\n", id, nodes[id].Delivered)
	}
	_, forward, err := receive(nodes["A"], msg)
	if err != nil {
		panic(err)
	}
	fmt.Println("repeated message suppressed:", !forward && nodes["A"].Delivered == 1)
	bad := msg
	bad.Payload = "altered-but-id-unchanged"
	_, _, err = receive(nodes["B"], bad)
	fmt.Println("mismatched ID rejected:", err != nil)
	fresh := &Node{Seen: make(map[[32]byte]Seen)}
	zero := msg
	zero.TTL = 0
	delivered, forwarded, err := receive(fresh, zero)
	if err != nil {
		panic(err)
	}
	fmt.Printf("TTL=0: delivered=%t forwarded=%t\n", delivered, forwarded)
}
