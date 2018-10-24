// Copyright (C) 2013-2018 by Maxim Bublis <b@codemonkey.ru>
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package uuid provides implementation of Universally Unique Identifier (UUID).
// Supported versions are 1, 3, 4 and 5 (as specified in RFC 4122) and
// version 2 (as specified in DCE 1.1).
package uuid

import (
	guuid "github.com/gofrs/uuid"
)

// Size of a UUID in bytes.
const Size = guuid.Size

// UUID representation compliant with specification
// described in RFC 4122.
type UUID = guuid.UUID

// UUID versions
const (
	V1 = guuid.V1
	V2 = guuid.V2
	V3 = guuid.V3
	V4 = guuid.V4
	V5 = guuid.V5
)

// UUID layout variants.
const (
	VariantNCS       = guuid.VariantNCS
	VariantRFC4122   = guuid.VariantRFC4122
	VariantMicrosoft = guuid.VariantMicrosoft
	VariantFuture    = guuid.VariantFuture
)

// UUID DCE domains.
const (
	DomainPerson = guuid.DomainPerson
	DomainGroup  = guuid.DomainGroup
	DomainOrg    = guuid.DomainOrg
)

// String parse helpers.
var (
	urnPrefix  = []byte("urn:uuid:")
	byteGroups = []int{8, 4, 4, 4, 12}
)

// Nil is special form of UUID that is specified to have all
// 128 bits set to zero.
var Nil = guuid.Nil

// Predefined namespace UUIDs.
var (
	NamespaceDNS  = guuid.NamespaceDNS
	NamespaceURL  = guuid.NamespaceURL
	NamespaceOID  = guuid.NamespaceOID
	NamespaceX500 = guuid.NamespaceX500
)

// Must is a helper that wraps a call to a function returning (UUID, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations such as
//	var packageUUID = uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"));
func Must(u UUID, err error) UUID {
	return guuid.Must(u, err)
}
