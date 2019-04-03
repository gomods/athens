package storage

import (
	"fmt"

	"github.com/technosophos/moniker"
)

func RandomBucketName(prefix string) string {
	// moniker is a cool library to produce mostly unique, human-readable names
	// see https://github.com/technosophos/moniker for more details
	namer := moniker.New()
	return fmt.Sprintf("%s_%s", prefix, namer.NameSep("_"))
}
