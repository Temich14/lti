// Package sessionproto holds the canonical SessionService protobuf schema (embedded for the dynamic gRPC client).
package sessionproto

import _ "embed"

//go:embed session.proto
var SessionProto []byte

//go:embed google/protobuf/timestamp.proto
var TimestampProto []byte
