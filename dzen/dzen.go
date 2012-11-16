package dzen
import(
	"github.com/tshadwell/senbar/textwidth"
	"strings"
	"strconv"
	"math"
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
func hasSwitches(t string) int{
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
func Pie(radius uint8, perc float64, colourA, colourB string) string{
	radiusString:=strconv.Itoa(int(radius))
	degreeString:=strconv.Itoa(int(math.Floor(360*perc)))
	return "^ib(1)^fg(" + colourA + ")^c(" + radiusString + ")^p(-" + radiusString + ")^fg(" + colourB + ")^c(" +radiusString + "-" + degreeString + ")^fg()"
}
func AlignRight(text string, iconWidth int, xFontName string) string{
	/*
	 * This function removes all special stuff from text,
	 * takes into account 'sized special stuff' like rectangles,
	 * and icons (if provided), then does calculations to align the text
	 * to the right. Easy!
	 */
	oText := text
	modifier :=0
	caretPos:= hasSwitches(oText)
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
		caretPos=hasSwitches(oText)
	}
	return "^p(_RIGHT)^p(-"+strconv.Itoa(int(textwidth.Get(xFontName, oText))+modifier)+")"+text+"^p()"
}
