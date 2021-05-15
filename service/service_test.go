package service

import "testing"

func TestCompare(t *testing.T) {
	go DataConsume()
	Compare()
}
