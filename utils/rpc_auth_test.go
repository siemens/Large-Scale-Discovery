/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"bytes"
	"encoding/gob"
	"io"
	"strings"
	"testing"
)

// readerOnly hides any ReadByte method so the wrapped reader is not an io.ByteReader, mirroring a real net.Conn. A
// bare *bytes.Buffer would satisfy io.ByteReader and hide the over-read that gob otherwise causes on real connections.
type readerOnly struct{ r io.Reader }

func (ro readerOnly) Read(p []byte) (int, error) { return ro.r.Read(p) }

// TestConnectionHandshake verifies the connection-auth handshake over a non-ByteReader source (like a real net.Conn):
// the secret a client gob-encodes is read back identically, and decoding through the unbuffered byteReader consumes
// exactly the handshake, leaving the following bytes (the net/rpc stream on a real connection) untouched. Decoding
// without byteReader would let gob's internal bufio.Reader read ahead and swallow those trailing bytes.
func TestConnectionHandshake(t *testing.T) {
	cases := []string{"", "dev_secret", strings.Repeat("x", 64), strings.Repeat("y", 10000)}
	for _, secret := range cases {
		var buf bytes.Buffer

		// Client side: encode the secret, then append bytes simulating the subsequent net/rpc stream
		if err := gob.NewEncoder(&buf).Encode(secret); err != nil {
			t.Fatalf("encode handshake (len=%d) failed: %v", len(secret), err)
		}
		trailing := "RPC-STREAM-FOLLOWS"
		buf.WriteString(trailing)

		// Server side: decode through the unbuffered byteReader over a non-ByteReader source, exactly as ServeRpc does
		var got string
		if err := gob.NewDecoder(byteReader{readerOnly{&buf}}).Decode(&got); err != nil {
			t.Fatalf("decode handshake (len=%d) failed: %v", len(secret), err)
		}
		if got != secret {
			t.Fatalf("handshake mismatch: got len %d, want len %d", len(got), len(secret))
		}

		// The trailing net/rpc stream must remain intact (no over-read)
		if rest := buf.String(); rest != trailing {
			t.Fatalf("decoder over-read: trailing = %q, want %q", rest, trailing)
		}
	}
}
