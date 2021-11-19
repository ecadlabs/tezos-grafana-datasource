package block

import (
	"fmt"
	"math/big"

	"github.com/ecadlabs/jtree"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

type Expr interface {
	Kind() string
}

type Int struct {
	Int big.Int `json:"int,string"`
}

func (*Int) Kind() string { return "int" }

type String struct {
	String string `json:"string"`
}

func (*String) Kind() string { return "string" }

type Bytes struct {
	Bytes model.Bytes `json:"bytes"`
}

func (*Bytes) Kind() string { return "bytes" }

type Sequence []Expr

func (Sequence) Kind() string { return "sequence" }

type Prim struct {
	Prim   string   `json:"prim"`
	Args   Sequence `json:"args,omitempty"`
	Annots []string `json:"annots,omitempty"`
}

func (*Prim) Kind() string { return "prim" }

func exprFunc(n jtree.Node, ctx *jtree.Context) (Expr, error) {
	switch node := n.(type) {
	case jtree.Array:
		var out Sequence
		return out, n.Decode(&out, jtree.OpCtx(ctx))
	case jtree.Object:
		var out Expr
		switch {
		case node.FieldByName("prim") != nil:
			out = new(Prim)
		case node.FieldByName("int") != nil:
			out = new(Int)
		case node.FieldByName("string") != nil:
			out = new(String)
		case node.FieldByName("bytes") != nil:
			out = new(Bytes)
		default:
			return nil, fmt.Errorf("malformed expression %#v", node)
		}
		var tmp interface{} = out
		err := n.Decode(tmp, jtree.OpCtx(ctx))
		return out, err
	default:
		return nil, fmt.Errorf("unexpected node %t", node)
	}
}

func init() {
	jtree.RegisterType(exprFunc)
}
