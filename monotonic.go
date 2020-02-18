package pen

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync/atomic"
)

type Monotonic struct {
	indexFD           *os.File
	dataFD            *os.File
	current           uint64
	currentDataOffset uint64
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
	currentDataOffset, err := dataFD.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, err
	}

	current, err := FixedLen(indexFD, 8)
	if err != nil {
		return nil, err
	}

	return &Monotonic{indexFD: indexFD, dataFD: dataFD, currentDataOffset: uint64(currentDataOffset), current: current}, nil
}

func (m *Monotonic) AppendAt(index uint64, b []byte) error {
	// append in data
	actualSize := len(b) + 16

	currentDataOffset := atomic.AddUint64(&m.currentDataOffset, uint64(actualSize))
	currentDataOffset -= uint64(actualSize)

	err := WriteAtWriter64(m.dataFD, currentDataOffset, b)
	if err != nil {
		return err
	}

	o := make([]byte, 8)
	binary.LittleEndian.PutUint64(o, currentDataOffset)

	err = FixedWriteAt(m.indexFD, index, o)
	if err != nil {
		return err
	}
	return nil
}

func (m *Monotonic) Append(b []byte) (uint64, error) {
	current := atomic.AddUint64(&m.current, 1)
	current--
	err := m.AppendAt(current, b)
	if err != nil {
		return 0, err
	}
	return current, nil
}

func (m *Monotonic) MustAppend(b []byte) uint64 {
	d, err := m.Append(b)
	if err != nil {
		panic(err)
	}
	return d
}

func (m *Monotonic) MustRead(id uint64) []byte {
	d, err := m.Read(id)
	if err != nil {
		panic(err)
	}
	return d
}

func (m *Monotonic) Read(id uint64) ([]byte, error) {
	o := make([]byte, 8)
	err := FixedReadAt(m.indexFD, id, o)
	if err != nil {
		return nil, err
	}
	off := binary.LittleEndian.Uint64(o)
	data, err := ReadFromReader64(m.dataFD, off, 16)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *Monotonic) Count() (uint64, error) {
	return FixedLen(m.indexFD, 8)
}

func (m *Monotonic) MustCount() uint64 {
	d, err := m.Count()
	if err != nil {
		panic(err)
	}
	return d
}

func (m *Monotonic) Sync() error {
	err1 := m.dataFD.Sync()
	err2 := m.indexFD.Sync()

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil

}
func (m *Monotonic) Close() error {
	err1 := m.dataFD.Close()
	err2 := m.indexFD.Close()

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}
