package random

import (
	"math/rand"
	"time"
)

var alpha = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"abcdefghijklmnopqrstuvwxyz")

func NewRandomString(size int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixMicro()))

	res := make([]rune, size)
	for i := range res {
		res[i] = alpha[rnd.Intn(len(alpha))]
	}

	return string(res)
}
