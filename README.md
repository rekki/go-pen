## github.com/rekki/go-pen: simple file append/read/scan with checksum

[![Build Status](https://travis-ci.org/rekki/go-pen.svg?branch=master)](https://travis-ci.org/rekki/go-pen) [![codecov](https://codecov.io/gh/rekki/go-pen/branch/master/graph/badge.svg)](https://codecov.io/gh/rekki/go-pen) [![GoDoc](https://godoc.org/github.com/rekki/go-pen?status.svg)](https://godoc.org/github.com/rekki/go-pen)

```

format is:
  16 byte header
  XX variable length data

  header:
     4 bytes LE len(data) [1] // LE = Little Endian
     4 bytes LE HASH(data)[2] // go-metro
     4 bytes MAGIC        [3] // 0xbeef
     4 bytes LE HASH(1 2 3)   // hash of the first 12 bytes
  data:
     ..
     ..

```

It has checksum of the data and checksum of the header(also
checksuming the checksum of the data), which makes it quite safe and
robust.


---
# pen
--
    import "github.com/rekki/go-pen"

Package the provides file append with header and checksum

example usage:

    w, err := NewWriter(filename)
    if err != nil {
    	panic(err)
    }

    docID, err := w.Append([]byte("hello world"))
    if err != nil {
    	panic(err)
    }

    // ...
    r, err := NewReader(filename)
    if err != nil {
    	panic(err)
    }
    data, _, err := r.Read(docID)
    if err != nil {
    	panic(err)
    }
    log.Printf("%s",string(data))

## Usage

```go
var EBADSLT = errors.New("checksum mismatch")
```

```go
var MAGIC = []byte{0xb, 0xe, 0xe, 0xf} // change it if you wish

```

```go
var PAD = uint32(64)
```
the offsets are 32 bit, but usually you want to store more than 4gb of data so
we just pad things to minimum 64 byte chunks

#### func  Hash

```go
func Hash(s []byte) uint64
```
exposed go-metro Hash but handy to be exported for debug purposes

#### func  ReadFromReader

```go
func ReadFromReader(reader io.ReaderAt, offset uint32) ([]byte, uint32, error)
```
Reads specific offset. returns data, nextOffset, error. You can
ReadFromReader(nextOffset) if you want to read the next document, or use the
Scan() helper

#### func  ScanFromReader

```go
func ScanFromReader(reader io.ReaderAt, offset uint32, cb func(uint32, []byte) error) error
```
Scan ReaderAt, if the callback returns error this error is returned as the Scan
error

#### type OffsetWriter

```go
type OffsetWriter struct {
}
```


#### func  NewOffsetWriter

```go
func NewOffsetWriter(fn string) (*OffsetWriter, error)
```
Creates new OffsetWriter, used to store one integer in a file with a checksum.
Usually used for when you want to store your position in a log file that you
want to scan from certain offset. example:

    ow, err := NewOffsetWriter(filename)
    if err != nil {
    	panic(err)
    }
    offset, err := ow.ReadOrDefault(0) // whatever the stored value is or 0
    if err == nil {
    	panic(err)
    }

#### func (*OffsetWriter) Close

```go
func (ow *OffsetWriter) Close() error
```

#### func (*OffsetWriter) ReadOrDefault

```go
func (ow *OffsetWriter) ReadOrDefault(def int64) (int64, error)
```
Read the offset or if empty/corrupt/not existant return default value

#### func (*OffsetWriter) SetOffset

```go
func (ow *OffsetWriter) SetOffset(offset int64) error
```
Write the offsed and its checksum format is:

    8 byte LE offset
    8 byte LE checksum // go-metro(offset bytes)

#### type Reader

```go
type Reader struct {
}
```


#### func  NewReader

```go
func NewReader(filename string) (*Reader, error)
```
Create New AppendReader (you just nice wrapper around ReadFromReader adn
ScanFromReader) it is *safe* to use it concurrently Example usage

    r, err := NewReader(filename)
    if err != nil {
    	panic(err)
    }
    // read specific offset
    data, _, err := r.Read(docID)
    if err != nil {
    	panic(err)
    }
    // scan from specific offset
    err = r.Scan(0, func(offset uint32, data []byte) error {
    	log.Printf("%v",data)
    	return nil
    })

#### func (*Reader) Close

```go
func (ar *Reader) Close() error
```

#### func (*Reader) Read

```go
func (ar *Reader) Read(offset uint32) ([]byte, uint32, error)
```
Read at specific offset (just wrapper around ReadFromReader)

#### func (*Reader) Scan

```go
func (ar *Reader) Scan(offset uint32, cb func(uint32, []byte) error) error
```
Scan the open file, if the callback returns error this error is returned as the
Scan error. just a wrapper around ScanFromReader.

#### type Writer

```go
type Writer struct {
}
```


#### func  NewWriter

```go
func NewWriter(filename string) (*Writer, error)
```
Creates new writer and seeks to the end The writer is *safe* to be used
concurrently, because it uses bump pointer like allocation of the offset.
example usage:

    w, err := NewWriter(filename)
    if err != nil {
    	panic(err)
    }

    docID, err := w.Append([]byte("hello world"))
    if err != nil {
    	panic(err)
    }

    r, err := NewReader(filename)
    if err != nil {
    	panic(err)
    }
    data, _, err := r.Read(docID)
    if err != nil {
    	panic(err)
    }
    log.Printf("%s",string(data))

#### func (*Writer) Append

```go
func (fw *Writer) Append(encoded []byte) (uint32, error)
```
Append bytes to the end of file format is:

    16 byte header
    XX variable length data

    header:
       4 bytes LE len(data) [1] // LE = Little Endian
       4 bytes LE HASH(data)[2] // go-metro
       4 bytes MAGIC        [3] // 0xbeef
       4 bytes LE HASH(1 2 3)   // hash of the first 12 bytes
    data:
       ..
       ..

Then the blob(header + data) is padded to PAD size using ((uint32(blobSize) +
PAD - 1) / PAD).

it returns the addressable offset that you can use ReadFromReader() on

#### func (*Writer) Close

```go
func (fw *Writer) Close() error
```

#### func (*Writer) Sync

```go
func (fw *Writer) Sync() error
```
