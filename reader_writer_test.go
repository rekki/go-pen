package pen

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"sync/atomic"
	"testing"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type Case struct {
	id       uint32
	next     uint32
	document uint32
	data     []byte
}

func TestScan(t *testing.T) {
	dir, err := ioutil.TempDir("", "forward")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fw, err := NewWriter(path.Join(dir, "forward"))
	if err != nil {
		t.Fatal(err)
	}
	reader, err := NewReader(path.Join(dir, "forward"), 1000)
	if err != nil {
		t.Fatal(err)
	}
	cnt := uint64(0)
	cases := []Case{}
	done := make(chan []Case)
	randomized := [][]byte{}
	for i := 0; i < 1000; i++ {
		data := []byte(fmt.Sprintf("%s-%s", "face", RandStringRunes(i)))
		randomized = append(randomized, data)
	}

	for k := 0; k < 100; k++ {
		go func() {
			out := []Case{}
			for i := 0; i < 1000; i++ {
				data := randomized[i]
				atomic.AddUint64(&cnt, 1)
				off, next, err := fw.Append(data)
				if err != nil {
					panic(err)
				}
				if next == off {
					panic("same")
				}
				out = append(out, Case{id: uint32(i), document: off, data: data, next: next})
			}
			done <- out
		}()
	}
	for k := 0; k < 100; k++ {
		x := <-done
		cases = append(cases, x...)
	}

	for _, v := range cases {
		data, next, err := reader.Read(v.document)
		if err != nil {
			t.Fatal(err)
		}

		if next != v.next {
			t.Fatalf("expected %d got %d", v.next, next)
		}
		if !bytes.Equal(v.data, data) {
			t.Fatalf("data mismatch, expected %v got %v", hex.EncodeToString(v.data), hex.EncodeToString(data))
		}
	}
	n := uint64(0)
	err = reader.Scan(0, func(data []byte, offset, next uint32) error {
		n++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := cnt
	if n != expected {
		t.Fatalf("expected %d got %d", expected, n)
	}
	fw.Close()
}

func TestReaderCorrupt(t *testing.T) {
	dir, err := ioutil.TempDir("", "forward")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fn := path.Join(dir, "forward")
	for i := 0; i < 100; i++ {
		for j := 0; j < 32; j++ {

			randomData := make([]byte, rand.Intn(10240))
			rand.Read(randomData)
			validData := make([]byte, 1000+rand.Intn(10240))
			rand.Read(validData)

			file, err := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
			if err != nil {
				t.Fatal(err)
			}

			fw, err := NewWriter(fn)
			if err != nil {
				t.Fatal(err)
			}

			reader, err := NewReader(fn, 1000)
			if err != nil {
				t.Fatal(err)
			}
			off, _, err := fw.Append(validData)
			if err != nil {
				t.Fatal(err)
			}

			data, _, err := reader.Read(off)
			if err != nil {
				t.Fatal("unexpected error")
			}

			if !bytes.Equal(validData, data) {
				t.Fatal("bytes mismatch")
			}
			n := 0
			err = reader.Scan(0, func(data []byte, offset, next uint32) error {
				n++
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if n != 1 {
				t.Fatal("expected 1")
			}

			// corrupt one byte at a time
			_, err = file.WriteAt(randomData, int64(j))
			if err != nil {
				t.Fatal(err)
			}

			_, _, err = reader.Read(off)
			if err == nil {
				t.Fatal("expected error")
			}

			n = 0
			err = reader.Scan(0, func(data []byte, offset, next uint32) error {
				n++
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if n != 0 {
				t.Fatal("expected 0")
			}

			reader.Close()
			fw.Close()
			file.Close()
		}
	}
}

func TestErrorOpen(t *testing.T) {
	dir, err := ioutil.TempDir("", "forward")
	if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(dir)

	_, err = NewWriter(path.Join(dir, "forward"))
	if err == nil {
		t.Fatal("expected error")
	}

	_, err = NewReader(path.Join(dir, "forward"), 1000)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPad(t *testing.T) {
	for i := 1; i <= 64; i++ {
		PAD = uint32(i)
		TestReadWriteBasic(t)
	}
}

func TestParallel(t *testing.T) {
	dir, err := ioutil.TempDir("", "forward")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fw, err := NewWriter(path.Join(dir, "forward"))
	if err != nil {
		t.Fatal(err)
	}
	reader, err := NewReader(path.Join(dir, "forward"), 1000)
	if err != nil {
		t.Fatal(err)
	}
	cnt := uint64(0)
	cases := []Case{}
	done := make(chan []Case)
	randomized := [][]byte{}
	for i := 0; i < 1000; i++ {
		data := []byte(fmt.Sprintf("%s-%s", "face", RandStringRunes(i)))
		randomized = append(randomized, data)
	}

	for k := 0; k < 100; k++ {
		go func() {
			out := []Case{}
			for i := 0; i < 1000; i++ {
				data := randomized[i]
				atomic.AddUint64(&cnt, 1)
				off, _, err := fw.Append(data)
				if err != nil {
					panic(err)
				}
				out = append(out, Case{id: uint32(i), document: off, data: data})
			}
			done <- out
		}()
	}
	for k := 0; k < 100; k++ {
		x := <-done
		cases = append(cases, x...)
	}

	for _, v := range cases {
		data, _, err := reader.Read(v.document)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(v.data, data) {
			t.Fatalf("data mismatch, expected %v got %v", hex.EncodeToString(v.data), hex.EncodeToString(data))
		}
	}
	err = fw.Sync()
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadWriteBasic(t *testing.T) {
	for i := 16; i < 40960; i += 10000 {
		dir, err := ioutil.TempDir("", "forward")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dir)
		fw, err := NewWriter(path.Join(dir, "forward"))
		if err != nil {
			t.Fatal(err)
		}
		defer fw.Close()
		reader, err := NewReader(path.Join(dir, "forward"), i)
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		cnt := uint64(0)
		cases := []Case{}
		for i := 0; i < 1000; i++ {
			data := []byte(RandStringRunes(i))
			atomic.AddUint64(&cnt, 1)
			off, _, err := fw.Append(data)
			if err != nil {
				t.Fatal(err)
			}
			cases = append(cases, Case{id: uint32(i), document: off, data: data})
		}

		for _, v := range cases {
			data, _, err := reader.Read(v.document)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(v.data, data) {
				t.Fatalf("data mismatch, expected %v got %v", v.data, data)
			}
		}
	}
}

func TestHelloWorld(t *testing.T) {
	// used by the docs
	dir, err := ioutil.TempDir("", "forward")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	filename := path.Join(dir, "f")

	w, err := NewWriter(filename)
	if err != nil {
		panic(err)
	}

	id, _, err := w.Append([]byte("hello world"))
	if err != nil {
		panic(err)
	}

	r, err := NewReader(filename, 16)
	if err != nil {
		panic(err)
	}
	data, _, err := r.Read(id)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(data, []byte("hello world")) {
		panic("mismatch")
	}
}

func TestOverwrite(t *testing.T) {
	// used by the docs
	dir, err := ioutil.TempDir("", "forward")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	filename := path.Join(dir, "f")

	w, err := NewWriter(filename)
	if err != nil {
		panic(err)
	}

	r, err := NewReader(filename, 0)
	if err != nil {
		panic(err)
	}

	for i := 1; i < 1000; i++ {
		v := RandStringRunes(i)
		id, _, err := w.Append([]byte(v))
		if err != nil {
			panic(err)
		}

		data, _, err := r.Read(id)
		if err != nil {
			panic(err)
		}

		if !bytes.Equal(data, []byte(v)) {
			panic("mismatch")
		}

		err = w.Overwrite(id, []byte(v+"a"))
		if err != EOVERFLOW {
			panic("expected EOVERFLOW")
		}

		err = w.Overwrite(id, []byte(v[:len(v)-1]))
		if err != nil {
			panic(err)
		}

		data, _, err = r.Read(id)
		if err != nil {
			panic(err)
		}

		if !bytes.Equal(data, []byte(v[:len(v)-1])) {
			panic("mismatch")
		}
	}
}

func TestReaderBadArgs(t *testing.T) {
	filename := "/tmp/non_existing_file_test"
	_, err := NewReader(filename, 15)
	if err != EINVAL {
		panic(err)
	}

	_, err = NewReader(filename, 16)
	if !os.IsNotExist(err) {
		panic(err)
	}

	_, err = NewReader(filename, 0)
	if !os.IsNotExist(err) {
		panic(err)
	}
}
