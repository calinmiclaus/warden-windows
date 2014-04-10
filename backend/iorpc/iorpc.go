package iorpc

import (
	"io"
)

type RPC struct {
	in  io.WriteCloser
	out io.ReadCloser
}

func New(in io.WriteCloser, out io.ReadCloser) RPC {
	return RPC{
		in:  in,
		out: out,
	}
}

func (rpc RPC) Write(data []byte) (int, error) {
	return rpc.in.Write(data)
}

func (rpc RPC) Read(target []byte) (int, error) {
	return rpc.out.Read(target)
}

func (rpc RPC) Close() error {
	if err := rpc.in.Close(); err != nil {
		return err
	}

	if err := rpc.out.Close(); err != nil {
		return err
	}

	return nil
}
