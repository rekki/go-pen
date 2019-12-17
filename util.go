package pen

import "github.com/dgryski/go-metro"

// exposed go-metro Hash but handy to be exported for debug purposes
func Hash(s []byte) uint64 {
	return metro.Hash64(s, 0)
}
