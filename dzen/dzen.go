
//Package dzen provides some functions to help with using dzen in Go.
package dzen
import(
	"github.com/TShadwell/senbar/textwidth"
	"strings"
	"strconv"
	//"fmt"
)
func contains(haystack []int, needle int) bool{
	for _, hay := range haystack{
		if hay == needle{
			return true
		}
	}
	return false
}
//HasSwitches returns the number of dzen switches (^()) not negated (^^) in a given string.
func HasSwitches(t string) int{
	reference := make([]int, 0)
	for i, chr := range t{
		if chr == '^' && !contains(reference, i){
			if t[i+1] != '^'{
				if t[i-1] != '^'{
					return i
				} else {
					reference=append(reference, i+1)
				}
			}
		}
	}
	return -1
}
//Pie returns a string of dzen commands to generate a pie with a sector
//in colourB whose proportion to 360 degrees is the same as perc.
func Pie(radius uint8, perc float64, colourA, colourB string) string{
	radiusString:=strconv.Itoa(int(radius))
	degreeString:=strconv.Itoa(int(math.Floor(360*perc)))
	return "^ib(1)^fg(" + colourA + ")^c(" + radiusString + ")^p(-" + radiusString + ")^fg(" + colourB + ")^c(" +radiusString + "-" + degreeString + ")^fg()"
}
//AlignRight aligns given text to the right hand side, parsing some
//special flags like ^r(), ^ro() (rectangle) and ^c(), ^co() to factor them in.
//The iconWidth parameter can be used to specify the extra width to factor in when
//an icon is encountered.
//This function is still very unintelligent, so don't do anything unexpected!
func AlignRight(text string, iconWidth int, xFontName string) string{
	oText := text
	modifier :=0
	caretPos:= HasSwitches(oText)
	for caretPos != -1{
		x:=oText[caretPos:]
		if iconWidth != -1{
			commandStart:=strings.IndexRune(x, '(')
			command:=x[1:commandStart]
			if command == "i"{
				//If we encounter an icon
				modifier+=iconWidth
			} else if command == "r" || command == "ro"{
				//Special clause if we encounter a rectangle
				// -- we can calculate a rectangle's width.
				commandEnd:=strings.IndexRune(x, ')')
				rectangle:=x[commandStart+1:commandEnd]
				if rectangle!= ""{
					val, err := strconv.Atoi(strings.Split(rectangle, "x")[0])
					if err !=nil{
						panic(err)
					}
					modifier += val
				}
			} else if command == "c" || command =="co"{
				commandEnd := strings.IndexRune(x, ')')
				circle := x[commandStart+1:commandEnd]
				if circle!=""{
					val, err := strconv.Atoi(strings.Split(circle, "-")[0])
					if err !=nil{
						panic(err)
					}
					modifier+=val
				}
			}
		}
		x=x[strings.IndexRune(x,')')+1:]
		oText=oText[:caretPos]+x
		caretPos=HasSwitches(oText)
	}
	return "^p(_RIGHT)^p(-"+strconv.Itoa(int(textwidth.Get(xFontName, oText))+modifier)+")"+text+"^p()"
}
