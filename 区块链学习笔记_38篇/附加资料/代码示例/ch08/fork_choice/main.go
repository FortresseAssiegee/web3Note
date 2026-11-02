package main

import (
	"fmt"
	"sort"
)

type Block struct {
	ID, Parent string
	Slot       uint64
}
type Vote struct {
	Block string
	Slot  uint64
}
type Demo struct {
	Blocks map[string]Block
	Weight map[string]uint64
	Latest map[string]Vote
	Anchor string
}

func (d Demo) descendant(child, ancestor string) bool {
	for child != "" {
		if child == ancestor {
			return true
		}
		b, ok := d.Blocks[child]
		if !ok {
			return false
		}
		child = b.Parent
	}
	return false
}

func (d Demo) vote(validator, block string, voteSlot uint64) error {
	if _, ok := d.Weight[validator]; !ok {
		return fmt.Errorf("unknown validator")
	}
	b, ok := d.Blocks[block]
	if !ok || b.Slot > voteSlot || !d.descendant(block, d.Anchor) {
		return fmt.Errorf("invalid target relative to trusted anchor")
	}
	if old, ok := d.Latest[validator]; ok && voteSlot <= old.Slot {
		return fmt.Errorf("stale or equal-slot vote")
	}
	d.Latest[validator] = Vote{block, voteSlot}
	return nil
}

func (d Demo) score(block string) uint64 {
	var weight uint64
	for who, vote := range d.Latest {
		if d.descendant(vote.Block, block) {
			weight += d.Weight[who]
		}
	}
	return weight
}

// Educational heaviest-subtree walk, not Ethereum's full fork-choice algorithm.
// No BLS, epochs, proposer boost, justified filtering, slashing, or networking.
func (d Demo) head() string {
	head := d.Anchor
	for {
		var children []string
		for id, b := range d.Blocks {
			if b.Parent == head {
				children = append(children, id)
			}
		}
		if len(children) == 0 {
			return head
		}
		sort.Strings(children)
		best := children[0]
		for _, candidate := range children[1:] {
			if d.score(candidate) > d.score(best) {
				best = candidate
			}
		}
		head = best
	}
}

func main() {
	d := Demo{
		Blocks: map[string]Block{
			"R": {"R", "", 0}, "A": {"A", "R", 1}, "B": {"B", "R", 1},
			"A2": {"A2", "A", 2}, "B2": {"B2", "B", 3},
		},
		Weight: map[string]uint64{"v1": 32, "v2": 32, "v3": 64},
		Latest: map[string]Vote{}, Anchor: "R",
	}
	for _, v := range []struct {
		who, block string
		slot       uint64
	}{
		{"v1", "A2", 2}, {"v2", "A2", 2}, {"v3", "B2", 3},
	} {
		if err := d.vote(v.who, v.block, v.slot); err != nil {
			panic(err)
		}
	}
	fmt.Printf("A=%d B=%d head=%s (demo tie rule)\n", d.score("A"), d.score("B"), d.head())
	if err := d.vote("v2", "B2", 4); err != nil {
		panic(err)
	}
	fmt.Printf("latest-vote replacement: A=%d B=%d head=%s\n", d.score("A"), d.score("B"), d.head())
	fmt.Println("old vote rejected:", d.vote("v2", "A2", 2) != nil)
	d.Anchor = "B"
	fmt.Println("trusted anchor B -> head:", d.head())
	fmt.Println("conflicting-branch vote rejected:", d.vote("v1", "A2", 5) != nil)
}
