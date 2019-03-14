default: vet test

vet:
	go vet .

test:
	go test .

doc: README.md

.PHONY: default test vet

README.md: README.md.tpl $(wildcard *.go)
	becca -package $(subst $(GOPATH)/src/,,$(PWD))
