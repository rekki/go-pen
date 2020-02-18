package pen

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestMonotonic(t *testing.T) {
	dir, err := ioutil.TempDir("", "forwardzz")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	m, err := NewMonotonic(path.Join(dir, "a"))
	if err != nil {
		panic(err)
	}
	if m.MustCount() != 0 {
		t.Fatalf("expected 0, got: %v", m.MustCount())
	}

	for i := uint64(0); i < 1000; i++ {
		data := []byte(RandStringRunes(int(i)))
		if m.MustCount() != i {
			t.Fatalf("expected %d got %d", i, m.MustCount())
		}
		id := m.MustAppend(data)
		if id != i {
			t.Fatalf("expected %d got %d", i, id)
		}

		if m.MustCount() != i+1 {
			t.Fatalf("%d: expected %d got %d", i, i+1, m.MustCount())
		}

		r := m.MustRead(i)
		if !bytes.Equal(r, data) {
			t.Fatalf("bad read i: %d, got %v expected %v", i, r, data)
		}
	}
	err = m.Sync()
	if err != nil {
		t.Fatal(err)
	}

	err = m.Close()
	if err != nil {
		t.Fatal(err)
	}
}
