package pen

import (
	"encoding/binary"
	"fmt"
	"os"
)

type Monotonic struct {
	indexFD *os.File
	dataFD  *os.File
	index   *Writer
	data    *Writer
}

func NewMonotonic(fn string) (*Monotonic, error) {
	dataFD, err := os.OpenFile(fmt.Sprintf("%s.data", fn), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	indexFD, err := os.OpenFile(fmt.Sprintf("%s.index", fn), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return NewMonotonicFromFile(indexFD, dataFD)

}
func NewMonotonicFromFile(indexFD, dataFD *os.File) (*Monotonic, error) {
	index, err := NewWriterFromFile(indexFD)
	if err != nil {
		return nil, err
	}

	data, err := NewWriterFromFile(dataFD)
	if err != nil {
		return nil, err
	}

	return &Monotonic{indexFD: indexFD, dataFD: dataFD, index: index, data: data}, nil
}

func (m *Monotonic) Append(b []byte) (uint32, error) {
	// append in data
	offset, _, err := m.data.Append(b)
	if err != nil {
		return 0, err
	}
	o := make([]byte, 4)
	binary.LittleEndian.PutUint32(o, uint32(offset))

	id, _, err := m.index.Append(o)
	if err != nil {
		return 0, err
	}
	return uint32(id), nil
}

func (m *Monotonic) MustAppend(b []byte) uint32 {
	d, err := m.Append(b)
	if err != nil {
		panic(err)
	}
	return d
}

func (m *Monotonic) MustRead(id uint32) []byte {
	d, err := m.Read(id)
	if err != nil {
		panic(err)
	}
	return d
}

func (m *Monotonic) Read(id uint32) ([]byte, error) {
	// PAD shoud fit 16 + 4
	o, _, err := ReadFromReader(m.indexFD, id, 16+4)
	if err != nil {
		return nil, err
	}
	off := binary.LittleEndian.Uint32(o)
	data, _, err := ReadFromReader(m.dataFD, off, 16)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *Monotonic) Count() (uint32, error) {
	s, err := m.indexFD.Stat()
	if err != nil {
		return 0, err
	}

	return uint32((s.Size() + int64(PAD) - 1) / int64(PAD)), nil
}

func (m *Monotonic) MustCount() uint32 {
	d, err := m.Count()
	if err != nil {
		panic(err)
	}
	return d
}

func (m *Monotonic) Sync() error {
	err1 := m.data.Sync()
	err2 := m.index.Sync()

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil

}
func (m *Monotonic) Close() error {
	err1 := m.data.Close()
	err2 := m.index.Close()

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}
