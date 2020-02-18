package pen

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestMonotonicTruncate(t *testing.T) {
	dir, err := ioutil.TempDir("", "forwardzz")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	m, err := NewMonotonic(path.Join(dir, "a"))
	if err != nil {
		panic(err)
	}

	for i := uint64(0); i < 100; i++ {
		data := []byte(RandStringRunes(int(i)))
		if m.Count() != i {
			t.Fatalf("expected %d got %d", i, m.Count())
		}

		id := m.MustAppend(data)
		if id != i {
			t.Fatalf("expected %d got %d", i, id)
		}

		if m.Count() != i+1 {
			t.Fatalf("%d: expected %d got %d", i, i+1, m.Count())
		}

		r := m.MustRead(i)
		if !bytes.Equal(r, data) {
			t.Fatalf("bad read i: %d, got %v expected %v", i, r, data)
		}

		r = m.MustLast()
		if !bytes.Equal(r, data) {
			t.Fatalf("bad read i: %d, got %v expected %v", i, r, data)
		}

		for j := uint64(1); j < 10; j++ {
			xid := m.MustAppend(data)
			if xid != i+j {
				t.Fatalf("i: %d, expected %d got %d", i, i+j, xid)
			}
		}

		err = m.TruncateAt(id + 1)
		if err != nil {
			panic(err)
		}

		r = m.MustLast()
		if !bytes.Equal(r, data) {
			t.Fatalf("bad read i: %d, got %v expected %v", i, r, data)
		}

		if m.Count() != i+1 {
			t.Fatalf("%d: expected %d got %d", i, i+1, m.Count())
		}

	}

	err = m.Close()
	if err != nil {
		t.Fatal(err)
	}
}

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
	if m.Count() != 0 {
		t.Fatalf("expected 0, got: %v", m.Count())
	}

	for i := uint64(0); i < 1000; i++ {
		data := []byte(RandStringRunes(int(i)))
		if m.Count() != i {
			t.Fatalf("expected %d got %d", i, m.Count())
		}
		id := m.MustAppend(data)
		if id != i {
			t.Fatalf("expected %d got %d", i, id)
		}

		if m.Count() != i+1 {
			t.Fatalf("%d: expected %d got %d", i, i+1, m.Count())
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
