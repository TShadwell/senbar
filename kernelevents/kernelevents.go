//package kernel events provides a method of listening for kernel level events.
package kernelevents

import "C"

import (
	"encoding/binary"
	"os"
)

type C_timeval struct {
	Tv_sec  C.long
	Tv_usec C.long
}
type Input_event struct {
	Time        C_timeval
	Designation C.ushort
	Code        C.ushort
	Value       C.uint
}

//Get takes a path to a kernel event file and calls a function with the event when one happens.
func Get(path string, process func(Input_event)) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	go (func() {
		for {
			var inp Input_event
			errbin := binary.Read(file, binary.LittleEndian, &inp)
			if errbin != nil {
				panic(errbin)
			}
			process(inp)

		}
	})()
	return nil
}
