package types

import (
	"encoding/hex"
	"fmt"
)

type Address [20]uint8

func (a Address) String() string {
	return hex.EncodeToString(a.ToSlice())
}

func (a Address) ToSlice() []byte {
	return a[:]
}
func AddressFromBytes(b []byte) Address {

	if len(b) != 20 {
		msg := fmt.Sprintf("given bytes with length %d should be 32", len(b))
		panic(msg)
	}
	var value [20]uint8

	for i := 0; i < 20; i++ {
		value[i] = b[i]
	}

	return Address(value)

}
