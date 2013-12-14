// Copyright 2013 Frank Schroeder. All rights reserved. MIT licensed.

package properties

type Properties struct {
	m map[string]string
}

// returns the value for the given key
func (p *Properties) Get(key string) (value string, ok bool) {
	value, ok = p.m[key]
	return value, ok
}

// sets the property key = value and returns the previous value if exists or an empty string
func (p *Properties) Set(key, value string) (prevValue string) {
	prevValue, ok := p.m[key]
	if !ok {
		prevValue = ""
	}

	p.m[key] = value
	return prevValue
}

// returns the number of keys
func (p *Properties) Len() int {
	return len(p.m)
}
