package bytes

import (
	"fmt"
	"strings"
	"testing"
)

func TestContentParser_ModuleName(t *testing.T) {

	moduleName := "github.com/gomods/athens"

	FQN := fmt.Sprintf(`module "%s"`, moduleName)

	parser := NewContentParser([]byte(FQN))

	got, err := parser.ModuleName()
	if err != nil {
		t.Fatalf("Expected to find a module name... Got %v", err)
	}

	if !strings.EqualFold(got, moduleName) {
		t.Fatalf(`Module names do not match.. \n 
Expected %s .. Got %s`, moduleName, got)
	}
}
