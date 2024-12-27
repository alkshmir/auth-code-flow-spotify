package main

import (
	"math/rand"
	"time"
)

type RandSource interface {
	Intn(n int) int
}

type DefaultRandSource struct{}

func (d DefaultRandSource) Intn(n int) int {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n)
}
