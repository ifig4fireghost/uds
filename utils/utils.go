package utils

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
)

func Generate(keyword string) string {
	m := md5.New()
	m.Write([]byte(keyword))
	return "/tmp/" + hex.EncodeToString(m.Sum(nil)) + ".socket"
}

func Quit(ret int) {
	os.Exit(ret)
}

func Decode(src string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return nil, err
	}
	b := bytes.NewReader(data)
	var out bytes.Buffer
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	io.Copy(&out, r)
	return out.Bytes(), nil
}

func Encode(src []byte) string {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(src)
	w.Close()

	return base64.StdEncoding.EncodeToString(in.Bytes())
}
