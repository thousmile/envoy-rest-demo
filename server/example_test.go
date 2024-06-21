package main

import (
	"fmt"
	"testing"
	"time"
)

func Test01(t *testing.T) {
	duration, _ := time.ParseDuration("0.1s")
	fmt.Println(duration)
}
