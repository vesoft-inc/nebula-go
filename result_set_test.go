/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package nebula_go

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vesoft-inc/nebula-go/v2/nebula"
	"github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

func TestIsEmpty(t *testing.T) {
	value := nebula.Value{}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, "", valWrap.String())
	assert.Equal(t, true, valWrap.IsEmpty())
}

func TestAsNull(t *testing.T) {
	null := nebula.NullType___NULL__
	value := nebula.Value{NVal: &null}
	valWrap := ValueWrapper{&value}
	res, _ := valWrap.AsNull()
	assert.Equal(t, "__NULL__", valWrap.String())
	assert.Equal(t, value.GetNVal(), res)
}

func TestAsBool(t *testing.T) {
	bval := new(bool)
	*bval = true
	value := nebula.Value{BVal: bval}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsBool())
	assert.Equal(t, "true", valWrap.String())
	res, _ := valWrap.AsBool()
	assert.Equal(t, value.GetBVal(), res)
}

func TestAsInt(t *testing.T) {
	val := new(int64)
	*val = 100
	value := nebula.Value{IVal: val}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsInt())
	assert.Equal(t, "100", valWrap.String())
	res, _ := valWrap.AsInt()
	assert.Equal(t, value.GetIVal(), res)
}

func TestAsFloat(t *testing.T) {
	val := new(float64)
	*val = 100.111
	value := nebula.Value{FVal: val}
	valWrap := ValueWrapper{&value}
	val2 := new(float64)
	*val2 = 100.00
	value2 := nebula.Value{FVal: val2}
	valWrap2 := ValueWrapper{&value2}
	assert.Equal(t, "100.111", valWrap.String())
	assert.Equal(t, "100.0", valWrap2.String())
	assert.Equal(t, true, valWrap.IsFloat())
	res, _ := valWrap.AsFloat()
	assert.Equal(t, value.GetFVal(), res)
}

func TestAsString(t *testing.T) {
	val := "test_string"
	value := nebula.Value{SVal: []byte(val)}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsString())
	assert.Equal(t, "\"test_string\"", valWrap.String())
	res, _ := valWrap.AsString()
	assert.Equal(t, string(value.GetSVal()), res)
}

func TestAsList(t *testing.T) {
	var valList = []*nebula.Value{
		&nebula.Value{SVal: []byte("elem1")},
		&nebula.Value{SVal: []byte("elem2")},
		&nebula.Value{SVal: []byte("elem3")},
	}
	value := nebula.Value{
		LVal: &nebula.NList{Values: valList},
	}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, "[\"elem1\", \"elem2\", \"elem3\"]", valWrap.String())
	assert.Equal(t, true, valWrap.IsList())

	res, _ := valWrap.AsList()
	for i := 0; i < len(res); i++ {
		strTemp, err := res[i].AsString()
		if err != nil {
			t.Error(err.Error())
		}
		assert.Equal(t, string(valList[i].GetSVal()), strTemp)
	}
}

func TestAsDedupList(t *testing.T) {
	var valList = []*nebula.Value{
		&nebula.Value{SVal: []byte("elem1")},
		&nebula.Value{SVal: []byte("elem2")},
		&nebula.Value{SVal: []byte("elem3")},
	}
	value := nebula.Value{
		UVal: &nebula.NSet{Values: valList},
	}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, "[\"elem1\", \"elem2\", \"elem3\"]", valWrap.String())
	assert.Equal(t, true, valWrap.IsSet())

	res, _ := valWrap.AsList()
	for i := 0; i < len(res); i++ {
		strTemp, err := res[i].AsString()
		if err != nil {
			t.Error(err.Error())
		}
		assert.Equal(t, string(valList[i].GetSVal()), strTemp)
	}
}

func TestAsMap(t *testing.T) {
	valueMap := make(map[string]*nebula.Value)
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("key%d", i)
		val := fmt.Sprintf("val%d", i)
		valueMap[key] = &nebula.Value{SVal: []byte(val)}
	}
	mval := nebula.NMap{Kvs: valueMap}
	value := nebula.Value{MVal: &mval}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, "{key0: \"val0\", key1: \"val1\", key2: \"val2\"}", valWrap.String())
	assert.Equal(t, true, valWrap.IsMap())
	vMap := value.GetMVal().Kvs
	valWrapMap, err := valWrap.AsMap()
	if err != nil {
		t.Error(err.Error())
	}
	for i := 0; i < len(vMap); i++ {
		key := fmt.Sprintf("key%d", i)
		str, _ := valWrapMap[key].AsString()
		assert.Equal(t, string(vMap[key].GetSVal()), str)
	}
}

// TODO: add tests for AsTime/Date/DateTime when service supports timezone
func TestAsDate(t *testing.T) {
	value := nebula.Value{DVal: &nebula.Date{2020, 12, 25}}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsDate())
	assert.Equal(t, "2020-12-25", valWrap.String())
}

func TestAsTime(t *testing.T) {
	value := nebula.Value{TVal: &nebula.Time{13, 12, 25, 29}}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsTime())
	assert.Equal(t, "13:12:25.029", valWrap.String())
}

func TestAsDateTime(t *testing.T) {
	value := nebula.Value{DtVal: &nebula.DateTime{2020, 12, 25, 13, 12, 25, 29}}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsDateTime())
	assert.Equal(t, "2020-12-25T13:12:25.029", valWrap.String())
}

func TestAsNode(t *testing.T) {
	value := nebula.Value{VVal: getVertex("Adam", 3, 5)}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsVertex())
	assert.Equal(t,
		"(\"Adam\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})",
		valWrap.String())
	res, _ := valWrap.AsNode()
	node, _ := genNode(value.GetVVal())
	assert.Equal(t, *node, *res)

	// Vertex without tag
	value = nebula.Value{VVal: getVertex("Adam", 0, 0)}
	valWrap = ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsVertex())
	assert.Equal(t,
		"(\"Adam\")",
		valWrap.String())
	res, _ = valWrap.AsNode()
	node, _ = genNode(value.GetVVal())
	assert.Equal(t, *node, *res)
}

func TestAsRelationship(t *testing.T) {
	// [:classmate "Alice"->"Bob" @100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]
	value := nebula.Value{EVal: getEdge("Alice", "Bob", 5)}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsEdge())
	assert.Equal(t, "[:classmate \"Alice\"->\"Bob\" @100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]", valWrap.String())
	res, _ := valWrap.AsRelationship()
	relationship, _ := genRelationship(value.GetEVal())
	assert.Equal(t, *relationship, *res)

	// edge without prop
	value = nebula.Value{EVal: getEdge("Alice", "Bob", 0)}
	valWrap = ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsEdge())
	assert.Equal(t, "[:classmate \"Alice\"->\"Bob\" @100 {}]", valWrap.String())
	res, _ = valWrap.AsRelationship()
	relationship, _ = genRelationship(value.GetEVal())
	assert.Equal(t, *relationship, *res)
}

func TestAsPathWrapper(t *testing.T) {
	//("Tim Duncan" :tag0{prop0: 0, prop1: 1})-[:serve@0]->("Spurs")<-[:serve@0]-("Tony Parker" :tag0{prop0: 0, prop1: 1})
	value := nebula.Value{PVal: getPath("Alice", 5)}
	valWrap := ValueWrapper{&value}
	assert.Equal(t, true, valWrap.IsPath())
	assert.Equal(t,
		"<(\"Alice\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]->"+
			"(\"vertex0\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})<-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]-"+
			"(\"vertex1\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]->"+
			"(\"vertex2\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})<-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]-"+
			"(\"vertex3\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]->"+
			"(\"vertex4\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})>",
		valWrap.String())
	res, _ := valWrap.AsPath()
	path, _ := genPathWrapper(value.GetPVal())
	assert.Equal(t, *path, *res)
}

func TestNode(t *testing.T) {
	vertex := getVertex("Tom", 3, 5)
	node, err := genNode(vertex)
	if err != nil {
		t.Errorf(err.Error())
	}

	assert.Equal(t, "\"Tom\"", node.GetID().String())
	assert.Equal(t, true, node.HasTag("tag1"))
	assert.Equal(t, []string{"tag0", "tag1", "tag2"}, node.GetTags())
	keys, _ := node.Keys("tag1")
	keysCopy := make([]string, len(keys))
	copy(keysCopy, keys)
	sort.Strings(keysCopy)
	assert.Equal(t, []string{"prop0", "prop1", "prop2", "prop3", "prop4"}, keysCopy)
	props, _ := node.Properties("tag1")
	for i := 0; i < len(keysCopy); i++ {
		actualVal, err := props[keysCopy[i]].AsInt()
		if err != nil {
			t.Errorf(err.Error())
		}
		assert.Equal(t, int64(i), actualVal)
	}
}

func TestRelationship(t *testing.T) {
	edge := getEdge("Tom", "Lily", 5)
	relationship, err := genRelationship(edge)
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Equal(t, "\"Tom\"", relationship.GetSrcVertexID().String())
	assert.Equal(t, "\"Lily\"", relationship.GetDstVertexID().String())
	assert.Equal(t, "classmate", relationship.GetEdgeName())
	assert.Equal(t, int64(100), relationship.GetRanking())
	keys := relationship.Keys()
	keysCopy := make([]string, len(keys))
	copy(keysCopy, keys)
	sort.Strings(keysCopy)
	assert.Equal(t, []string{"prop0", "prop1", "prop2", "prop3", "prop4"}, keysCopy)
	props := relationship.Properties()
	for i := 0; i < len(keysCopy); i++ {
		actualVal, err := props[keysCopy[i]].AsInt()
		if err != nil {
			t.Errorf(err.Error())
		}
		assert.Equal(t, int64(i), actualVal)
	}
}

func TestPathWrapper(t *testing.T) {
	path := getPath("Tom", 5)
	pathWrapper, err := genPathWrapper(path)
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Equal(t, 5, pathWrapper.GetPathLength())
	node, err := genNode(getVertex("Tom", 3, 5))
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Equal(t, true, pathWrapper.ContainsNode(*node))
	relationship, err := genRelationship(getEdge("Tom", "vertex0", 5))
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Equal(t, true, pathWrapper.ContainsRelationship(relationship))

	var nodeList []Node
	nodeList = append(nodeList, *node)
	for i := 0; i < 5; i++ {
		genNode, err := genNode(getVertex(fmt.Sprintf("vertex%d", i), 3, 5))
		if err != nil {
			t.Errorf(err.Error())
		}
		nodeList = append(nodeList, *genNode)
	}

	var relationshipList []*Relationship
	relationshipList = append(relationshipList, relationship)
	for i := 0; i < 4; i++ {
		var edge *nebula.Edge
		if i%2 == 0 {
			edge = getEdge(fmt.Sprintf("vertex%d", i+1), fmt.Sprintf("vertex%d", i), 5)
		} else {
			edge = getEdge(fmt.Sprintf("vertex%d", i), fmt.Sprintf("vertex%d", i+1), 5)
		}
		newRelationship, err := genRelationship(edge)
		if err != nil {
			t.Errorf(err.Error())
		}
		relationshipList = append(relationshipList, newRelationship)
	}

	l1 := pathWrapper.GetNodes()
	for i := 0; i < len(nodeList); i++ {
		assert.Equal(t, nodeList[i].GetID(), l1[i].GetID())
	}
	l2 := pathWrapper.GetRelationships()
	for i := 0; i < len(relationshipList); i++ {
		assert.Equal(t, true, relationshipList[i].IsEqualTo(l2[i]))
	}
	// Check segments
	segList := pathWrapper.GetSegments()
	srcList := []string{"\"Tom\"", "\"vertex1\"", "\"vertex1\"", "\"vertex3\"", "\"vertex3\""}
	dstList := []string{"\"vertex0\"", "\"vertex0\"", "\"vertex2\"", "\"vertex2\"", "\"vertex4\""}
	for i := 0; i < len(segList); i++ {
		assert.Equal(t, srcList[i], segList[i].startNode.GetID().String())
		assert.Equal(t, dstList[i], segList[i].endNode.GetID().String())
	}
	startNode, _ := pathWrapper.GetStartNode()
	endNode, _ := pathWrapper.GetEndNode()
	assert.Equal(t, "\"Tom\"", startNode.GetID().String())
	assert.Equal(t, "\"vertex4\"", endNode.GetID().String())
}

func TestResultSet(t *testing.T) {
	respWithNil := &graph.ExecutionResponse{
		graph.ErrorCode_E_STATEMENT_EMPTY,
		1000,
		nil,
		nil,
		nil,
		nil,
		nil}
	resultSetWithNil := genResultSet(respWithNil)
	assert.Equal(t, ErrorCode_E_STATEMENT_EMPTY, resultSetWithNil.GetErrorCode())
	assert.Equal(t, int32(1000), resultSetWithNil.GetLatency())
	assert.Equal(t, "", resultSetWithNil.GetErrorMsg())
	assert.Equal(t, "", resultSetWithNil.GetSpaceName())
	assert.Equal(t, "", resultSetWithNil.GetComment())
	assert.Equal(t, false, resultSetWithNil.IsSucceed())

	planDesc := graph.PlanDescription{
		[]*graph.PlanNodeDescription{
			&graph.PlanNodeDescription{
				[]byte("Project"),
				0,
				[]byte("__Project_0"),
				[]*graph.Pair{},
				[]*graph.ProfilingStats{},
				nil,
				[]int64{2}},
			&graph.PlanNodeDescription{
				[]byte("Start"),
				2,
				[]byte("__Start_2"),
				[]*graph.Pair{},
				[]*graph.ProfilingStats{},
				nil,
				[]int64{}},
		},
		map[int64]int64{0: 0, 2: 1},
		[]byte("dot"),
		0,
	}

	resp := &graph.ExecutionResponse{
		graph.ErrorCode_SUCCEEDED,
		1000,
		getDateset(),
		[]byte("test_space"),
		[]byte("test_err_msg"),
		&planDesc,
		[]byte("test_comment")}
	resultSet := genResultSet(resp)

	assert.Equal(t, ErrorCode_SUCCEEDED, resultSet.GetErrorCode())
	assert.Equal(t, int32(1000), resultSet.GetLatency())
	assert.Equal(t, "test_err_msg", resultSet.GetErrorMsg())
	assert.Equal(t, "test_space", resultSet.GetSpaceName())
	assert.Equal(t, "test_comment", resultSet.GetComment())
	assert.Equal(t, true, resultSet.IsSucceed())

	rowSize := resultSet.GetRowSize()
	colSize := resultSet.GetColSize()
	assert.Equal(t, 1, rowSize)
	assert.Equal(t, 5, colSize)
	assert.Equal(t, false, resultSet.IsEmpty())

	expectedColNames := []string{"col0_int", "col1_string", "col2_vertex", "col3_edge", "col4_path"}
	colNames := resultSet.GetColNames()
	for i := 0; i < len(colNames); i++ {
		assert.Equal(t, expectedColNames[i], colNames[i])
	}

	record, err := resultSet.GetRowValuesByIndex(0)
	if err != nil {
		t.Fatalf(err.Error())
	}
	temp, err := record.GetValueByIndex(0)
	_, err = temp.AsNode()
	assert.EqualError(t, err, "Failed to convert value int to Node, value is not an vertex")
	temp, err = record.GetValueByColName("col2")
	assert.EqualError(t, err, "Failed to get values, given column name 'col2' does not exist")
	val, _ := record.GetValueByColName("col2_vertex")
	node, _ := val.AsNode()
	assert.Equal(t, "\"Tom\"", node.GetID().String())

	// Check get row values
	_, err = resultSet.GetRowValuesByIndex(10)
	assert.EqualError(t, err, "Failed to get Value, the index is out of range")

	vlist := record._record

	expected_v1, err := vlist[0].AsInt()
	expected_v2, err := vlist[1].AsString()
	expected_v3, err := vlist[2].AsNode()
	expected_v4, err := vlist[3].AsRelationship()
	expected_v5, err := vlist[4].AsPath()

	v1 := int64(1)
	v2 := "value1"
	v3, err := genNode(getVertex("Tom", 3, 5))
	v4, err := genRelationship(getEdge("Tom", "Lily", 5))
	v5, err := genPathWrapper(getPath("Tom", 3))

	assert.Equal(t, v1, expected_v1)
	assert.Equal(t, v2, expected_v2)
	assert.Equal(t, v3.GetID(), expected_v3.GetID())
	assert.Equal(t, true, v4.IsEqualTo(expected_v4))
	assert.Equal(t, true, v5.IsEqualTo(expected_v5))

	// Check plan description
	assert.Equal(t,
		"digraph exec_plan "+
			"{\n\trankdir=BT;\n\t\"Project_0\"[label=\"{Project_0|outputVar: "+
			"__Project_0|inputVar: }\", shape=Mrecord];\n\t\"Start_2\"->\"Project_0\";"+
			"\n\t\"Start_2\"[label=\"{Start_2|outputVar: __Start_2|inputVar: }\", shape=Mrecord];\n}",
		resultSet.MakeDotGraph())
}

func TestAsStringTable(t *testing.T) {
	resp := &graph.ExecutionResponse{
		graph.ErrorCode_SUCCEEDED,
		1000,
		getDateset(),
		[]byte("test_space"),
		[]byte("test"),
		graph.NewPlanDescription(),
		[]byte("test_comment")}
	resultSet := genResultSet(resp)

	table := resultSet.AsStringTable()
	var r string
	for i := 0; i < len(table); i++ {
		for _, col := range table[i] {
			r += col + ", "
		}
		if i == 0 {
			assert.Equal(t,
				"col0_int, col1_string, col2_vertex, col3_edge, col4_path, ",
				r)
		}
		if i == 1 {
			assert.Equal(t,
				"1, \"value1\", "+
					"(\"Tom\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} :tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}), [:classmate \"Tom\"->\"Lily\" @100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}], "+
					"<(\"Tom\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]->"+
					"(\"vertex0\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})<-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]-"+
					"(\"vertex1\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]->"+
					"(\"vertex2\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})>, ",
				r)
		}
		r = ""
	}
}

func TestIntVid(t *testing.T) {
	vertex := getVertexInt(101, 3, 5)
	node, err := genNode(vertex)
	if err != nil {
		t.Errorf(err.Error())
	}

	assert.Equal(t, "101", node.GetID().String())
	assert.Equal(t, true, node.HasTag("tag1"))
	assert.Equal(t, []string{"tag0", "tag1", "tag2"}, node.GetTags())
	keys, _ := node.Keys("tag1")
	keysCopy := make([]string, len(keys))
	copy(keysCopy, keys)
	sort.Strings(keysCopy)
	assert.Equal(t, []string{"prop0", "prop1", "prop2", "prop3", "prop4"}, keysCopy)
	props, _ := node.Properties("tag1")
	for i := 0; i < len(keysCopy); i++ {
		actualVal, err := props[keysCopy[i]].AsInt()
		if err != nil {
			t.Errorf(err.Error())
		}
		assert.Equal(t, int64(i), actualVal)
	}
	assert.Equal(t, true, node.GetID().IsInt())
}

func getVertex(vid string, tagNum int, propNum int) *nebula.Vertex {
	var tags []*nebula.Tag
	var vidVal = nebula.NewValue()
	vidVal.SVal = []byte(vid)

	for i := 0; i < tagNum; i++ {
		props := make(map[string]*nebula.Value)
		for j := 0; j < propNum; j++ {
			value := setIVal(j)
			key := fmt.Sprintf("prop%d", j)
			props[key] = value
		}
		tag := nebula.Tag{
			Name:  []byte(fmt.Sprintf("tag%d", i)),
			Props: props,
		}
		tags = append(tags, &tag)
	}
	return &nebula.Vertex{
		Vid:  vidVal,
		Tags: tags,
	}
}

func getVertexInt(vid int, tagNum int, propNum int) *nebula.Vertex {
	var tags []*nebula.Tag
	var vidVal = nebula.NewValue()
	newNum := new(int64)
	*newNum = int64(vid)
	vidVal.IVal = newNum

	for i := 0; i < tagNum; i++ {
		props := make(map[string]*nebula.Value)
		for j := 0; j < propNum; j++ {
			value := setIVal(j)
			key := fmt.Sprintf("prop%d", j)
			props[key] = value
		}
		tag := nebula.Tag{
			Name:  []byte(fmt.Sprintf("tag%d", i)),
			Props: props,
		}
		tags = append(tags, &tag)
	}
	return &nebula.Vertex{
		Vid:  vidVal,
		Tags: tags,
	}
}

func getEdge(srcID string, dstID string, propNum int) *nebula.Edge {
	var srcVidVal = nebula.NewValue()
	var dstVidVal = nebula.NewValue()
	srcVidVal.SVal = []byte(srcID)
	dstVidVal.SVal = []byte(dstID)

	props := make(map[string]*nebula.Value)
	for i := 0; i < propNum; i++ {
		value := setIVal(i)
		props[fmt.Sprintf("prop%d", i)] = value
	}

	return &nebula.Edge{
		Src:     srcVidVal,
		Dst:     dstVidVal,
		Type:    1,
		Name:    []byte("classmate"),
		Ranking: 100,
		Props:   props,
	}
}

func getPath(startID string, stepNum int) *nebula.Path {
	var steps []*nebula.Step
	for i := 0; i < stepNum; i++ {
		props := make(map[string]*nebula.Value)
		for j := 0; j < 5; j++ {
			value := setIVal(j)
			props[fmt.Sprintf("prop%d", j)] = value
		}
		var edgeType nebula.EdgeType
		edgeType = 1
		if i%2 != 0 {
			edgeType = -1
		}
		dstID := getVertex(fmt.Sprintf("vertex%d", i), 3, 5)
		steps = append(steps, &nebula.Step{
			Dst:     dstID,
			Type:    edgeType,
			Name:    []byte("classmate"),
			Ranking: 100,
			Props:   props,
		})
	}
	start := getVertex(startID, 3, 5)
	return &nebula.Path{
		Src:   start,
		Steps: steps,
	}
}

func getDateset() *nebula.DataSet {
	colNames := [][]byte{
		[]byte("col0_int"),
		[]byte("col1_string"),
		[]byte("col2_vertex"),
		[]byte("col3_edge"),
		[]byte("col4_path"),
	}
	var v1 = nebula.NewValue()
	newNum := new(int64)
	*newNum = int64(1)
	v1.IVal = newNum
	var v2 = nebula.NewValue()
	v2.SVal = []byte("value1")
	var v3 = nebula.NewValue()
	v3.VVal = getVertex("Tom", 3, 5)
	var v4 = nebula.NewValue()
	v4.EVal = getEdge("Tom", "Lily", 5)
	var v5 = nebula.NewValue()
	v5.PVal = getPath("Tom", 3)

	var valueList []*nebula.Value
	valueList = []*nebula.Value{v1, v2, v3, v4, v5}
	var rows []*nebula.Row
	row := &nebula.Row{
		valueList,
	}
	rows = append(rows, row)
	return &nebula.DataSet{
		ColumnNames: colNames,
		Rows:        rows,
	}
}

func setIVal(ival int) *nebula.Value {
	var value = nebula.NewValue()
	newNum := new(int64)
	*newNum = int64(ival)
	value.IVal = newNum
	return value
}
