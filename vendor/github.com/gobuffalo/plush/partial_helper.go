package plush

import (
	"html/template"
	"strings"

	"github.com/pkg/errors"
)

// PartialFeeder is callback function should implemented on application side.
type PartialFeeder func(string) (string, error)

func partialHelper(name string, data map[string]interface{}, help HelperContext) (template.HTML, error) {
	if help.Context == nil || help.Context.data == nil {
		return "", errors.New("invalid context. abort")
	}
	for k, v := range data {
		help.Set(k, v)
	}

	pf, ok := help.Value("partialFeeder").(func(string) (string, error))
	if !ok {
		return "", errors.New("could not found partial feeder from helpers")
	}

	var part string
	var err error
	if part, err = pf(name); err != nil {
		return "", err
	}

	if part, err = Render(part, help.Context); err != nil {
		return "", err
	}

	if layout, ok := data["layout"].(string); ok {
		return partialHelper(
			layout,
			map[string]interface{}{"yield": template.HTML(part)},
			help)
	}

	if ct, ok := help.Value("contentType").(string); ok {
		if strings.Contains(ct, "javascript") && strings.HasSuffix(name, ".html") {
			part = template.JSEscapeString(string(part))
		}
	}
	return template.HTML(part), err
}
