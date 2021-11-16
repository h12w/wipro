//go:generate sh -c "cat kpax.bnf | wiprogen --package=kpax"
//go:generate go fmt
package kpax

import "errors"

var (
	ErrSizeMismatch = errors.New("size mismatch")
	ErrCRCMismatch  = errors.New("crc mismatch")
)
