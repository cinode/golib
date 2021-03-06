package cipherfactory

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestFactoryCreation(t *testing.T) {

	f := Create()

	if f == nil {
		t.Fatal("Could not create default cipher factory")
	}

	if f.GetMinKeySourceBytes() < 16 {
		t.Fatal("Invalid minimal key size, 16 bytes (128 bits) is minimum to successfully protect cinode")
	}

}

func TestFactoryEncryptorCreation(t *testing.T) {

	f := Create()

	buff := &bytes.Buffer{}

	key := make([]byte, f.GetMinKeySourceBytes())
	iv := make([]byte, 0)

	enc, keyStr, err := f.CreateEncryptor(key, iv, buff)

	if err != nil {
		t.Fatalf("Couldn't create encryptor: %v", err)
	}

	if enc == nil {
		t.Fatal("Nil encoder received")
	}

	if keyStr == "" {
		t.Fatal("Nil key returned")
	}
}

func TestFactoryKeyLengthReuqirement(t *testing.T) {

	f := Create()

	for l := range []byte{0, 1, byte(f.GetMinKeySourceBytes() - 1)} {
		_, _, err := f.CreateEncryptor(make([]byte, l), make([]byte, 0), &bytes.Buffer{})
		if err != ErrInsufficientKeySource {
			t.Errorf("Did not get proper error about insufficient key source data size, key length: %v", l)
		}
	}

}

func TestFactoryIVLength(t *testing.T) {

	f := Create()
	key := make([]byte, f.GetMinKeySourceBytes())

	for l := range []byte{0, 1, 3, 8, 16, 32, 64, 28} {
		_, _, err := f.CreateEncryptor(key, make([]byte, l), &bytes.Buffer{})
		if err != nil {
			t.Errorf("Error for IV size: %v", l)
		}
	}

}

func TestFactoryEncryptorCreationFailure(t *testing.T) {

	f := Create()

	buff := &bytes.Buffer{}

	key := make([]byte, f.GetMinKeySourceBytes())
	iv := make([]byte, 0)

	for _, l := range []int{0, 1, f.GetMinKeySourceBytes() - 1} {

		enc, keyStr, err := f.CreateEncryptor(key[:l], iv, buff)

		if err == nil {
			t.Fatal("Did create encryptor with insufficient key size")
		}

		if enc != nil {
			t.Fatal("Got encryptor although error reported")
		}

		if keyStr != "" {
			t.Fatal("Got key although error reported")
		}
	}
}

func TestFactoryEncryptorDecryptorPair(t *testing.T) {

	testSet := [][]byte{
		[]byte{},
		[]byte{47},
		[]byte{13, 17},
		[]byte{54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76},
		make([]byte, 1089),
	}

	// Last entry in the set will be set to random data
	rand.Read(testSet[len(testSet)-1])

	for _, testData := range testSet {

		f := Create()

		buff := &bytes.Buffer{}

		key := make([]byte, f.GetMinKeySourceBytes())
		iv := make([]byte, 0)

		enc, keyStr, err := f.CreateEncryptor(key, iv, buff)

		if err != nil {
			t.Fatalf("Error creating encryptor: %v", err)
		}

		n, err := enc.Write(testData)

		if err != nil {
			t.Fatalf("Error writing to encryptor: %v", err)
		}

		if n != len(testData) {
			t.Fatalf("Not enough data written to the encryptor, requested: %v, got %v", len(testData), n)
		}

		dec, err := f.CreateDecryptor(keyStr, iv, buff)

		if err != nil {
			t.Fatalf("Error creating decryptor: %v", err)
		}

		if dec == nil {
			t.Fatalf("Didn't get decryptor")
		}

		buff2 := &bytes.Buffer{}

		_, err = io.Copy(buff2, dec)
		if err != nil {
			t.Fatalf("Couldn't decode data: %v", err)
		}

		if bytes.Compare(buff2.Bytes(), testData) != 0 {
			t.Fatal("Decryptor returned invalid data")
		}
	}
}

func TestFactoryFailDecryptor(t *testing.T) {

	f := Create()

	for _, ks := range []string{"", "a", "z", "FFABCDABCDABCDABCDABCDABCDABCDABCD", "010000000000000000"} {
		_, err := f.CreateDecryptor(ks, make([]byte, 0), &bytes.Buffer{})
		if err == nil {
			t.Error("Could create decryptor although the key is invalid")
		}
	}

}

func TestFactoryHasher(t *testing.T) {

	f := Create()

	h, err := f.CreateHasher()
	if err != nil {
		t.Fatalf("Couldn't create hasher: %v", err)
	}
	if h == nil {
		t.Fatal("Did not get valid hasher")
	}

	h.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	hash := h.Sum(nil)

	if hash == nil {
		t.Fatalf("Invalid hasher: didn't create hash sum")
	}

	if len(hash) < 16 {
		t.Fatalf("Invalid size of generated hash, at least 16 bytes is required")
	}
}
