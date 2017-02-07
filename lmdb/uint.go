package lmdb

/*
#include <stdlib.h>
#include <stdio.h>
#include <limits.h>
#include "lmdb.h"
*/
import "C"

import "unsafe"

// UintMax is the largest value for a C.uint type and the largest value that
// can be stored in a Uint object.  On 64-bit architectures it is likely that
// UintMax is less than the largest value of Go's uint type.
//
// Applications on 64-bit architectures that would like to store a 64-bit
// unsigned value should use the UintptrData type instead of the UintData
// type.
const UintMax = C.UINT_MAX

const uintSize = unsafe.Sizeof(C.uint(0))
const gouintSize = unsafe.Sizeof(uint(0))

// Uint returns a UintData containing the value C.uint(x).  If the value
// passed to Uint is greater than UintMax a runtime panic will occur.
//
// Applications on 64-bit architectures that want to store a 64-bit unsigned
// value should use Uintptr type instead of Uint.
func Uint(x uint) *UintData {
	return newUintData(x)
}

// GetUint interprets the bytes of b as a C.uint and returns the uint value.
func GetUint(b []byte) (x uint, ok bool) {
	if uintptr(len(b)) != uintSize {
		return 0, false
	}
	x = uint(*(*C.uint)(unsafe.Pointer(&b[0])))
	if uintSize != gouintSize && C.uint(x) != *(*C.uint)(unsafe.Pointer(&b[0])) {
		// overflow
		return 0, false
	}
	return x, true
}

// UintMulti is a FixedPage implementation that stores C.uint-sized data
// values.
type UintMulti struct {
	page []byte
}

var _ FixedPage = (*UintMulti)(nil)

// WrapUintMulti converts a page of contiguous uint value into a UintMulti.
// WrapUintMulti panics if len(page) is not a multiple of
// unsife.Sizeof(C.uint(0)).
//
// See mdb_cursor_get and MDB_GET_MULTIPLE.
func WrapUintMulti(page []byte) *UintMulti {
	if len(page)%int(uintSize) != 0 {
		panic("argument is not a page of uint values")
	}

	return &UintMulti{page}
}

// Len implements FixedPage.
func (m *UintMulti) Len() int {
	return len(m.page) / int(uintSize)
}

// Stride implements FixedPage.
func (m *UintMulti) Stride() int {
	return int(uintSize)
}

// Size implements FixedPage.
func (m *UintMulti) Size() int {
	return len(m.page)
}

// Page implements FixedPage.
func (m *UintMulti) Page() []byte {
	return m.page
}

// Uint returns the uint value at index i.
func (m *UintMulti) Uint(i int) uint {
	data := m.page[i*int(uintSize) : (i+1)*int(uintSize)]
	x := uint(*(*C.uint)(unsafe.Pointer(&data[0])))
	if uintSize > gouintSize && C.uint(x) != *(*C.uint)(unsafe.Pointer(&data[0])) {
		panic("value oveflows uint")
	}
	return x
}

// Append returns the UintMulti result of appending x to m as C.uint data.  If
// the value passed to Append is greater than UintMax a runtime panic will
// occur.
//
// Applications on 64-bit architectures that want to store a 64-bit unsigned
// value should use UintptrMulti type instead of UintMulti.
func (m *UintMulti) Append(x uint) *UintMulti {
	if uintSize < gouintSize && x > UintMax {
		panic("value overflows unsigned int")
	}

	var buf [uintSize]byte
	*(*C.uint)(unsafe.Pointer(&buf[0])) = C.uint(x)
	return &UintMulti{append(m.page, buf[:]...)}
}

// UintData is a Data implementation that contains C.uint-sized data.
type UintData [uintSize]byte

var _ Data = (*UintData)(nil)

func newUintData(x uint) *UintData {
	v := new(UintData)
	v.SetUint(x)
	return v
}

// Uint returns contained data as a uint value.
func (v *UintData) Uint() uint {
	x := *(*C.uint)(unsafe.Pointer(&(*v)[0]))
	if uintSize > gouintSize && C.uint(uint(x)) != x {
		panic("value overflows unsigned int")
	}
	return uint(x)
}

// SetUint stores the value of x as a C.uint in v.  The value of x must not be
// greater than UintMax otherwise a runtime panic will occur.
func (v *UintData) SetUint(x uint) {
	if uintSize < gouintSize && x > UintMax {
		panic("value overflows unsigned int")
	}

	*(*C.uint)(unsafe.Pointer(&(*v)[0])) = C.uint(x)
}

func (v *UintData) tobytes() []byte {
	return (*v)[:]
}

// cuint is a helper type for tests because tests cannot import C
type cuint C.uint

func (x cuint) C() C.uint {
	return C.uint(x)
}
