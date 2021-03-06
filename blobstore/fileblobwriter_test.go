// Copyright 2013 The Cinode Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package blobstore

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"strings"
	"testing"
)

type blobTest struct {
	blobHex string
	key     string
	bid     string
}

type blobFinalizer interface {
	Finalize() (bid string, key string, err error)
}

func hexDump(buff []byte) string {
	if len(buff) > 10 {
		return hex.EncodeToString(buff[:10]) + "..."
	}
	return hex.EncodeToString(buff)
}

func blobValidation(t *testing.T, test blobTest, result blobFinalizer, m BlobStorage) {

	key := strings.Replace(test.key, " ", "", -1)
	bid := strings.Replace(test.bid, " ", "", -1)

	rbid, rkey, err := result.Finalize()
	if err != nil {
		t.Error(err)
	}

	if rbid != bid {
		t.Errorf("Invalid blob id generated, got: %v..., expected: %v...", rbid[:16], bid[:16])
	}

	if rkey != key {
		t.Errorf("Invalid key generated, got: %v..., expected: %v...", rkey[:16], key[:16])
	}

	reader, err := m.NewBlobReader(rbid)
	if err != nil {
		t.Errorf("Couldn't open the blob with id: %v... for reading: %v", rbid[:16], err)
	} else {

		readBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Errorf("Couldn't read the blob with id: %v..., error: %v", rbid[:16], err)
		} else {

			blobHex := strings.Replace(test.blobHex, " ", "", -1)
			splitPos := strings.Index(blobHex, "...")
			if splitPos >= 0 {
				dataBefore, _ := hex.DecodeString(blobHex[:splitPos])
				dataAfter, _ := hex.DecodeString(blobHex[splitPos+3:])

				if !bytes.Equal(readBytes[:len(dataBefore)], dataBefore) {
					t.Errorf("The blob with id: %v... has invalid content (starting bytes), got: %v, expected %v",
						rbid[:16],
						hexDump(readBytes),
						hexDump(dataBefore))
				}

				if !bytes.Equal(readBytes[len(readBytes)-len(dataAfter):], dataAfter) {
					t.Errorf("The blob with id: %v... has invalid content (ending bytes), got: %v, expected %v",
						rbid[:16],
						hexDump(readBytes[len(readBytes)-len(dataAfter):]),
						hexDump(dataAfter))
				}

			} else {
				blob, _ := hex.DecodeString(blobHex)
				if !bytes.Equal(readBytes, blob) {
					t.Errorf("The blob with id: %v... has invalid content, got: %v, expected %v",
						rbid[:16],
						hexDump(readBytes),
						hexDump(blob))
				}
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////////

var simpleFileTests = []struct {
	content string
	test    blobTest
}{
	{ // Empty file
		"",
		blobTest{
			"01 eb",
			"01 7b54b668 36c1fbdd 13d2441d 9e1434dc 62ca677f b68f5fe6 6a464baa decdbd00",
			"b4f5a7bb 878c0cec 9cb4bd6a e8bb175a 7ea59c1a 048c5ab7 c119990d 0041cb9c fb67c2aa 9e6fada8 11271977 7b4b80ff ada80205 f8ebe698 1c0ade97 ff3df8e5",
		},
	},
	{ // File with single 'a' character
		"a",
		blobTest{
			"01 8f14",
			"01 504ce2f6 de7e3338 9deb73b2 1f765570 ad2b9f2a a8aaec83 28f47b48 bc3e841f",
			"c9d30a99 38ecea16 bed58efe 5ad5b998 927a56da 7c8c36c1 ee13292d ec79aa50 c5613fc9 0d80c37a 77a5a422 691d1967 693a1236 892e228a d95ed6fe 4b505d85",
		},
	},
	{ // Programmer's challenge
		"Hello World!",
		blobTest{
			"01 855e296f 95d1eaf3 feb7d48c e0",
			"01 ac9d2591 34ccef98 7f9f4df3 115b0b7a 24b379cb ebb2aaa9 1ed811c8 cf5e0907",
			"82aeef20 2165cf11 930ea44a 9ad8337a ea355d63 751a7260 552e3e01 4ad6313b ca69c83f a4e35555 31d44a10 25708183 784af0e2 002562b7 260559ce 0e7af262",
		},
	},
	{ // Alphabet
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		blobTest{
			"01 f0ead942 12737b28 60ea35e3 1c7dd176 b5620968 2c3a6792 1d464823 13c245d4 551c765c 3ca851d7 f375911a 66e6b52b 650d51ea c3",
			"01 b11ef5de bd728940 485629e3 42c572bc c5b103d7 b56de27b 07f901b4 abcdb5d4",
			"4cfb056a 184d4377 eff9fc3e 8364906a f4b3b3c9 467c2fb8 245382bd d535ea17 f8a63abc 190a9253 9bd92951 52f112d3 365d4910 737b9f9f 3e0eb2f2 eef40648",
		},
	},
}

func TestSimpleFiles(t *testing.T) {

	for _, test := range simpleFileTests {

		m := NewMemoryBlobStorage()
		bw := FileBlobWriter{Storage: m}
		bw.Write([]byte(test.content))

		blobValidation(t, test.test, &bw, m)
	}
}

func TestSimpleFileBorder(t *testing.T) {
	m := NewMemoryBlobStorage()
	bw := FileBlobWriter{Storage: m}

	b := make([]byte, 1024)
	for i, _ := range b {
		b[i] = 'a'
	}
	for i := 0; i < 16*1024; i++ {
		bw.Write(b)
	}

	blobValidation(
		t,
		blobTest{
			"0155fff9dc9655f537ef8ce5353b0ba71cba4f21e5dceb7088e9652c764b2b5bc80e7c0de9c1fc2e...ed6939a903cb6b3c7e732aa5f819064e0d8daede01af8977c327756464fbcbacdf1ada087116472e",
			"01bf10a3a98a6bf052317e37199dcea98ec846a258ff1023023c30acd86e35e40e",
			"a81ab8676d6fd4a6492dc817de80897e7b504d6bc7743c55cfd44f2863be6bed5c8ef02e7177666547abf9bf4646adf09764477330a968a1255cfb43f7cb4b50",
		},
		&bw,
		m,
	)

}

///////////////////////////////////////////////////////////////////////////////

func TestSplitFiles1(t *testing.T) {

	m := NewMemoryBlobStorage()
	bw := FileBlobWriter{Storage: m}

	// Fill the blob data
	b := make([]byte, 1024)
	for i, _ := range b {
		b[i] = 'a'
	}
	for i := 0; i < 16*1024; i++ {
		bw.Write(b)
	}
	bw.Write(b[:1])

	blobValidation(
		t,
		blobTest{
			"01a19f05435d5b1bc15ca651c37c517ad482efb4cfa76b6e46a3e48367d478049fd3c5550ac160bf...713b00729c26c6cb415226d4024264d5778fe10a9f31549abfc6bffe8fd6be4aa9094a3e4262c053",
			"01bffd8d7830029b88367640a067ce1e0220a929fdd20c0a9157f6e1e094b19ff2",
			"f8615f370c23b1bf7b654ed19aadc5e2011ff98d139cd1a05be588a8f4d03af375f3598a10b138e9106702945c7c1642827fa807d70a44454585ec5251d45b8a",
		},
		&bw,
		m,
	)
}
