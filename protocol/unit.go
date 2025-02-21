package protocol

import (
	"bytes"
	"fmt"
	bytes2 "github.com/injoyai/base/bytes"
	"github.com/injoyai/conv"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"math"
	"strings"
	"time"
)

// String 字节先转小端,再转字符
func String(bs []byte) string {
	return string(bytes2.Reverse(bs))
}

// Bytes 任意类型转小端字节
func Bytes(n any) []byte {
	return bytes2.Reverse(conv.Bytes(n))
}

// Reverse 字节倒序
func Reverse(bs []byte) []byte {
	x := make([]byte, len(bs))
	for i, v := range bs {
		x[len(bs)-i-1] = v
	}
	return x
}

// Uint32 字节通过小端方式转为uint32
func Uint32(bs []byte) uint32 {
	return conv.Uint32(bytes2.Reverse(bs))
}

// Uint16 字节通过小端方式转为uint16
func Uint16(bs []byte) uint16 {
	return conv.Uint16(bytes2.Reverse(bs))
}

func UTF8ToGBK(text []byte) []byte {
	r := bytes.NewReader(text)
	decoder := transform.NewReader(r, simplifiedchinese.GBK.NewDecoder()) //GB18030
	content, _ := io.ReadAll(decoder)
	return bytes.ReplaceAll(content, []byte{0x00}, []byte{})
}

func DecodeCode(code string) (Exchange, string, error) {
	if len(code) != 8 {
		return 0, "", fmt.Errorf("股票代码长度错误,例如:SZ000001")
	}
	switch strings.ToLower(code[:2]) {
	case ExchangeSH.String():
		return ExchangeSH, code[2:], nil
	case ExchangeSZ.String():
		return ExchangeSZ, code[2:], nil
	default:
		return 0, "", fmt.Errorf("股票代码错误,例如:SZ000001")
	}
}

func FloatUnit(f float64) (float64, string) {
	m := []string{"万", "亿"}
	unit := ""
	for i := 0; f > 1e4 && i < len(m); f /= 1e4 {
		unit = m[i]
	}
	return f, unit
}

func FloatUnitString(f float64) string {
	m := []string{"万", "亿", "万亿", "亿亿", "万亿亿", "亿亿亿"}
	unit := ""
	for i := 0; f > 1e4 && i < len(m); i++ {
		unit = m[i]
		f /= 1e4
	}
	if unit == "" {
		return conv.String(f)
	}
	return fmt.Sprintf("%0.2f%s", f, unit)
}

func IntUnitString(n int) string {
	return FloatUnitString(float64(n))
}

func Int64UnitString(n int64) string {
	return FloatUnitString(float64(n))
}

func GetHourMinute(bs [2]byte) string {
	n := Uint16(bs[:])
	h := n / 60
	m := n % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func GetTime(bs [4]byte, Type uint8) time.Time {
	switch Type {
	case TypeKlineMinute, TypeKlineMinute2, TypeKline5Minute, TypeKline15Minute, TypeKline30Minute, TypeKlineHour:

		yearMonthDay := Uint16(bs[:2])
		hourMinute := Uint16(bs[2:4])
		year := int(yearMonthDay>>11 + 2004)
		month := yearMonthDay % 2048 / 100
		day := int((yearMonthDay % 2048) % 100)
		hour := int(hourMinute / 60)
		minute := int(hourMinute % 60)
		return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local)

	default:

		yearMonthDay := Uint32(bs[:4])
		year := int(yearMonthDay / 10000)
		month := int((yearMonthDay % 10000) / 100)
		day := int(yearMonthDay % 100)
		return time.Date(year, time.Month(month), day, 15, 0, 0, 0, time.Local)

	}
}

func basePrice(code string) Price {
	if len(code) == 0 {
		return 1
	}
	switch code[:2] {
	case "60", "30", "68", "00":
		return 1
	default:
		return 10
	}
}

func getVolume(val uint32) (volume float64) {
	ivol := int32(val)
	logpoint := ivol >> (8 * 3)
	//hheax := ivol >> (8 * 3)          // [3]
	hleax := (ivol >> (8 * 2)) & 0xff // [2]
	lheax := (ivol >> 8) & 0xff       //[1]
	lleax := ivol & 0xff              //[0]

	//dbl_1 := 1.0
	//dbl_2 := 2.0
	//dbl_128 := 128.0

	dwEcx := logpoint*2 - 0x7f
	dwEdx := logpoint*2 - 0x86
	dwEsi := logpoint*2 - 0x8e
	dwEax := logpoint*2 - 0x96
	tmpEax := dwEcx
	if dwEcx < 0 {
		tmpEax = -dwEcx
	} else {
		tmpEax = dwEcx
	}

	dbl_xmm6 := 0.0
	dbl_xmm6 = math.Pow(2.0, float64(tmpEax))
	if dwEcx < 0 {
		dbl_xmm6 = 1.0 / dbl_xmm6
	}

	dbl_xmm4 := 0.0
	dbl_xmm0 := 0.0

	if hleax > 0x80 {
		tmpdbl_xmm3 := 0.0
		//tmpdbl_xmm1 := 0.0
		dwtmpeax := dwEdx + 1
		tmpdbl_xmm3 = math.Pow(2.0, float64(dwtmpeax))
		dbl_xmm0 = math.Pow(2.0, float64(dwEdx)) * 128.0
		dbl_xmm0 += float64(hleax&0x7f) * tmpdbl_xmm3
		dbl_xmm4 = dbl_xmm0
	} else {
		if dwEdx >= 0 {
			dbl_xmm0 = math.Pow(2.0, float64(dwEdx)) * float64(hleax)
		} else {
			dbl_xmm0 = (1 / math.Pow(2.0, float64(dwEdx))) * float64(hleax)
		}
		dbl_xmm4 = dbl_xmm0
	}

	dbl_xmm3 := math.Pow(2.0, float64(dwEsi)) * float64(lheax)
	dbl_xmm1 := math.Pow(2.0, float64(dwEax)) * float64(lleax)
	if (hleax & 0x80) > 0 {
		dbl_xmm3 *= 2.0
		dbl_xmm1 *= 2.0
	}
	volume = dbl_xmm6 + dbl_xmm4 + dbl_xmm3 + dbl_xmm1
	return
}

// IsStock 是否是股票,示例sz000001
func IsStock(code string) bool {
	if len(code) != 8 {
		return false
	}
	code = strings.ToLower(code)
	switch {
	case code[0:2] == ExchangeSH.String() &&
		(code[2:3] == "6"):
		return true

	case code[0:2] == ExchangeSZ.String() &&
		(code[2:3] == "0" || code[2:4] == "30"):
		return true
	}
	return false
}
