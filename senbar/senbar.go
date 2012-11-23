package main
import(
	"github.com/TShadwell/senbar/i3"
	"github.com/TShadwell/senbar/dzen"
	"github.com/TShadwell/senbar/kernelevents"
	"github.com/TShadwell/senbar/kernelevents/event"
	//"github.com/TShadwell/senbar/alsa"

	"io"
	"os/exec"
	"os"
	"time"
	"strconv"
	"fmt"
	"strings"
)
const(
	BARHEIGHT = 12
	BARFONT = "clean"
	BARQUALIFIED_FONTNAME="-*-"+BARFONT+"-*-*-*-*-*-*-*-*-*-*-*-*"
	BARFG = "#efa603"
	BARBG = "#0c0201"
	TIMECOLOUR = "#ffffff"
	SELECTED_RECTANGLE_SIZE=5
	SELECTED_RECTANGLE_COLOUR="#FFFFFF"
	VISIBLE_FG = BARBG
	VISIBLE_BG = BARFG
	SOUND_FG=TIMECOLOUR
	DESKNUM_PADDING=3
)
type i3Bar struct{
	output i3.Output
	process *exec.Cmd
	in io.WriteCloser
}
type i3State struct{
	Outputs []i3.Output
	Workspaces map[string][]i3.Workspace
	Bars []i3Bar
	now time.Time
	vol uint8
	mute bool
}
var currentState i3State
var polling bool
func (bar *i3Bar) spawn(){
	bar.process=exec.Command(
	"dzen2",
	"-x", strconv.Itoa(int(bar.output.Rect.X)),
	"-y", strconv.Itoa(int(bar.output.Rect.Y)),
	"-w", strconv.Itoa(int(bar.output.Rect.Width)),
	"-h", strconv.Itoa(int(BARHEIGHT)),
	"-e","''",
	"-fn",BARFONT,
	"-bg",BARBG,
	"-fg",BARFG,
	"-ta","l",
	"-dock")
		pipe,err:=bar.process.StdinPipe()
	bar.in=pipe
	if err != nil{
		panic(err)
	}
	bar.process.Start()
}
func shell(fun string, arg ...string) (string, error){
	cmd := exec.Command(fun, arg...)
	out, err := cmd.Output()
	return string(out), err
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
func muteSound(){
	exec.Command("amixer", "set", "\"Master\"", "mute").Run()
}
func remove (bar []i3Bar, pos uint){
	bar[pos], bar = bar[len(bar)-1], bar[:len(bar)-1]
}
func ordinal(num uint64) string{
	units:=uint8(string(num)[0])
	if units<4{
		return  []string{
			"th",
			"st",
			"nd",
			"rd",
		}[units]
	}
	return  "th"

}
func ampmHour(hor int)int{
	if hor == 0{
		return 12;
	}
	return hor;
}
func fancyTime(d time.Time) string{
	day :=d.Day()
	var ampm string
	if d.Hour()>12{
		ampm="pm"
	} else{
		ampm="am"
	}
	return fmt.Sprintf("%s, %s %d%s- ^fg("+TIMECOLOUR+")%02d:%02d%s^fg()",
		[]string{
			"Sunday",
			"Monday",
			"Tuesday",
			"Wednesday",
			"Thursday",
			"Friday",
			"Saturday",
		}[day%7],
		[]string{
			"January",
			"February",
			"March",
			"April",
			"May",
			"June",
			"July",
			"August",
			"September",
			"October",
			"November",
			"December",
		}[d.Month()-1],
		day,
		ordinal(uint64(day)),
		ampmHour(d.Hour()%12),
		d.Minute(),
		ampm);
}
func (state *i3State) redraw(){
	toKill :=make([]uint, 0)
	for i, bar := range state.Bars{
		//Check still bound to active output
		workspaces, ok := state.Workspaces[bar.output.Name]
		if ok{
			out :=""
			for _, workspace := range workspaces{
					if workspace.Focused{
					out += "^fg("+VISIBLE_FG+")^bg("+VISIBLE_BG+")"
				}
				numString:=strconv.Itoa(int(workspace.Num))
				out+="^r("+strconv.Itoa(DESKNUM_PADDING+SELECTED_RECTANGLE_SIZE)+"x0)" + numString
				if workspace.Name != numString{
					if workspace.Num == 0{
						out+= workspace.Name
					} else {
						out+=" : " +strings.Trim(workspace.Name, numString)
					}
				}
				out+= "^fg("+SELECTED_RECTANGLE_COLOUR+")^r("+strconv.Itoa(DESKNUM_PADDING)+"x0)^p(_TOP)^p(-2)"
				if workspace.Visible{
					out+="^r"
				} else {
					out+="^ro"
				}
				out+="("+strconv.Itoa(SELECTED_RECTANGLE_SIZE)+"x"+ strconv.Itoa(SELECTED_RECTANGLE_SIZE) + ")^p()^fg()^bg()"
			}

			//Bar icons
			out += " ^fg(" + SOUND_FG + ")"
			if currentState.mute{
				out += dzen.Icon("spkr_02.xbm");
			}else{
				out += dzen.Icon("spkr_01.xbm")
			}
			out += " " + strconv.Itoa(int(currentState.vol)) + "^fg()"


			out+=dzen.AlignRight(fancyTime(currentState.now), -1, BARQUALIFIED_FONTNAME)
			bar.in.Write([]byte(out+"\n"))
		}else{
			toKill=append(toKill, uint(i))
		}
	}
	if len(toKill) >0 {
		for _, n := range toKill{
			remove(state.Bars,n)
		}
	}
}

func makeBars() ([]i3Bar,[]i3.Output){
	outputs:=i3.GetActiveOutputs()
	//Make the slice that will store the bars
	bars := make([]i3Bar, len(outputs))
	//Make a bar for each output
	for i, output :=range outputs{
		bars[i]=i3Bar{}
		bars[i].output=output
		bars[i].spawn()
	}
	return bars, outputs
}
//ignoreall discards all channel inputs until unlocked
func ignoreAll(x chan i3.EventResponse) (restart func()){
	rtn:=make(chan bool)
	go (func(stop chan bool){
		for{
			select{
				case <- stop:
					return
				default:
			}
		}
	})(rtn)
	restart=(func(){
		rtn <-true
	})
	return
}
func main(){
	//Get initial outputs

	//Subscribe to various events
	i3.Subscribe(
		"workspace",
		"output",
	)

	//Set initial state
	bars, outputs := makeBars()
	currentState=i3State{
		outputs,
		i3.WorkspacesPerDisplay(),
		bars,
		time.Now(),
		getVolume(),
		false}

	//Start threads
	go (func(){
		for{
			now:=time.Now();
			currentState.now=now;
			currentState.redraw()
			//Sleep until the next minute.
			time.Sleep(time.Duration(int64(60)-int64(now.Second()))*time.Second)
		}
	})()
	//Process various keypress events
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
			//fmt.Println(thisEvent)
			if flip{
				currentState.redraw()
			}

	})
	if err != nil{
		panic(err)
	}
	go (func(){
		for{
			<-i3.ChWorkspace
			currentState.Workspaces = i3.WorkspacesPerDisplay()
			currentState.redraw()
		}
	})()
	for{
		<-i3.ChOutput
		//Fix the desktop bgs
		exec.Command("nitrogen", "--restore").Start()
		restart:=ignoreAll(i3.ChOutput)
		//Record all bar processes
		newBars, outputs := makeBars()
		oldBars:= make([]*os.Process, len(newBars))
		for i, bar := range currentState.Bars{
			oldBars[i] = bar.process.Process
		}
		//Replace bars with new bars
		currentState.Bars = newBars
		currentState.Outputs = outputs
		//Kill off old bar processes
		for _, proc := range oldBars{
			proc.Kill()
		}
		currentState.redraw()
		restart()

	}
}
