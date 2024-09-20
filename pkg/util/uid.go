package util

import (
	"fmt"
	"hash/fnv"
)

func GetUID(addr string) string {
	h := fnv.New32a()
	h.Write([]byte(addr))
	return fmt.Sprintf("%d", (h.Sum32()))
}
