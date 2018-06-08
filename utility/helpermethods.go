package utility

import (
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"

	overflow "github.com/johncgriffin/overflow"
)

type MdsdTime struct {
	Seconds      int64
	MicroSeconds int64
}

func GetPropsStr(props map[string]interface{}) string {
	var propsStr string
	for key, value := range props {
		valueStr := ""
		if reflect.TypeOf(value).String() == "float64" {
			valueStr = strconv.FormatFloat(value.(float64), 'f', 5, 64)
		} else {
			valueStr = value.(string)
		}
		propsStr = propsStr + key + " : " + valueStr + " , "
	}
	return propsStr
}

//RETURNS: resourceId encoded by converting any letter other than alphanumerics to unicode as per UTF-16
func EncodeSpecialCharacterToUTF16(resourceId string) string {
	encodedStr := ""
	hex := ""
	var replacer = strings.NewReplacer("[", ":", "]", "")
	for _, c := range resourceId {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			hex = fmt.Sprintf("%04X", utf16.Encode([]rune(string(c))))
			encodedStr = encodedStr + replacer.Replace(hex)
		} else {
			encodedStr = encodedStr + string(c)
		}
	}
	return encodedStr
}

//RETURNS: difference of max value that can be held by time and number of 100 ns in current time.
func GetUTCTicks_DescendingOrder(lastSampleTimestamp string) (uint64, error) {

	currentTime, err := time.Parse(LAYOUT, lastSampleTimestamp)
	if err != nil {
		log.Println("E! ERROR while parsing timestamp " + lastSampleTimestamp + "in the layout " + LAYOUT)
		log.Print(err.Error())
		return 0, err
	}
	//maxValureDateTime := time.Date(9999, time.December, 31, 12, 59, 59, 59, time.UTC)
	//Ticks is the number of 100 nanoseconds from zero value of date
	//this value is copied from mdsd code.
	maxValueDateTimeInTicks := uint64(3155378975999999999)

	//zero time is taken to be 1 Jan,1970 instead of 1 Jan, 1 to avoid integer overflow.
	//The Sub() returns int64 and hence it can hold ony nanoseconds corresponding to 290yrs.
	zeroTime := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	diff := uint64(currentTime.Sub(zeroTime))
	currentTimeInTicks := diff / 100
	UTCTicks_DescendincurrentTimeOrder := maxValueDateTimeInTicks - currentTimeInTicks

	return UTCTicks_DescendincurrentTimeOrder, nil
}

//period is in the format "60"
//RETURNS: period in the format "PT1M"
func GetPeriodStr(period string) (string, error) {

	var periodStr string

	totalSeconds, err := strconv.Atoi(strings.Trim(period, "s"))

	if err != nil {
		log.Println("E! ERROR while parsing period." + period)
		log.Print(err.Error())
		return "", err
	}

	hour := (int)(math.Floor(float64(totalSeconds) / 3600))
	min := int(math.Floor(float64(totalSeconds-(hour*3600)) / 60))
	sec := totalSeconds - (hour * 3600) - (min * 60)
	periodStr = PT
	if hour > 0 {
		periodStr += strconv.Itoa(hour) + H
	}
	if min > 0 {
		periodStr += strconv.Itoa(min) + M
	}
	if sec > 0 {
		periodStr += strconv.Itoa(sec) + S
	}
	return periodStr, nil
}

func ToFileTime(mdsdTime MdsdTime) (int64, error) {
	//check for int64 overflow
	fileTimeSeconds, ok := overflow.Add64(EPOCH_DIFFERENCE, mdsdTime.Seconds)
	if ok == false {
		erMsg := "E! ERROR integer64 overflow while computing EPOCH_DIFFERENCE + mdsdTime.seconds " +
			strconv.FormatInt(EPOCH_DIFFERENCE, 10) + " + " + strconv.FormatInt(mdsdTime.Seconds, 10)
		log.Print(erMsg)
		err := errors.New(erMsg)
		return int64(0), err
	}

	fileTimeTickPerSecond, ok := overflow.Mul64(fileTimeSeconds, TICKS_PER_SECOND)
	if ok == false {
		erMsg := "E! ERROR integer64 overflow while computing fileTimeSeconds * TICKS_PER_SECOND " +
			strconv.FormatInt(fileTimeSeconds, 10) + " * " + strconv.FormatInt(TICKS_PER_SECOND, 10)
		log.Print(erMsg)
		err := errors.New(erMsg)
		return int64(0), err
	}
	fileTime, ok := overflow.Add64(fileTimeTickPerSecond, mdsdTime.MicroSeconds*10)
	if ok == false {
		erMsg := "E! ERROR integer64 overflow while computing fileTimeTickPerSecond + mdsdTime.microSeconds*10" +
			strconv.FormatInt(fileTimeTickPerSecond, 10) + " +* " + strconv.FormatInt(mdsdTime.MicroSeconds, 10)
		log.Print(erMsg)
		err := errors.New(erMsg)
		return int64(0), err
	}

	return fileTime, nil
}

func ToMdsdTime(fileTime int64) MdsdTime {
	mdsdTime := MdsdTime{0, 0}
	mdsdTime.MicroSeconds = (fileTime % TICKS_PER_SECOND) / 10
	mdsdTime.Seconds = (fileTime / TICKS_PER_SECOND) - EPOCH_DIFFERENCE
	return mdsdTime
}
