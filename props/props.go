package props

import (
	"fmt"
	"strings"
	"unicode"
)

type state int

const (
	statePropKey state = iota
	statePropValue
	stateComment
)

type Pair struct{ K, V string }

var truthyValues = []string{"yes", "y", "true"}

func (p Pair) Truthy() bool {
	for _, v := range truthyValues {
		if strings.EqualFold(p.V, v) {
			return true
		}
	}
	return false
}

type Pairs []Pair

// Find returns a Pair with with name, or nil
func (p *Pairs) Find(name string) *Pair {
	for _, p := range *p {
		if p.K == name {
			return &p
		}
	}
	return nil
}

// Fetch attempts to find a pair with a given name, and returns its value and
// true. Otherwise, returns an empty string and false.
func (p *Pairs) Fetch(name string) (string, bool) {
	if pair := p.Find(name); pair != nil {
		return pair.V, true
	} else {
		return "", false
	}
}

// FetchPair attempts to find a pair with a given name, and returns the pair
// representation and true. Returns nil and false otherwise.
func (p *Pairs) FetchPair(name string) (*Pair, bool) {
	if pair := p.Find(name); pair != nil {
		return pair, true
	}
	return nil, false
}

// Map transforms a given Pairs slice into a map
func (p Pairs) Map() map[string]string {
	result := map[string]string{}
	for _, p := range p {
		result[p.K] = p.V
	}
	return result
}

// Merge adds a given Pairs value into the current Pairs, overwriting any
// current value with values from the provided slice.
func (p *Pairs) Merge(in Pairs) {
	for _, pair := range in {
		if idx := p.indexOf(pair.K); idx != -1 {
			(*p)[idx] = pair
		} else {
			*p = append(*p, pair)
		}
	}
}

func (p Pairs) indexOf(key string) int {
	for i, c := range p {
		if c.K == key {
			return i
		}
	}
	return -1
}

// MustGet returns the value of a Pair with a given name, or panics.
func (p *Pairs) MustGet(name string) string {
	pair := p.Find(name)
	if pair == nil {
		panic("Pair with key " + name + " not found")
	}
	return pair.V
}

// ParseProperties takes a list of properties commonly contained within a
// `default.properties` file, and returns a Pairs slice representing them.
func ParseProperties(text string) (Pairs, error) {
	var result []Pair
	s := statePropKey
	var tmpKey strings.Builder
	var tmpValue strings.Builder
	for idx, chr := range []rune(text) {
		switch s {
		case statePropKey:
			if tmpKey.Len() == 0 && chr == ' ' || chr == '\t' || chr == '\r' || chr == '\n' {
				continue
			}
			if tmpKey.Len() == 0 && chr == '#' {
				s = stateComment
				continue
			}
			if tmpKey.Len() == 0 && !unicode.IsLetter(chr) {
				return nil, fmt.Errorf("unexpected char %c at position %d", chr, idx)
			}
			if chr == '=' {
				s = statePropValue
				continue
			}
			tmpKey.WriteRune(chr)
		case stateComment:
			if chr == '\n' {
				s = statePropKey
			}
		case statePropValue:
			if chr == '\n' {
				result = append(result, Pair{K: strings.TrimSpace(tmpKey.String()), V: strings.TrimSpace(tmpValue.String())})
				tmpKey.Reset()
				tmpValue.Reset()
				s = statePropKey
				continue
			}
			tmpValue.WriteRune(chr)
		}
	}

	if s == statePropKey && tmpKey.Len() > 0 {
		return nil, fmt.Errorf("unexpected end of input")
	} else if s == statePropValue {
		result = append(result, Pair{K: strings.TrimSpace(tmpKey.String()), V: strings.TrimSpace(tmpValue.String())})
	}
	return result, nil
}

// FromMap take a map and returns a Pairs based on its contents
func FromMap(in map[string]string) Pairs {
	var pairs Pairs
	for k, v := range in {
		pairs = append(pairs, Pair{K: k, V: v})
	}
	return pairs
}
