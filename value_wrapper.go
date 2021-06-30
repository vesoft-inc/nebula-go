/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package nebula_go

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/vesoft-inc/nebula-go/v2/nebula"
)

type ValueWrapper struct {
	value        *nebula.Value
	timezoneInfo timezoneInfo
}

func (valWrap ValueWrapper) IsEmpty() bool {
	return valWrap.GetType() == "empty"
}

func (valWrap ValueWrapper) IsNull() bool {
	return valWrap.value.IsSetNVal()
}

func (valWrap ValueWrapper) IsBool() bool {
	return valWrap.value.IsSetBVal()
}

func (valWrap ValueWrapper) IsInt() bool {
	return valWrap.value.IsSetIVal()
}

func (valWrap ValueWrapper) IsFloat() bool {
	return valWrap.value.IsSetFVal()
}

func (valWrap ValueWrapper) IsString() bool {
	return valWrap.value.IsSetSVal()
}

func (valWrap ValueWrapper) IsTime() bool {
	return valWrap.value.IsSetTVal()
}

func (valWrap ValueWrapper) IsDate() bool {
	return valWrap.value.IsSetDVal()
}

func (valWrap ValueWrapper) IsDateTime() bool {
	return valWrap.value.IsSetDtVal()
}

func (valWrap ValueWrapper) IsList() bool {
	return valWrap.value.IsSetLVal()
}

func (valWrap ValueWrapper) IsSet() bool {
	return valWrap.value.IsSetUVal()
}

func (valWrap ValueWrapper) IsMap() bool {
	return valWrap.value.IsSetMVal()
}

func (valWrap ValueWrapper) IsVertex() bool {
	return valWrap.value.IsSetVVal()
}

func (valWrap ValueWrapper) IsEdge() bool {
	return valWrap.value.IsSetEVal()
}

func (valWrap ValueWrapper) IsPath() bool {
	return valWrap.value.IsSetPVal()
}

func (valWrap ValueWrapper) AsNull() (nebula.NullType, error) {
	if valWrap.value.IsSetNVal() {
		return valWrap.value.GetNVal(), nil
	}
	return -1, fmt.Errorf("failed to convert value %s to Null", valWrap.GetType())
}

func (valWrap ValueWrapper) AsBool() (bool, error) {
	if valWrap.value.IsSetBVal() {
		return valWrap.value.GetBVal(), nil
	}
	return false, fmt.Errorf("failed to convert value %s to bool", valWrap.GetType())
}

func (valWrap ValueWrapper) AsInt() (int64, error) {
	if valWrap.value.IsSetIVal() {
		return valWrap.value.GetIVal(), nil
	}
	return -1, fmt.Errorf("failed to convert value %s to int", valWrap.GetType())
}

func (valWrap ValueWrapper) AsFloat() (float64, error) {
	if valWrap.value.IsSetFVal() {
		return valWrap.value.GetFVal(), nil
	}
	return -1, fmt.Errorf("failed to convert value %s to float", valWrap.GetType())
}

func (valWrap ValueWrapper) AsString() (string, error) {
	if valWrap.value.IsSetSVal() {
		return string(valWrap.value.GetSVal()), nil
	}
	return "", fmt.Errorf("failed to convert value %s to string", valWrap.GetType())
}

func (valWrap ValueWrapper) AsTime() (*TimeWrapper, error) {
	if valWrap.value.IsSetTVal() {
		rawTime := valWrap.value.GetTVal()
		time, err := genTimeWrapper(rawTime, valWrap.timezoneInfo)
		if err != nil {
			return nil, err
		}
		return time, nil
	}
	return nil, fmt.Errorf("failed to convert value %s to Time", valWrap.GetType())
}

func (valWrap ValueWrapper) AsDate() (*nebula.Date, error) {
	if valWrap.value.IsSetDVal() {
		return valWrap.value.GetDVal(), nil
	}
	return nil, fmt.Errorf("failed to convert value %s to Date", valWrap.GetType())
}

func (valWrap ValueWrapper) AsDateTime() (*DateTimeWrapper, error) {
	if valWrap.value.IsSetDtVal() {
		rawTimeDate := valWrap.value.GetDtVal()
		timeDate, err := genDateTimeWrapper(rawTimeDate, valWrap.timezoneInfo)
		if err != nil {
			return nil, err
		}
		return timeDate, nil
	}
	return nil, fmt.Errorf("failed to convert value %s to DateTime", valWrap.GetType())
}

func (valWrap ValueWrapper) AsList() ([]ValueWrapper, error) {
	if valWrap.value.IsSetLVal() {
		var varList []ValueWrapper
		vals := valWrap.value.GetLVal().Values
		for _, val := range vals {
			varList = append(varList, ValueWrapper{val, valWrap.timezoneInfo})
		}
		return varList, nil
	}
	return nil, fmt.Errorf("failed to convert value %s to List", valWrap.GetType())
}

func (valWrap ValueWrapper) AsDedupList() ([]ValueWrapper, error) {
	if valWrap.value.IsSetUVal() {
		var varList []ValueWrapper
		vals := valWrap.value.GetUVal().Values
		for _, val := range vals {
			varList = append(varList, ValueWrapper{val, valWrap.timezoneInfo})
		}
		return varList, nil
	}
	return nil, fmt.Errorf("failed to convert value %s to set(deduped list)", valWrap.GetType())
}

func (valWrap ValueWrapper) AsMap() (map[string]ValueWrapper, error) {
	if valWrap.value.IsSetMVal() {
		newMap := make(map[string]ValueWrapper)

		kvs := valWrap.value.GetMVal().Kvs
		for key, val := range kvs {
			newMap[key] = ValueWrapper{val, valWrap.timezoneInfo}
		}
		return newMap, nil
	}
	return nil, fmt.Errorf("failed to convert value %s to Map", valWrap.GetType())
}

func (valWrap ValueWrapper) AsNode() (*Node, error) {
	if !valWrap.value.IsSetVVal() {
		return nil, fmt.Errorf("failed to convert value %s to Node, value is not an vertex", valWrap.GetType())
	}
	vertex := valWrap.value.VVal
	node, err := genNode(vertex, valWrap.timezoneInfo)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (valWrap ValueWrapper) AsRelationship() (*Relationship, error) {
	if !valWrap.value.IsSetEVal() {
		return nil, fmt.Errorf("failed to convert value %s to Relationship, value is not an edge", valWrap.GetType())
	}
	edge := valWrap.value.EVal
	relationship, err := genRelationship(edge, valWrap.timezoneInfo)
	if err != nil {
		return nil, err
	}
	return relationship, nil
}

func (valWrap ValueWrapper) AsPath() (*PathWrapper, error) {
	if !valWrap.value.IsSetPVal() {
		return nil, fmt.Errorf("failed to convert value %s to PathWrapper, value is not an edge", valWrap.GetType())
	}
	path, err := genPathWrapper(valWrap.value.PVal, valWrap.timezoneInfo)
	if err != nil {
		return nil, err
	}
	return path, nil
}

// Returns the value type of value in the valWrap in string
func (valWrap ValueWrapper) GetType() string {
	if valWrap.value.IsSetNVal() {
		return "null"
	} else if valWrap.value.IsSetBVal() {
		return "bool"
	} else if valWrap.value.IsSetIVal() {
		return "int"
	} else if valWrap.value.IsSetFVal() {
		return "float"
	} else if valWrap.value.IsSetSVal() {
		return "string"
	} else if valWrap.value.IsSetDVal() {
		return "date"
	} else if valWrap.value.IsSetTVal() {
		return "time"
	} else if valWrap.value.IsSetDtVal() {
		return "datetime"
	} else if valWrap.value.IsSetVVal() {
		return "vertex"
	} else if valWrap.value.IsSetEVal() {
		return "edge"
	} else if valWrap.value.IsSetPVal() {
		return "path"
	} else if valWrap.value.IsSetLVal() {
		return "list"
	} else if valWrap.value.IsSetMVal() {
		return "map"
	} else if valWrap.value.IsSetUVal() {
		return "set"
	}
	return "empty"
}

// String() returns the value in the ValueWrapper as a string.
//
// Maps in the output will be sorted by key value in alphabetical order.
//
//  For vetex, the output is in form (vid: tagName{propKey: propVal, propKey2, propVal2}),
//  For edge, the output is in form (SrcVid)-[name]->(DstVid)@Ranking{prop1: val1, prop2: val2}
//  where arrow direction depends on edgeType.
//  For path, the output is in form (v1)-[name@edgeRanking]->(v2)-[name@edgeRanking]->(v3)
//
// For time, and dateTime, String returns the value calculated using the timezone offset
// from graph service by default.
func (valWrap ValueWrapper) String() string {
	value := valWrap.value
	if value.IsSetNVal() {
		return value.GetNVal().String()
	} else if value.IsSetBVal() {
		return fmt.Sprintf("%t", value.GetBVal())
	} else if value.IsSetIVal() {
		return fmt.Sprintf("%d", value.GetIVal())
	} else if value.IsSetFVal() {
		fStr := strconv.FormatFloat(value.GetFVal(), 'f', -1, 64)
		if !strings.Contains(fStr, ".") {
			fStr = fStr + ".0"
		}
		return fStr
	} else if value.IsSetSVal() {
		return `"` + string(value.GetSVal()) + `"`
	} else if value.IsSetDVal() { // Date yyyy-mm-dd
		date := value.GetDVal()
		dateWrapper, _ := genDateWrapper(date)
		return fmt.Sprintf("%04d-%02d-%02d",
			dateWrapper.getYear(),
			dateWrapper.getMonth(),
			dateWrapper.getDay())
	} else if value.IsSetTVal() { // Time HH:MM:SS.MSMSMS
		rawTime := value.GetTVal()
		time, _ := genTimeWrapper(rawTime, valWrap.timezoneInfo)
		localTime, _ := time.getLocalTime()
		return fmt.Sprintf("%02d:%02d:%02d.%06d",
			localTime.GetHour(),
			localTime.GetMinute(),
			localTime.GetSec(),
			localTime.GetMicrosec())
	} else if value.IsSetDtVal() { // DateTime yyyy-mm-ddTHH:MM:SS.MSMSMS
		rawDateTime := value.GetDtVal()
		dateTime, _ := genDateTimeWrapper(rawDateTime, valWrap.timezoneInfo)
		localDateTime, _ := dateTime.getLocalDateTime()
		return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d.%06d",
			localDateTime.GetYear(),
			localDateTime.GetMonth(),
			localDateTime.GetDay(),
			localDateTime.GetHour(),
			localDateTime.GetMinute(),
			localDateTime.GetSec(),
			localDateTime.GetMicrosec())
	} else if value.IsSetVVal() { // Vertex format: ("VertexID" :tag1{k0: v0,k1: v1}:tag2{k2: v2})
		vertex := value.GetVVal()
		node, _ := genNode(vertex, valWrap.timezoneInfo)
		return node.String()
	} else if value.IsSetEVal() { // Edge format: [:edge src->dst @ranking {propKey1: propVal1}]
		edge := value.GetEVal()
		relationship, _ := genRelationship(edge, valWrap.timezoneInfo)
		return relationship.String()
	} else if value.IsSetPVal() {
		// Path format:
		// ("VertexID" :tag1{k0: v0,k1: v1})-
		// [:TypeName@ranking {propKey1: propVal1}]->
		// ("VertexID2" :tag1{k0: v0,k1: v1} :tag2{k2: v2})-
		// [:TypeName@ranking {propKey2: propVal2}]->
		// ("VertexID3" :tag1{k0: v0,k1: v1})
		path := value.GetPVal()
		pathWrap, _ := genPathWrapper(path, valWrap.timezoneInfo)
		return pathWrap.String()
	} else if value.IsSetLVal() { // List
		lval := value.GetLVal()
		var strs []string
		for _, val := range lval.Values {
			strs = append(strs, ValueWrapper{val, valWrap.timezoneInfo}.String())
		}
		return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
	} else if value.IsSetMVal() { // Map
		// {k0: v0, k1: v1}
		mval := value.GetMVal()
		var keyList []string
		var output []string
		kvs := mval.Kvs
		for k := range kvs {
			keyList = append(keyList, k)
		}
		sort.Strings(keyList)
		for _, k := range keyList {
			output = append(output, fmt.Sprintf("%s: %s", k, ValueWrapper{kvs[k], valWrap.timezoneInfo}.String()))
		}
		return fmt.Sprintf("{%s}", strings.Join(output, ", "))
	} else if value.IsSetUVal() {
		// set to string
		uval := value.GetUVal()
		var strs []string
		for _, val := range uval.Values {
			strs = append(strs, ValueWrapper{val, valWrap.timezoneInfo}.String())
		}
		return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
	} else { // is empty
		return ""
	}
}
