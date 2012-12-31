/*
Package flagschema allows structs to act as schemas for flag capture.

The tags can have anything up to three colon separated fields.

The first field defines the default value of the flag. If it is not defined, it defaults to
the zero value of struct's field.

The second field defines the usage text of the flag. If it is not defined, it defaults to
nothing.

The last field defines alternate names (shorthands) for the flag. If the last field is specified, the
name of the struct's field is ignored, and the first name is assumed to be the default flag name.
All other names are listed as shorthands.

If '!' is the last character in the tag, the argument is imperitive and ParseArgs() will error if it is not set or is "" or "0". Such a trailing '!' can be escaped with '\'; only trailing '!'s are checked, so 
no escape is neccesary for other exclamation marks.
Imperitive arguments cannot have the value "" or "0".

flagschema accepts all types that package flag accepts; int, int64, uint, uint64,
string, float64 or time.Duration, as well as any struct that impliments flag.Value.
A type that impliments flag.Value that is used should return "", or "0" when
String is called if it has nor been Set or has a non-zero value.

The problems with "" and "0" occur with imperitive values because package flag provides
no way of checking if a value has been set. If the type impliments the Imperitive interface,
we can avoid these problems by callling isSet(), which should return true if Set() has been
called on the object.

*/
package flagschema

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const shorthandString = "Shorthand for "
const imperitiveString = "Required - "

type flagSpec struct {
	names      []string
	usage      string
	typ        reflect.Type
	deflt      string
	imperitive bool
}

type Imperitive interface{
	isSet()bool
	flag.Value
}


func getFlag(field reflect.StructField) (ths flagSpec) {
	ftlen := len(field.Tag)
	if ftlen > 0 && string(field.Tag[ftlen-1]) == "!" {
		if ftlen > 1 && string(field.Tag[ftlen-2]) != "\\" {
			field.Tag = field.Tag[:ftlen-1]
			ths.imperitive = true
		} else if ftlen > 2{
			field.Tag = reflect.StructTag(string(field.Tag[:len(field.Tag)-2]) + "!")
		} else {
			ths.imperitive = true
		}
	}
	parts := strings.Split(string(field.Tag), ";")
	switch useName := true; len(parts) {
	case 3:
		useName = false
		ths.names = strings.Split(parts[2], ",")
		fallthrough
	case 2:
		ths.usage = parts[1]
		fallthrough
	case 1:
		ths.deflt = parts[0]
		fallthrough
	case 0:
		ths.typ = field.Type
		if useName {
			ths.names = append(ths.names, strings.ToLower(field.Name))
		}
	default:
		panic("Too many fields!")
	}
	return
}

type sErr string

func (f sErr) Error() string {
	return "flagschema: " + string(f)
}

func (f sErr) String() string {
	return string(f)
}

type Flags struct {
	*flag.FlagSet
	Error error
	//A slice of names of arguments
	//that must be present.
	Imperitives []string
}

//Function AbortWithString causes os.Exit to be called with signal 1
//and for the program to print the string and PrintDefaults()
func (f Flags) AbortWithString(s string) {
	fmt.Println(s, "\nUsage:")
	f.PrintDefaults()
	os.Exit(1)
}

//Function AbortWithError causes os.Exit to be called with signal 1
//and the program to print the error and PrintDefaults().
func (f Flags) AbortWithError(e error) {
	f.AbortWithString(e.Error())
}

//Function ParseArgs calls FlagSet.Parse(os.Args[1:)
//as well as checking for imperitive values.
func (f Flags) ParseArgs() Flags {
	if f.Error != nil {
		panic(f.Error)
	}
	f.FlagSet.Parse(os.Args[1:])

	if f.Imperitives != nil {
		for _, v := range f.Imperitives {
			g:= f.FlagSet.Lookup(v)
			imp, isImperitive := g.Value.(Imperitive)
			switch hvalue := g.Value.String();{
				case isImperitive:
					if imp.isSet(){
						break
					}
					fallthrough
				case g == nil, hvalue == "0", hvalue == "":
					f.AbortWithString("The argument '" + v + "' must be present and non-zero.")
			}
		}
	}
	return f
}

//Function Set prepares the  flags for capture, and returns a *flag.FlagSet.
func Set(Name string, flags interface{}) (fS Flags) {
	fS.FlagSet = flag.NewFlagSet(Name, flag.ExitOnError)
	v := reflect.ValueOf(flags)
	if v.Kind() != reflect.Ptr || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		fS.Error = sErr("Interface must be non-nil pointer to struct")
		return
	}
	v, t := v.Elem(), v.Elem().Type()

	for i, end := 0, v.NumField(); i < end; i++ {
		thisValue := v.Field(i)
		thisFlag := getFlag(t.Field(i))

		ptr := thisValue.Addr().Interface()

		Tp := thisFlag.typ.Kind()
		if thisFlag.deflt == "" && Tp < reflect.Uintptr && Tp > reflect.Bool {

			thisFlag.deflt = "0"
		}

		//Deal with imperitive arguments
		if thisFlag.imperitive {
			fS.Imperitives = append(fS.Imperitives, thisFlag.names...)
			thisFlag.usage = imperitiveString + thisFlag.usage
		}
		switch thisFlag.typ.Kind() {
		case reflect.Bool:
			pt := ptr.(*bool)
			if thisFlag.deflt == "" {
				thisFlag.deflt = "false"
			}
			defltVal, err := strconv.ParseBool(thisFlag.deflt)
			if err != nil {
				fS.Error = sErr(err.Error())
				return
			}
			for n, name := range thisFlag.names {
				usage := thisFlag.usage
				if n != 0 {
					usage = shorthandString + thisFlag.names[0]
				}
				fS.BoolVar(pt, name, defltVal, usage)
			}
		case reflect.Int:
			pt := ptr.(*int)
			defltVal, err := strconv.ParseInt(thisFlag.deflt, 0, 64)

			if err != nil {
				fS.Error = sErr(err.Error())
				return
			}

			for n, name := range thisFlag.names {
				usage := thisFlag.usage
				if n != 0 {
					usage = shorthandString + thisFlag.names[0]
				}
				fS.IntVar(pt, name, int(defltVal), usage)
			}
		case reflect.Int64:
			pt := ptr.(*int64)
			defltVal, err := strconv.ParseInt(thisFlag.deflt, 0, 64)

			if err != nil {
				fS.Error = sErr(err.Error())
				return
			}

			for n, name := range thisFlag.names {
				usage := thisFlag.usage
				if n != 0 {
					usage = shorthandString + thisFlag.names[0]
				}
				flag.Int64Var(pt, name, defltVal, usage)
			}
		case reflect.Uint:
			pt := ptr.(*uint)
			defltVal, err := strconv.ParseUint(thisFlag.deflt, 0, 64)

			if err != nil {
				fS.Error = sErr(err.Error())
				return
			}

			for n, name := range thisFlag.names {
				usage := thisFlag.usage
				if n != 0 {
					usage = shorthandString + thisFlag.names[0]
				}
				fS.UintVar(pt, name, uint(defltVal), usage)
			}
		case reflect.Uint64:
			pt := ptr.(*uint64)
			defltVal, err := strconv.ParseUint(thisFlag.deflt, 0, 64)

			if err != nil {
				fS.Error = sErr(err.Error())
				return
			}

			for n, name := range thisFlag.names {
				usage := thisFlag.usage
				if n != 0 {
					usage = shorthandString + thisFlag.names[0]
				}
				flag.Uint64Var(pt, name, defltVal, usage)
			}
		case reflect.String:
			pt := ptr.(*string)

			for n, name := range thisFlag.names {
				usage := thisFlag.usage
				if n != 0 {
					usage = shorthandString + thisFlag.names[0]
				}
				fS.StringVar(pt, name, thisFlag.deflt, usage)
			}
		case reflect.Float64:
			pt := ptr.(*float64)
			defltVal, err := strconv.ParseFloat(thisFlag.deflt, 64)

			if err != nil {
				fS.Error = sErr(err.Error())
				return
			}

			for n, name := range thisFlag.names {
				usage := thisFlag.usage
				if n != 0 {
					usage = shorthandString + thisFlag.names[0]
				}
				flag.Float64Var(pt, name, defltVal, usage)
			}
		case reflect.Struct:
			if thisFlag.typ.PkgPath() == "time" && thisFlag.typ.Name() == "Duration" {
				pt := ptr.(*time.Duration)
				defltVal, err := time.ParseDuration(thisFlag.deflt)

				if err != nil {
					fS.Error = sErr(err.Error())
					return
				}
				for n, name := range thisFlag.names {
					usage := thisFlag.usage
					if n != 0 {
						usage = shorthandString + thisFlag.names[0]
					}
					fS.DurationVar(pt, name, defltVal, usage)
				}
			} else if pt, ok := ptr.(flag.Value); ok {
				for n, name := range thisFlag.names {
					usage := thisFlag.usage

					if n != 0 {
						usage = shorthandString + thisFlag.names[0]
					}
					fS.Var(pt, name, usage)
				}
			} else {
				fS.Error = sErr(thisFlag.typ.PkgPath() + "." + thisFlag.typ.Name() + " does not impliment flag.Value, or is not int, int64, uint, uint64, string, float64 or time.Duration.")
				return
			}

		}
	}
	return
}
