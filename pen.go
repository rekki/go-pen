// Package the provides file append with header and checksum
//
// example usage:
//	w, err := NewWriter(filename)
//	if err != nil {
//		panic(err)
//	}
//
//	docID, _, err := w.Append([]byte("hello world"))
//	if err != nil {
//		panic(err)
//	}
//
//	// ...
//	r, err := NewReader(filename, 4096)
//	if err != nil {
//		panic(err)
//	}
//	data, err := r.Read(docID)
//	if err != nil {
//		panic(err)
//	}
//	log.Printf("%s",string(data))
package pen
