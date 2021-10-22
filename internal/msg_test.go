package internal

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsg(t *testing.T) {
	args := COMPOUND4args{
		Tag:          Utf8str_cs("hello"),
		Minorversion: 123,
		Argarray: []Nfs_argop4{
			{
				Argop: OP_PUTFH,
				U:     PUTFH4res{},
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	XdrOut{Out: buffer}.Marshal("", &args)
	byteRes := buffer.Bytes()

	proc := XdrProc_NFSPROC4_COMPOUND{}
	XdrIn{In: bytes.NewReader(byteRes)}.Marshal("", proc.GetArg())
	cmpRes := proc.GetArg().(*COMPOUND4args)
	assert.Equal(t, "hello", string(cmpRes.Tag))
	assert.Equal(t, uint32(123), cmpRes.Minorversion)
	assert.Equal(t, OP_PUTFH, cmpRes.Argarray[0].Argop)
}
