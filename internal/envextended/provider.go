package envextended

import (
	"errors"
	"os"
	"strings"

	"github.com/tidwall/sjson"
)

type Env struct {
	prefix string
	delim  string
	cb     func(key string, value string) (string, interface{})
	out    string
}

func Provider(prefix, delim string, cb func(s string) string) *Env {
	e := &Env{
		prefix: prefix,
		delim:  delim,
		out:    "{}",
	}
	if cb != nil {
		e.cb = func(key string, value string) (string, interface{}) {
			return cb(key), value
		}
	}
	return e
}

// ProviderWithValue works exactly the same as Provider except the callback
// takes a (key, value) with the variable name and value and allows you
// to modify both. This is useful for cases where you may want to return
// other types like a string slice instead of just a string.
func ProviderWithValue(prefix, delim string, cb func(key string, value string) (string, interface{})) *Env {
	return &Env{
		prefix: prefix,
		delim:  delim,
		cb:     cb,
	}
}

// ReadBytes reads the contents of a file on disk and returns the bytes.
func (e *Env) ReadBytes() ([]byte, error) {
	// Collect the environment variable keys.
	var keys []string
	for _, k := range os.Environ() {
		if e.prefix != "" {
			if strings.HasPrefix(k, e.prefix) {
				keys = append(keys, k)
			}
		} else {
			keys = append(keys, k)
		}
	}

	for _, k := range keys {
		parts := strings.SplitN(k, "=", 2)

		var (
			key   string
			value interface{}
		)

		// If there's a transformation callback,
		// run it through every key/value.
		if e.cb != nil {
			key, value = e.cb(parts[0], parts[1])
			// If the callback blanked the key, it should be omitted
			if key == "" {
				continue
			}
		} else {
			key = parts[0]
			value = parts[1]
		}

		if err := e.set(key, value); err != nil {
			return []byte{}, err
		}
	}

	if e.out == "" {
		return []byte("{}"), nil
	}

	return []byte(e.out), nil
}

func (e *Env) set(key string, value interface{}) error {
	out, err := sjson.Set(e.out, strings.Replace(key, e.delim, ".", -1), value)
	if err != nil {
		return err
	}

	e.out = out

	return nil
}

// Read is not supported by the file provider.
func (e *Env) Read() (map[string]interface{}, error) {
	return nil, errors.New("envextended provider does not support this method")
}
