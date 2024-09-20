package main

import (
	"flag"

	"github.com/rkrmr33/1mcb/pkg/server"
)

var cfg server.Config

func main() {
	s, err := server.New(cfg)
	if err != nil {
		panic(err)
	}

	if err := s.Run(); err != nil {
		panic(err)
	}
}

func init() {
	flag.StringVar(&cfg.Addr, "addr", ":7888", "http server bind address")
	flag.Int64Var(&cfg.MaxBodySize, "max-body-size", 1<<32, "max size for the body of an incoming request in bytes")
	flag.Parse()
}
