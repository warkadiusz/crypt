// Package age implements secconf encoding using age as specified in the following
// format:
//
//   base64(gpg(gzip(data)))
//
// More about age: filippo.io/age
package age

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"io/ioutil"

	"filippo.io/age"
)

type Engine struct{}

// Decode decodes data using the secconf codec using `age`.
func (Engine) Decode(data []byte, secretKeyring io.Reader) ([]byte, error) {
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer(data))
	entityList, err := age.ParseIdentities(secretKeyring)
	if err != nil {
		return nil, err
	}
	md, err := age.Decrypt(decoder, entityList...)
	if err != nil {
		return nil, err
	}
	gzReader, err := gzip.NewReader(md)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()
	plaintextBytes, err := ioutil.ReadAll(gzReader)
	if err != nil {
		return nil, err
	}
	return plaintextBytes, nil
}

// Encode encodes data to a base64 encoded using the secconf codec.
// data is encrypted using `age` with all public keys found in the supplied keyring.
func (Engine) Encode(data []byte, keyring io.Reader) ([]byte, error) {
	entityList, err := age.ParseRecipients(keyring)
	if err != nil {
		return nil, err
	}
	buffer := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buffer)
	ageWriter, err := age.Encrypt(encoder, entityList...)
	if err != nil {
		return nil, err
	}
	gzWriter := gzip.NewWriter(ageWriter)
	if _, err := gzWriter.Write(data); err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}
	if err := ageWriter.Close(); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
