/*
 *
 * Copyright (c) 2024 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */
package nebula_go

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type _Prop struct {
	name  string
	value string
	isStr bool
}

type EntityBuilder struct {
	statement string
	name      string
	props     []_Prop
}

func NewEntityBuilder(name string) *EntityBuilder {
	return &EntityBuilder{
		name:  name,
		props: []_Prop{},
	}
}

func (b *EntityBuilder) SetProp(name string, value interface{}) *EntityBuilder {
	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.String:
		b.props = append(b.props, _Prop{

			name:  name,
			value: b.escapeStrVal(value.(string)),
			isStr: true,
		})
	case reflect.Int64:
		b.props = append(b.props, _Prop{

			name:  name,
			value: strconv.FormatInt(value.(int64), 10),
			isStr: false,
		})
	case reflect.Int32:
		b.props = append(b.props, _Prop{

			name:  name,
			value: strconv.FormatInt(int64(value.(int32)), 10),
			isStr: false,
		})
	case reflect.Int:
		b.props = append(b.props, _Prop{

			name:  name,
			value: strconv.Itoa(value.(int)),
			isStr: false,
		})
	case reflect.Float64:
		b.props = append(b.props, _Prop{

			name:  name,
			value: strconv.FormatFloat(value.(float64), 'g', -1, 64),
			isStr: false,
		})
	case reflect.Float32:
		b.props = append(b.props, _Prop{

			name:  name,
			value: strconv.FormatFloat(float64(value.(float32)), 'g', -1, 32),
			isStr: false,
		})
	case reflect.Bool:
		b.props = append(b.props, _Prop{

			name:  name,
			value: strconv.FormatBool(value.(bool)),
			isStr: false,
		})
	}
	return b
}

func (b *EntityBuilder) InsertVertex() *EntityBuilder {
	panic("implement me")
}

func (b *EntityBuilder) DeleteVertex(withEdge bool) *EntityBuilder {
	panic("implement me")
}

func (b *EntityBuilder) UpdateVertex() *EntityBuilder {
	panic("implement me")
}

func (b *EntityBuilder) UpsertVertex(vid interface{}) *EntityBuilder {
	id, isStr := b.parseVid(vid)
	q := ""
	if isStr {
		q = fmt.Sprintf(`UPSERT VERTEX ON %s "%s"`, b.name, id)
	} else {
		q = fmt.Sprintf("UPSERT VERTEX ON %s %s", b.name, id)
	}

	if len(b.props) > 0 {
		q += " SET "
		for _, p := range b.props {
			if p.isStr {
				q += fmt.Sprintf(`%s = "%s", `, p.name, p.value)
			} else {
				q += fmt.Sprintf(`%s = %s, `, p.name, p.value)
			}
		}
		q = strings.TrimSuffix(q, ", ")
	}

	b.statement = q + ";"

	return b
}

func (b *EntityBuilder) InsertEdge() *EntityBuilder {
	panic("implement me")
}

func (b *EntityBuilder) DeleteEdge() *EntityBuilder {
	panic("implement me")
}

func (b *EntityBuilder) UpdateEdge() *EntityBuilder {
	panic("implement me")
}

func (b *EntityBuilder) UpsertEdge(srcVid, dstVid interface{}) *EntityBuilder {
	src, dst, isStr := b.parseSrcAndDstVid(srcVid, dstVid)

	q := ""
	if isStr {
		q = fmt.Sprintf(`UPSERT EDGE ON %s "%s" -> "%s"`, b.name, src, dst)
	} else {
		q = fmt.Sprintf(`UPSERT EDGE ON %s "%s" -> "%s"`, b.name, src, dst)
	}

	if len(b.props) > 0 {
		q += " SET "
		for _, p := range b.props {
			if p.isStr {
				q += fmt.Sprintf(`%s = "%s", `, p.name, p.value)
			} else {
				q += fmt.Sprintf(`%s = %s, `, p.name, p.value)
			}
		}
		q = strings.TrimSuffix(q, ", ")
	}

	b.statement = q + ";"

	return b
}

func (b *EntityBuilder) String() string {
	return b.statement
}

func (b *EntityBuilder) Exec(pool *SessionPool) (*ResultSet, error) {
	return pool.ExecuteAndCheck(b.statement)
}

func (b *EntityBuilder) escapeStrVal(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func (b *EntityBuilder) parseVid(vid interface{}) (string, bool) {
	t := reflect.TypeOf(vid)
	switch t.Kind() {
	case reflect.Int64:
		return strconv.FormatInt(vid.(int64), 10), false
	case reflect.String:
		return vid.(string), true
	default:
		panic("invalid VID type")
	}
}

func (b *EntityBuilder) parseSrcAndDstVid(srcVid, dstVid interface{}) (string, string, bool) {
	tSrc := reflect.TypeOf(srcVid)
	tDst := reflect.TypeOf(dstVid)
	if tSrc.Kind() != tDst.Kind() {
		panic("src and dst VID type must be the same")
	}

	switch tSrc.Kind() {
	case reflect.Int64:
		srcVid := strconv.FormatInt(srcVid.(int64), 10)
		dstVid := strconv.FormatInt(dstVid.(int64), 10)
		return srcVid, dstVid, false
	case reflect.String:
		srcVid := srcVid.(string)
		dstVid := dstVid.(string)
		return srcVid, dstVid, true
	default:
		panic("invalid src and dst VID type")
	}
}
