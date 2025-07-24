package uast

import (
	"encoding/json"
	"io"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

func Query(input io.Reader, q string) ([]*node.Node, error) {
	var n node.Node

	dec := json.NewDecoder(input)

	if err := dec.Decode(&n); err != nil {
		return nil, err
	}

	return n.FindDSL(q)
}
