// Package the provides file append with header and checksum
//
// example usage:
//	w, err := NewAppendWriter(filename)
//	if err != nil {
//		panic(err)
//	}
//
//	docID, err := w.Append([]byte("hello world"))
//	if err != nil {
//		panic(err)
//	}
//
//      // ...
//	r, err := NewAppendReader(filename)
//	if err != nil {
//		panic(err)
//	}
//	data, _, err := r.Read(docID)
//	if err != nil {
//		panic(err)
//	}
//      log.Printf("%s",string(data))
package append
