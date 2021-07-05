package nebula_go

import (
	"fmt"
	"testing"
	"time"
)

func TestStmtWrapper(t *testing.T) {
	time.Local = time.UTC
	type args struct {
		stmt   string
		params map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple insert vertex",
			args: args{
				stmt: "INSERT VERTEX t2 (name, age) VALUES ?{vid}:(?{name}, ?{age});",
				params: map[string]interface{}{
					"vid":  "vid_1",
					"name": "foo",
					"age":  123,
				},
			},
			want:    `INSERT VERTEX t2 (name, age) VALUES "vid_1":("foo", 123);`,
			wantErr: false,
		},
		{
			name: "simple insert edge",
			args: args{
				stmt: "INSERT EDGE e2 (name, age) VALUES ?{from}->?{to}@1:(?{name}, ?{age});",
				params: map[string]interface{}{
					"from": "vid_1",
					"to":   "vid_2",
					"name": "foo",
					"age":  123,
				},
			},
			want: `INSERT EDGE e2 (name, age) VALUES "vid_1"->"vid_2"@1:("foo", 123);`,

			wantErr: false,
		},
		{
			name: "simple update edge",
			args: args{
				stmt: `UPDATE EDGE on serve ?{from} -> ?{to}@0
        SET start_year = start_year + 1
        WHEN end_year > ?{year}
        YIELD start_year, end_year;`,
				params: map[string]interface{}{
					"from": "vid_1",
					"to":   "vid_2",
					"year": 2001,
				},
			},
			want: `UPDATE EDGE on serve "vid_1" -> "vid_2"@0
        SET start_year = start_year + 1
        WHEN end_year > 2001
        YIELD start_year, end_year;`,

			wantErr: false,
		},

		{
			name: "simple match",
			args: args{
				stmt: "MATCH p=(v:player{name:?{name}})-[e:follow*1..3]->(v2) RETURN v2 AS Friends",
				params: map[string]interface{}{
					"name": "Tim Duncan",
				},
			},
			want:    `MATCH p=(v:player{name:"Tim Duncan"})-[e:follow*1..3]->(v2) RETURN v2 AS Friends`,
			wantErr: false,
		},
		{
			name: "time",
			args: args{
				stmt: `INSERT VERTEX date1(p1, p2, p3, p4)
				VALUES "test1":(?{date}, ?{time}, ?{datetime}, ?{timestamp});`,
				params: map[string]interface{}{
					"date":      DateOfNumber(2021, 3, 17),
					"time":      TimeOfNumber(17, 53, 59),
					"datetime":  DateTimeOf(time.Date(2021, 3, 17, 17, 53, 59, 0, time.Local)),
					"timestamp": TimestampOf(time.Date(2021, 3, 17, 17, 53, 59, 0, time.Local)),
				},
			},
			want: `INSERT VERTEX date1(p1, p2, p3, p4)
				VALUES "test1":(date("2021-03-17"), time("17:53:59.000"), datetime("2021-03-17T17:53:59"), 1616003639);`,
			wantErr: false,
		},
		{
			name: "match with list string",
			args: args{
				stmt: `MATCH (v:player { name: ?{name} })--(v2)
        WHERE id(v2) IN ?{list} RETURN v2;`,
				params: map[string]interface{}{
					"name": "Tim Duncan",
					"list": []string{"player101", "player102"},
				},
			},
			want: `MATCH (v:player { name: "Tim Duncan" })--(v2)
        WHERE id(v2) IN ["player101","player102"] RETURN v2;`,
			wantErr: false,
		},
		{
			name: "match with list int",
			args: args{
				stmt: `MATCH (v:player { name: ?{name} })--(v2)
        WHERE id(v2) IN ?{list} RETURN v2;`,
				params: map[string]interface{}{
					"name": "Tim Duncan",
					"list": []int32{123, 456},
				},
			},
			want: `MATCH (v:player { name: "Tim Duncan" })--(v2)
        WHERE id(v2) IN [123,456] RETURN v2;`,
			wantErr: false,
		},

		{
			name: "map list",
			args: args{
				stmt: `YIELD {key: 'Value', listKey: ?{m}}`,
				params: map[string]interface{}{
					"m": []interface{}{map[string]string{"inner": "Map1"}, map[string]int{"inner": 123}},
				},
			},
			want:    `YIELD {key: 'Value', listKey: [{inner:"Map1"},{inner:123}]}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StmtWrapper(tt.args.stmt, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("StmtWrapper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StmtWrapper() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleStmtWrapper() {

	stmt, err := StmtWrapper("INSERT VERTEX t2 (name, age) VALUES ?{vid}:(?{name}, ?{age});", map[string]interface{}{
		"vid":  "vid_1",
		"name": "foo",
		"age":  123,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("insert stmt:", stmt)

	stmt, err = StmtWrapper("INSERT EDGE e2 (name, age) VALUES ?{from}->?{to}@1:(?{name}, ?{age});",
		map[string]interface{}{
			"from": "vid_1",
			"to":   "vid_2",
			"name": "foo",
			"age":  123,
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("insert edge stmt:", stmt)

	stmt, err = StmtWrapper("MATCH (v:player { born: ?{born} })--(v2) WHERE id(v2) IN ?{list} RETURN v2;",
		map[string]interface{}{
			"name": DateOfNumber(2000, 10, 1),
			"list": 123,
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("match stmt:", stmt)

	stmt, err = StmtWrapper("YIELD {listMap: ?{map}, listKey: ?{list}};",
		map[string]interface{}{
			"map":  map[string]string{"foo": "bar"},
			"list": []string{"foo", "bar"},
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("map list stmt:", stmt)
}
