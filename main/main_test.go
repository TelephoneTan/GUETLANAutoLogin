package main

import (
	"net/http"
	_ "net/http/pprof"
	"testing"
)

func Test(t *testing.T) {
	go http.ListenAndServe("0.0.0.0:6061", nil)
	run([]string{"p", "id", "password", "校园网", "0"})
}
