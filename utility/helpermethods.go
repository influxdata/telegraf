package utility

import (
	"crypto/md5"
	"encoding/base64"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
)

//https://golang.org/pkg/crypto/md5/
func Getmd5Hash(content string) (string, error) {
	md5HashStr := ""
	md5Hasher := md5.New()
	data := []byte(content)
	v, er := md5Hasher.Write(data)
	if er != nil {
		log.Println("Error while calculating md5 hash of block " + content + er.Error())
		return "", er
	}
	log.Println(string(v))
	md5HashStr = base64.StdEncoding.EncodeToString(md5Hasher.Sum(nil))
	l := len(md5HashStr)
	log.Println(string(l))
	return md5HashStr, nil
}

func GetPropsStr(props map[string]interface{}) string {
	var propsStr string
	for key, value := range props {
		valueStr := ""
		if reflect.TypeOf(value).String() == "float64" {
			valueStr = strconv.FormatFloat(value.(float64), 'E', -1, 64)
		} else {
			valueStr = value.(string)
		}
		propsStr = propsStr + key + " : " + valueStr + " , "
	}
	return propsStr
}

func GetPeriodStr(period string) (string, error) {

	var periodStr string

	totalSeconds, err := strconv.Atoi(strings.Trim(period, "s"))

	if err != nil {
		log.Println(" Error while parsing period." + period)
		log.Print(err.Error())
		return "", err
	}

	hour := (int)(math.Floor(float64(totalSeconds) / 3600))
	min := int(math.Floor(float64(totalSeconds-(hour*3600)) / 60))
	sec := totalSeconds - (hour * 3600) - (min * 60)
	periodStr = "PT"
	if hour > 0 {
		periodStr += strconv.Itoa(hour) + "H"
	}
	if min > 0 {
		periodStr += strconv.Itoa(min) + "M"
	}
	if sec > 0 {
		periodStr += strconv.Itoa(sec) + "S"
	}
	return periodStr, nil
}
