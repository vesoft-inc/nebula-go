package nebula_go

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// StmtWrapper create stmt with a template and params.
// StmtWrapper use?{} as the parameter tag, Similarly, support Date, Time, DateTime, Timestamp, List, Map in ngql,
func StmtWrapper(stmt string, params map[string]interface{}) (string, error) {
	sb := strings.Builder{}

	for i := 0; i < len(stmt); i++ {

		if i >= len(stmt)-3 {
			sb.WriteString(stmt[i:])
			break
		}

		if stmt[i] == '?' && stmt[i+1] == '{' {
			var j = i
			for ; i < len(stmt); j++ {
				if stmt[j] == '}' {
					break
				}
			}
			k := stmt[i+2 : j]

			v, has := params[k]
			if !has {
				return "", fmt.Errorf("param %s not found", k)
			}
			val, err := getValue(v)
			if err != nil {
				return "", err
			}

			sb.WriteString(val)
			i = j
		} else {
			sb.WriteByte(stmt[i])
		}
	}

	return sb.String(), nil
}

func getValue(v interface{}) (string, error) {
	if f, ok := v.(queryFormatter); ok {
		return f.QueryFormat(), nil
	}
	return getValueReflect(v)
}

func getValueReflect(v interface{}) (string, error) {
	vt := reflect.ValueOf(v)

	switch vt.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(vt.Bool()), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(vt.Float(), 'f', 10, 64), nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return strconv.FormatInt(vt.Int(), 10), nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		u := vt.Uint()
		if u > math.MaxInt64 {
			return "", fmt.Errorf("not supported beyond int64")
		}
		return strconv.FormatUint(u, 10), nil
	case reflect.String:
		return "\"" + vt.String() + "\"", nil
	case reflect.Slice:
		return sliceFormat(v)
	case reflect.Map:
		return mapFormat(v)

	}
	return "", fmt.Errorf("param type not support ")
}

func sliceFormat(x interface{}) (string, error) {
	//Optimal performance is greater than reflection
	sb := strings.Builder{}

	sb.WriteByte('[')

	switch v := x.(type) {
	case []interface{}:
		for i := range v {
			val, err := getValue(v[i])
			if err != nil {
				return "", err
			}
			sb.WriteString(val)
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []string:
		for i := range v {
			sb.WriteString("\"" + v[i] + "\"")
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []int64:
		for i := range v {
			sb.WriteString(strconv.FormatInt(v[i], 10))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []uint64:
		for i := range v {
			if v[i] > math.MaxInt64 {
				return "", fmt.Errorf("not supported beyond int64")
			}
			sb.WriteString(strconv.FormatUint(v[i], 10))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []int:
		for i := range v {
			sb.WriteString(strconv.Itoa(v[i]))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []uint8:
		for i := range v {
			sb.WriteString(strconv.FormatUint(uint64(v[i]), 10))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []int8:
		for i := range v {
			sb.WriteString(strconv.FormatInt(int64(v[i]), 10))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}

	case []uint16:
		for i := range v {
			sb.WriteString(strconv.FormatUint(uint64(v[i]), 10))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []int16:
		for i := range v {
			sb.WriteString(strconv.FormatInt(int64(v[i]), 10))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []uint32:
		for i := range v {
			sb.WriteString(strconv.FormatUint(uint64(v[i]), 10))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []int32:
		for i := range v {
			sb.WriteString(strconv.FormatInt(int64(v[i]), 10))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []float64:
		for i := range v {
			sb.WriteString(strconv.FormatFloat(v[i], 'f', 10, 64))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	case []float32:
		for i := range v {
			sb.WriteString(strconv.FormatFloat(float64(v[i]), 'f', 10, 64))
			if i != len(v)-1 {
				sb.WriteByte(',')
			}
		}
	default:
		return "", fmt.Errorf("param type not support ")
	}

	sb.WriteByte(']')
	return sb.String(), nil
}

func mapFormat(x interface{}) (string, error) {
	//Optimal performance is greater than reflection
	sb := strings.Builder{}
	sb.WriteByte('{')

	switch m := x.(type) {
	case map[string]interface{}:
		cnt := 0
		for k, v := range m {
			val, err := getValue(v)
			if err != nil {
				return "", err
			}
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(val)
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}

	case map[string]string:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString("\"" + v + "\"")
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]float64:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatFloat(v, 'f', 10, 64))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]float32:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatFloat(float64(v), 'f', 10, 64))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]int:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(v))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]int8:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatInt(int64(v), 10))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]uint8:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatUint(uint64(v), 10))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]int16:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatInt(int64(v), 10))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]uint16:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatUint(uint64(v), 10))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]int32:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatInt(int64(v), 10))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]uint32:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatUint(uint64(v), 10))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]int64:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatInt(v, 10))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]uint64:
		cnt := 1
		for k, v := range m {
			if v > math.MaxInt64 {
				return "", fmt.Errorf("not supported beyond int64")
			}
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatUint(v, 10))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	case map[string]bool:
		cnt := 1
		for k, v := range m {
			cnt += 1
			sb.WriteString(k)
			sb.WriteByte(':')
			sb.WriteString(strconv.FormatBool(v))
			if cnt == len(m) {
				sb.WriteByte(',')
			}
		}
	default:
		return "", fmt.Errorf("param type not support ")
	}
	sb.WriteByte('}')
	return sb.String(), nil
}

type queryFormatter interface {
	QueryFormat() string
}

const (
	dateFormat     = "2006-01-02"
	timeFormat     = "15:04:05.000"
	datetimeFormat = "2006-01-02T15:04:05"
)

type Date struct {
	time.Time
}

func (t Date) QueryFormat() string {
	return "date(\"" + t.Format(dateFormat) + "\")"
}

func DateOf(t time.Time) Date {
	return Date{Time: t}
}

func DateOfNumber(y, m, d int) Date {
	return Date{time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)}
}

type Time struct {
	time.Time
}

func (t Time) QueryFormat() string {
	return "time(\"" + t.Format(timeFormat) + "\")"
}

func TimeOfNumber(h, m, s int) Time {
	return Time{time.Date(0, 0, 0, h, m, s, 0, time.Local)}
}

func TimeOf(t time.Time) Time {
	return Time{Time: t}
}

type DateTime struct {
	time.Time
}

func (t DateTime) QueryFormat() string {
	return "datetime(\"" + t.Format(datetimeFormat) + "\")"
}

func DateTimeOf(t time.Time) DateTime {
	return DateTime{Time: t}
}

type Timestamp struct {
	data int64
}

func (t Timestamp) QueryFormat() string {
	return strconv.FormatInt(t.data, 10)
}

func TimestampOfNumber(t int64) Timestamp {
	return Timestamp{data: t}
}
func TimestampOf(t time.Time) Timestamp {
	return Timestamp{data: t.Unix()}
}

type NULL struct{}

func (n NULL) QueryFormat() string {
	return "null"
}
