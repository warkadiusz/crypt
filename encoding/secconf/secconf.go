package secconf

import (
	"io"
)

type Secconf interface {
	Decode(data []byte, secretKeyring io.Reader) ([]byte, error)
	Encode(data []byte, publicKeyring io.Reader) ([]byte, error)
}
