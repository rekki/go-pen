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
	reader, err := NewReader(path.Join(dir, "forward"))
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
				off, err := fw.Append(data)
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
	n := uint64(0)
	err = reader.Scan(0, func(offset uint32, data []byte) error {
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
	err = ioutil.WriteFile(fn, make([]byte, 1024), 0600)
	if err != nil {
		t.Fatal(err)
	}

	reader, err := NewReader(fn)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = reader.Read(0)
	if err == nil {
		t.Fatal("expected error")
	}

	// test empty scan of corrupt file
	err = reader.Scan(0, func(offset uint32, data []byte) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	fw, err := NewWriter(fn)
	if err != nil {
		t.Fatal(err)
	}

	dataCase := []byte("abc")
	off, err := fw.Append(dataCase)
	if err != nil {
		t.Fatal(err)
	}

	data, _, err := reader.Read(off)
	if err != nil {
		t.Fatal("expected error")
	}

	if !bytes.Equal(dataCase, data) {
		t.Fatal("bytes mismatch")
	}
	n := 0
	err = reader.Scan(0, func(offset uint32, data []byte) error {
		n++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatal("expected 1")
	}

	reader.Close()
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

	_, err = NewReader(path.Join(dir, "forward"))
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
	reader, err := NewReader(path.Join(dir, "forward"))
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
				off, err := fw.Append(data)
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
	reader, err := NewReader(path.Join(dir, "forward"))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	cnt := uint64(0)
	cases := []Case{}
	for i := 0; i < 1000; i++ {
		data := []byte(RandStringRunes(i))
		atomic.AddUint64(&cnt, 1)
		off, err := fw.Append(data)
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

	id, err := w.Append([]byte("hello world"))
	if err != nil {
		panic(err)
	}

	r, err := NewReader(filename)
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
