package main

import "testing"

func TestNetwork(t *testing.T) {
	err := checkInternet("http://127.0.0.1")
	if err != nil {
		t.Error(err)
	}
}
