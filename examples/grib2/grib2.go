//go:generate sh -c "cat grib2.bnf | wiprogen --package=grib2"
//go:generate go fmt
package grib2

type (
	bytes2 = [2]byte
	bytes4 = [4]byte
)
