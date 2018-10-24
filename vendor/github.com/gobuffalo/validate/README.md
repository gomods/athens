# github.com/gobuffalo/validate
[![Build Status](https://travis-ci.org/gobuffalo/validate.svg?branch=master)](https://travis-ci.org/gobuffalo/validate) [![GoDoc](https://godoc.org/github.com/gobuffalo/validate?status.svg)](https://godoc.org/github.com/gobuffalo/validate)

This package provides a framework for writing validations for Go applications. It does provide you with few validators, but if you need others you can easly build them.

## Installation

```bash
$ go get github.com/gobuffalo/validate
```

## Usage

Using validate is pretty easy, just define some `Validator` objects and away you go.

Here is a pretty simple example:

```go
package main

import (
	"log"

	v "github.com/gobuffalo/validate"
)

type User struct {
	Name  string
	Email string
}

func (u *User) IsValid(errors *v.Errors) {
	if u.Name == "" {
		errors.Add("name", "Name must not be blank!")
	}
	if u.Email == "" {
		errors.Add("email", "Email must not be blank!")
	}
}

func main() {
	u := User{Name: "", Email: ""}
	errors := v.Validate(&u)
	log.Println(errors.Errors)
  // map[name:[Name must not be blank!] email:[Email must not be blank!]]
}
```

In the previous example I wrote a single `Validator` for the `User` struct. To really get the benefit of using go-validator, as well as the Go language, I would recommend creating distinct validators for each thing you want to validate, that way they can be run concurrently.

```go
package main

import (
	"fmt"
	"log"
	"strings"

	v "github.com/gobuffalo/validate"
)

type User struct {
	Name  string
	Email string
}

type PresenceValidator struct {
	Field string
	Value string
}

func (v *PresenceValidator) IsValid(errors *v.Errors) {
	if v.Value == "" {
		errors.Add(strings.ToLower(v.Field), fmt.Sprintf("%s must not be blank!", v.Field))
	}
}

func main() {
	u := User{Name: "", Email: ""}
	errors := v.Validate(&PresenceValidator{"Email", u.Email}, &PresenceValidator{"Name", u.Name})
	log.Println(errors.Errors)
        // map[name:[Name must not be blank!] email:[Email must not be blank!]]
}
```

That's really it. Pretty simple and straight-forward Just a nice clean framework for writing your own validators. Use in good health.

## Built-in Validators

To make it even simpler, this package has a children package with some nice built-in validators.

```go
package main

import (
	"log"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type User struct {
	Name  string
	Email string
}


func main() {
	u := User{Name: "", Email: ""}
	errors := validate.Validate(
		&validators.EmailIsPresent{Name: "Email", Field: u.Email, Message: "Mail is not in the right format."},
		&validators.StringIsPresent{Field: u.Name, Name: "Name"},
	)
	log.Println(errors.Errors)
	// map[name:[Name can not be blank.] email:[Mail is not in the right format.]]
}
```

All fields are required for each validators, except Message (every validator has a default error message).

### Available Validators

A full list of available validators can be found at [https://godoc.org/github.com/gobuffalo/validate/validators](https://godoc.org/github.com/gobuffalo/validate/validators).
