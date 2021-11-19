package block_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"testing"

	"github.com/ecadlabs/jtree"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlock(t *testing.T) {
	dir, _ := os.Getwd()
	f := os.DirFS(dir)
	files, err := fs.Glob(f, "testdata/*.json")
	require.NoError(t, err)

	for _, test := range files {
		t.Run(test, func(t *testing.T) {
			fd, err := os.Open(test)
			require.NoError(t, err)
			defer fd.Close()

			dec := jtree.NewDecoder(fd)
			dec.DisallowUnknownFields()

			var b block.Block
			assert.NoError(t, dec.Decode(&b))

			hdr := b.GetHeader()
			buf, _ := json.MarshalIndent(hdr, "", "    ")
			t.Log(string(buf))

			buf, _ = json.MarshalIndent(b.Metadata, "", "    ")
			t.Log(string(buf))

			stat := b.Stat()
			buf, _ = json.MarshalIndent(stat, "", "    ")
			t.Log(string(buf))
		})
	}
}
