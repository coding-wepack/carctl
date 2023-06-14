package randutil

import (
	"math/rand"
	"time"
	"unsafe"
)

const (
	lowercaseCharWithNumberBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	uppercaseCharWithNumberBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterBytes                  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterNumberBytes            = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandString returns a random string with n characters,
// including lowercase letters and numbers.
func RandString(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(lowercaseCharWithNumberBytes) {
			b[i] = lowercaseCharWithNumberBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
