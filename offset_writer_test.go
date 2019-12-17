package append

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestOffsetWriter(t *testing.T) {
	dir, err := ioutil.TempDir("", "forward")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fn := path.Join(dir, "ff")
	ow, err := NewOffsetWriter(fn)
	if err != nil {
		t.Fatal(err)
	}
	v, err := ow.ReadOrDefault(2)
	if err != nil {
		t.Fatal(err)
	}

	if v != 2 {
		t.Fatal("expected 2")
	}

	for i := 0; i < 1000; i++ {
		err = ow.SetOffset(int64(i))
		if err != nil {
			t.Fatal(err)
		}

		v, err = ow.ReadOrDefault(2)
		if err != nil {
			t.Fatal(err)
		}
		if v != int64(i) {
			t.Fatalf("got %d expected %d", v, i)
		}
		err = ow.Close()
		if err != nil {
			t.Fatal(err)
		}
		ow, err = NewOffsetWriter(fn)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestFileOffsetCorrupt(t *testing.T) {
	dir, err := ioutil.TempDir("", "forward")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fn := path.Join(dir, "ff")
	err = ioutil.WriteFile(fn, make([]byte, 16), 0600)
	if err != nil {
		t.Fatal(err)
	}

	ow, err := NewOffsetWriter(fn)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ow.ReadOrDefault(2)
	if err == nil {
		t.Fatal("expected error")
	}
}
