package kernelevents
import "C"

import(
	"os"
	"encoding/binary"
)
type C_timeval struct{
	Tv_sec C.long
	Tv_usec C.long
}
type Input_event struct{
	Time	C_timeval
	Designation	C.ushort
	Code	C.ushort
	Value	C.uint
}
//GetKernelEvents returns a channel down which unpacked event data is sent,
// the events are read from the given path (/dev/input/eventX).
func Get(path string, process func(Input_event)) (error){
	file, err := os.Open(path)
	if err != nil{
		return err
	}
	go (func(){for{
		var inp Input_event
		errbin := binary.Read(file, binary.LittleEndian, &inp)
		if errbin != nil{
			panic(errbin)
		}
		process(inp)

	}})()
	return nil
}
