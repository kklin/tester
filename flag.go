package main

import (
	"strings"
)

type stringSet map[string]struct{}

// String is the method to format the flag's value, part of the flag.Value interface.
func (ss stringSet) String() string {
	var keys []string
	for k := range ss {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
func (ss stringSet) Set(value string) error {
	ss[value] = struct{}{}
	return nil
}
