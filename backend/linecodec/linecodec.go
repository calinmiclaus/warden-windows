package linecodec

import (
	"io"
	"net/rpc"
)

type LineDelimitedClientCodec struct {
	rpc.ClientCodec

	destination io.Writer
}

func New(destination io.Writer, codec rpc.ClientCodec) LineDelimitedClientCodec {
	return LineDelimitedClientCodec{
		ClientCodec: codec,

		destination: destination,
	}
}

func (codec LineDelimitedClientCodec) WriteRequest(r *rpc.Request, param interface{}) error {
	err := codec.ClientCodec.WriteRequest(r, param)
	if err != nil {
		return err
	}

	_, err = codec.destination.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}
