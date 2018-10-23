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

package uuid

import (
	guuid "github.com/gofrs/uuid"
)

// NewV1 returns UUID based on current timestamp and MAC address.
var NewV1 = guuid.NewV1

// NewV2 returns DCE Security UUID based on POSIX UID/GID.
var NewV2 = guuid.NewV2

// NewV3 returns UUID based on MD5 hash of namespace UUID and name.
var NewV3 = guuid.NewV3

// NewV4 returns random generated UUID.
var NewV4 = guuid.NewV4

// NewV5 returns UUID based on SHA-1 hash of namespace UUID and name.
var NewV5 = guuid.NewV5

// Generator provides interface for generating UUIDs.
type Generator = guuid.Generator
