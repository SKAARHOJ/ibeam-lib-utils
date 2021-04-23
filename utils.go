package ibeam_lib_utils

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func Qint(cond bool, ifTrue int, ifFalse int) int {
	if cond {
		return ifTrue
	} else {
		return ifFalse
	}
}
func Qstr(cond bool, ifTrue string, ifFalse string) string {
	if cond {
		return ifTrue
	} else {
		return ifFalse
	}
}
func Quint32(cond bool, ifTrue uint32, ifFalse uint32) uint32 {
	if cond {
		return ifTrue
	} else {
		return ifFalse
	}
}

func IntExplode(str string, token string) []uint32 {
	outputIntegers := make([]uint32, 0)
	strSplit := strings.Split(str, token)
	for _, val := range strSplit {
		myInt, _ := strconv.Atoi(val)
		outputIntegers = append(outputIntegers, uint32(myInt))
	}

	return outputIntegers
}

func IntImplode(integers []uint32, token string) string {
	outputStrs := make([]string, 0)
	for _, val := range integers {
		outputStrs = append(outputStrs, strconv.Itoa(int(val)))
	}

	return strings.Join(outputStrs, token)
}

func StringImplodeRemoveTrailingEmpty(strings []string, token string) string {
	outputStr := ""
	fill := false
	for i := len(strings) - 1; i >= 0; i-- {
		val := strings[i]

		if len(val) > 0 {
			fill = true
		}
		if fill {
			outputStr = token + val + outputStr
		}
	}

	if len(outputStr) > 0 {
		return outputStr[1:len(outputStr)]
	} else {
		return ""
	}
}

func MapValue(x int, in_min int, in_max int, out_min int, out_max int) int {
	return (x-in_min)*(out_max-out_min)/(in_max-in_min) + out_min
}

func MapIntToFloat(x int, in_min int, in_max int, out_min float64, out_max float64) float64 {
	return float64(x-in_min)*(out_max-out_min)/float64(in_max-in_min) + out_min
}

func MapFloatToInt(x float64, in_min float64, in_max float64, out_min int, out_max int) int {
	return int(math.Round((x-in_min)*float64(out_max-out_min)/(in_max-in_min) + float64(out_min)))
}

func ConstrainValue(input int, minimumValue int, maximumValue int) int {
	if input < minimumValue {
		return minimumValue
	} else if input > maximumValue {
		return maximumValue
	} else {
		return input
	}
}

func ConstrainValueU32(input uint32, minimumValue uint32, maximumValue uint32) uint32 {
	if input < minimumValue {
		return minimumValue
	} else if input > maximumValue {
		return maximumValue
	} else {
		return input
	}
}

func MapAndConstrainValue(x int, in_min int, in_max int, out_min int, out_max int) int {
	return ConstrainValue(MapValue(x, in_min, in_max, out_min, out_max), out_min, out_max)
}

func IsIntIn(needle int, haystack []int) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func Intval(str string) int {
	intval, _ := strconv.Atoi(str)
	return intval
}

func IndexValueToInt(splitTextString []string, index int) int {
	if index < len(splitTextString) {
		value, _ := strconv.Atoi(splitTextString[index])
		return value
	} else {
		return 0
	}
}

func IndexValueToString(splitTextString []string, index int) string {
	if index < len(splitTextString) {
		return splitTextString[index]
	} else {
		return ""
	}
}

func Debug(msg interface{}) {
	jsonRes, _ := json.MarshalIndent(msg, "", "\t")
	jsonStr := string(jsonRes)
	fmt.Println("DEBUG:\n", jsonStr)
}

func StripEmptyJSONObjects(jsonStr *string) {
	re, _ := regexp.Compile(",?\"[^\"]+\":{}")

	for re.MatchString(*jsonStr) {
		*jsonStr = re.ReplaceAllString(*jsonStr, "")
		*jsonStr = strings.ReplaceAll(*jsonStr, "{,", "{")
	}
}

func ReportChangesInState(cur interface{}, prev interface{}, path string, removed bool, incoming chan []byte) {
	v := reflect.ValueOf(cur)
	v2 := reflect.ValueOf(prev)
	typeOfS := v.Type()
	modeStr := Qstr(removed, "REM", "NEW")

	for i := 0; i < v.NumField(); i++ {
		switch typeOfS.Field(i).Type.Kind() {
		case reflect.Struct:
			if prev != nil && v2.NumField() > i {
				ReportChangesInState(v.Field(i).Interface(), v2.Field(i).Interface(), path+"/"+typeOfS.Field(i).Name, removed, incoming)
			} else {
				ReportChangesInState(v.Field(i).Interface(), nil, path+"/"+typeOfS.Field(i).Name, removed, incoming)
			}
		case reflect.Slice:
			rr := reflect.ValueOf(v.Field(i).Interface())
			rr2 := reflect.ValueOf(v2.Field(i).Interface())
			//typeOfSrr := rr.Type()
			for j := 0; j < rr.Len(); j++ {
				var secondVar interface{}
				if rr2.Len() > j {
					secondVar = rr2.Index(j).Interface()
				} else {
					secondVar = nil
				}

				ReportChangesInState(rr.Index(j).Interface(), secondVar, path+"/"+typeOfS.Field(i).Name+"/"+strconv.Itoa(j), removed, incoming)
			}
		case reflect.String:
			if prev != nil && v2.NumField() > i {
				if !removed && v.Field(i).Interface().(string) != v2.Field(i).Interface().(string) {
					incoming <- []byte(fmt.Sprintf("%s %s=%s", "CHG", path+"/"+typeOfS.Field(i).Name, v.Field(i).Interface())) // v.Field(i).Interface()
				}
			} else {
				if removed {
					incoming <- []byte(fmt.Sprintf("%s %s", modeStr, path+"/"+typeOfS.Field(i).Name)) // v.Field(i).Interface()
				} else {
					incoming <- []byte(fmt.Sprintf("%s %s=%s", modeStr, path+"/"+typeOfS.Field(i).Name, v.Field(i).Interface())) // v.Field(i).Interface()
				}
			}
		case reflect.Int:
			if prev != nil && v2.NumField() > i {
				if !removed && v.Field(i).Interface().(int) != v2.Field(i).Interface().(int) {
					incoming <- []byte(fmt.Sprintf("%s %s=%d", "CHG", path+"/"+typeOfS.Field(i).Name, v.Field(i).Interface())) // v.Field(i).Interface()
				}
			} else {
				if removed {
					incoming <- []byte(fmt.Sprintf("%s %s", modeStr, path+"/"+typeOfS.Field(i).Name)) // v.Field(i).Interface()
				} else {
					incoming <- []byte(fmt.Sprintf("%s %s=%d", modeStr, path+"/"+typeOfS.Field(i).Name, v.Field(i).Interface())) // v.Field(i).Interface()
				}
			}
		case reflect.Float32:
			if prev != nil && v2.NumField() > i {
				if !removed && v.Field(i).Interface().(float32) != v2.Field(i).Interface().(float32) {
					incoming <- []byte(fmt.Sprintf("%s %s=%f", "CHG", path+"/"+typeOfS.Field(i).Name, v.Field(i).Interface())) // v.Field(i).Interface()
				}
			} else {
				if removed {
					incoming <- []byte(fmt.Sprintf("%s %s", modeStr, path+"/"+typeOfS.Field(i).Name)) // v.Field(i).Interface()
				} else {
					incoming <- []byte(fmt.Sprintf("%s %s=%f", modeStr, path+"/"+typeOfS.Field(i).Name, v.Field(i).Interface())) // v.Field(i).Interface()
				}
			}
		default:

		}
	}
}
