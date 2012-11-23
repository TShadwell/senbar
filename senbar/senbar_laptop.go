
// +build laptop

package main
import(
	"github.com/TShadwell/senbar/kernelevents"
	"github.com/TShadwell/senbar/kernelevents/event"
	"github.com/TShadwell/senbar/dzen"
	"strconv"
	"os/exec"
	"strings"

)
func volumeIcon(out string) string{
			out += " ^fg(" + SOUND_FG + ")"
			if currentState.mute{
				out += dzen.Icon("spkr_02.xbm");
			}else{
				out += dzen.Icon("spkr_01.xbm")
			}
			out += " " + strconv.Itoa(int(currentState.vol)) + "^fg()"
}
func laptop(){
	voldn:=exec.Command(
		"amixer",
		"-c",
		"0",
		"sset",
		"Master",
		"Playback",
		"1%-")

	volup:=exec.Command(
		"amixer",
		"-c",
		"0",
		"sset",
		"Master",
		"Playback",
		"1%+")


	err := kernelevents.Get("/dev/input/event0", func(thisEvent kernelevents.Input_event){
			flip:= true
			switch thisEvent.Code{
				case	event.KEY_VOLUMEDOWN:
					currentState.vol=uint8(getVolume()-1)
					voldn.Run()
				case	event.KEY_VOLUMEUP:
					currentState.vol=uint8(getVolume()-1)
					volup.Run()
				case	event.KEY_MUTE:
					if thisEvent.Value == 1{
						currentState.mute = !currentState.mute
					}
				default: flip = false
			}
			if flip{
				currentState.redraw()
			}

	})
	if err != nil{
		panic(err)
	}
}
func getVolume() uint8{
	volRaw, _ := shell("amixer", "-c", "0", "get", "Master")
	for _, x := range strings.Split(volRaw, "\n"){
		if x[2:6] == "Mono"{
			out, _ :=strconv.Atoi(strings.SplitN(
				strings.SplitN(x,"%", 2)[0],
				"[",
				2,
			)[1])
			return uint8(out+1)
		}
	}
	panic("Unable to get Volume")
}

