package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type Block struct {
	Height   uint64
	Previous [32]byte
	Data     string
	Hash     [32]byte
}

func digest(b Block) [32]byte {
	var encoded bytes.Buffer
	encoded.WriteString("LEARNING-HASH-CHAIN-v1\x00")
	var n [8]byte
	binary.BigEndian.PutUint64(n[:], b.Height)
	encoded.Write(n[:])
	encoded.Write(b.Previous[:])
	binary.BigEndian.PutUint64(n[:], uint64(len(b.Data)))
	encoded.Write(n[:])
	encoded.WriteString(b.Data)
	return sha256.Sum256(encoded.Bytes())
}

func appendBlock(chain []Block, data string) []Block {
	b := Block{Height: uint64(len(chain)), Data: data}
	if len(chain) > 0 {
		b.Previous = chain[len(chain)-1].Hash
	}
	b.Hash = digest(b)
	return append(chain, b)
}

func verify(chain []Block, trustedGenesis [32]byte) error {
	if len(chain) == 0 {
		return fmt.Errorf("empty chain")
	}
	for i, b := range chain {
		if b.Height != uint64(i) {
			return fmt.Errorf("height mismatch at %d", i)
		}
		if digest(b) != b.Hash {
			return fmt.Errorf("content hash mismatch at %d", i)
		}
		if i == 0 {
			if b.Previous != ([32]byte{}) || b.Hash != trustedGenesis {
				return fmt.Errorf("wrong genesis anchor")
			}
		} else if b.Previous != chain[i-1].Hash {
			return fmt.Errorf("parent link mismatch at %d", i)
		}
	}
	return nil
}

func rewriteFrom(chain []Block, from int) {
	for i := from; i < len(chain); i++ {
		if i > 0 {
			chain[i].Previous = chain[i-1].Hash
		}
		chain[i].Hash = digest(chain[i])
	}
}

func main() {
	chain := appendBlock(nil, "initial rules")
	chain = appendBlock(chain, "Alice pays Bob 10 demo units")
	chain = appendBlock(chain, "Bob pays Carol 3 demo units")
	trustedGenesis := chain[0].Hash
	trustedTip := chain[len(chain)-1].Hash

	fmt.Println("original:", verify(chain, trustedGenesis))
	chain[1].Data = "Alice pays Bob 90 demo units"
	fmt.Println("change data only:", verify(chain, trustedGenesis))
	chain[1].Hash = digest(chain[1])
	fmt.Println("repair one hash:", verify(chain, trustedGenesis))
	rewriteFrom(chain, 1)
	fmt.Println("rewrite all following hashes:", verify(chain, trustedGenesis))
	fmt.Println("still matches earlier trusted tip:", chain[len(chain)-1].Hash == trustedTip)
}
