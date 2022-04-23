package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/bketelsen/crypt/encoding/secconf/age"
	"github.com/bketelsen/crypt/encoding/secconf/openpgp"
	"io/ioutil"
	"log"
	"os"

	"github.com/bketelsen/crypt/backend"
	"github.com/bketelsen/crypt/backend/consul"
	"github.com/bketelsen/crypt/backend/etcd"
	"github.com/bketelsen/crypt/encoding/secconf"
)

func getCmd(flagset *flag.FlagSet) {
	flagset.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s get [args...] key\n", os.Args[0])
		flagset.PrintDefaults()
	}
	flagset.StringVar(&secretKeyring, "secret-keyring", ".secring.gpg", "path to armored secret keyring")
	flagset.Parse(os.Args[2:])
	key := flagset.Arg(0)
	if key == "" {
		flagset.Usage()
		os.Exit(1)
	}
	backendStore, err := getBackendStore(backendName, endpoint)
	if err != nil {
		log.Fatal(err)
	}
	if plaintext {
		value, err := getPlain(key, backendStore)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", value)
		return
	}
	encryptionEngine, err := getEncryptionEngine(encryptionEngine)
	if err != nil {
		log.Fatal(err)
	}
	value, err := getEncrypted(key, secretKeyring, backendStore, encryptionEngine)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", value)
}

func versionCmd(flagset *flag.FlagSet, ver string) {
	flagset.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s version\n", os.Args[0])
		flagset.PrintDefaults()
	}
	fmt.Println(ver)
}

func getEncrypted(key, keyring string, store backend.Store, encryptionEngine secconf.Secconf) ([]byte, error) {
	var value []byte
	kr, err := os.Open(keyring)
	if err != nil {
		return value, err
	}
	defer kr.Close()
	data, err := store.Get(key)
	if err != nil {
		return value, err
	}
	value, err = encryptionEngine.Decode(data, kr)
	if err != nil {
		return value, err
	}
	return value, err
}

func getPlain(key string, store backend.Store) ([]byte, error) {
	var value []byte
	data, err := store.Get(key)
	if err != nil {
		return value, err
	}
	return data, err
}

func listCmd(flagset *flag.FlagSet) {
	flagset.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s list [args...] key\n", os.Args[0])
		flagset.PrintDefaults()
	}
	flagset.StringVar(&secretKeyring, "secret-keyring", ".secring.gpg", "path to armored secret keyring")
	flagset.Parse(os.Args[2:])
	key := flagset.Arg(0)
	if key == "" {
		flagset.Usage()
		os.Exit(1)
	}
	backendStore, err := getBackendStore(backendName, endpoint)
	if err != nil {
		log.Fatal(err)
	}
	if plaintext {
		list, err := listPlain(key, backendStore)
		if err != nil {
			log.Fatal(err)
		}
		for _, kv := range list {
			fmt.Printf("%s: %s", kv.Key, kv.Value)
		}
		return
	}

	encryptionEngine, err := getEncryptionEngine(encryptionEngine)
	if err != nil {
		log.Fatal(err)
	}

	list, err := listEncrypted(key, secretKeyring, backendStore, encryptionEngine)
	if err != nil {
		log.Fatal(err)
	}
	for _, kv := range list {
		fmt.Printf("%s: %s", kv.Key, kv.Value)
	}
}

func listEncrypted(key, keyring string, store backend.Store, encryptionEngine secconf.Secconf) (backend.KVPairs, error) {
	kr, err := os.Open(keyring)
	if err != nil {
		return nil, err
	}
	defer kr.Close()

	data, err := store.List(key)
	if err != nil {
		return nil, err
	}
	for i, kv := range data {
		data[i].Value, err = encryptionEngine.Decode(kv.Value, kr)
		kr.Seek(0, 0)
		if err != nil {
			return nil, err
		}
	}
	return data, err
}

func listPlain(key string, store backend.Store) (backend.KVPairs, error) {
	data, err := store.List(key)
	if err != nil {
		return nil, err
	}
	return data, err
}

func setCmd(flagset *flag.FlagSet) {
	flagset.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s set [args...] key file\n", os.Args[0])
		flagset.PrintDefaults()
	}
	flagset.StringVar(&keyring, "keyring", ".pubring.gpg", "path to armored public keyring")
	flagset.Parse(os.Args[2:])
	key := flagset.Arg(0)
	if key == "" {
		flagset.Usage()
		os.Exit(1)
	}
	data := flagset.Arg(1)
	if data == "" {
		flagset.Usage()
		os.Exit(1)
	}
	backendStore, err := getBackendStore(backendName, endpoint)
	if err != nil {
		log.Fatal(err)
	}
	d, err := ioutil.ReadFile(data)
	if err != nil {
		log.Fatal(err)
	}

	if plaintext {
		err := setPlain(key, backendStore, d)
		if err != nil {
			log.Fatal(err)
			return
		}
		return
	}
	encryptionEngine, err := getEncryptionEngine(encryptionEngine)
	if err != nil {
		log.Fatal(err)
	}
	err = setEncrypted(key, keyring, d, backendStore, encryptionEngine)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func setPlain(key string, store backend.Store, d []byte) error {
	err := store.Set(key, d)
	return err
}

func setEncrypted(key, keyring string, d []byte, store backend.Store, encryptionEngine secconf.Secconf) error {
	kr, err := os.Open(keyring)
	if err != nil {
		return err
	}
	defer kr.Close()
	secureValue, err := encryptionEngine.Encode(d, kr)
	if err != nil {
		return err
	}
	err = store.Set(key, secureValue)
	return err
}

func getBackendStore(provider string, endpoint string) (backend.Store, error) {
	if endpoint == "" {
		switch provider {
		case "consul":
			endpoint = "127.0.0.1:8500"
		case "etcd":
			endpoint = "http://127.0.0.1:4001"
		}
	}
	machines := []string{endpoint}
	switch provider {
	case "etcd":
		return etcd.New(machines)
	case "consul":
		return consul.New(machines)
	default:
		return nil, errors.New("invalid backend " + provider)
	}
}

func getEncryptionEngine(name string) (secconf.Secconf, error) {
	switch name {
	case "pgp":
		return openpgp.Engine{}, nil
	case "age":
		return age.Engine{}, nil
	}

	return nil, errors.New("invalid encryption engine " + name)
}
