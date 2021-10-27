/* Copyright (c) 2021 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package nebula_go

import (
	"container/list"
	"sync"
)

// Queue goroutine safe for list
type Queue struct {
	list   list.List
	rwLock sync.RWMutex
}

func (q *Queue) Len() int { return q.list.Len() }

func (q *Queue) Front() *list.Element {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	return q.list.Front()
}

func (q *Queue) Back() *list.Element {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	return q.list.Back()
}

func (q *Queue) Remove(e *list.Element) interface{} {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	return q.list.Remove(e)
}

func (q *Queue) RemoveByValue(v interface{}) interface{} {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	for ele := q.list.Front(); ele != nil; ele = ele.Next() {
		if ele.Value == v {
			return q.list.Remove(ele)
		}
	}
	return nil
}

func (q *Queue) PushFront(v interface{}) *list.Element {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	return q.list.PushFront(v)
}

func (q *Queue) PushBack(v interface{}) *list.Element {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	return q.list.PushBack(v)
}
func (q *Queue) MoveToFront(e *list.Element) {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	q.list.MoveToFront(e)
}

func (q *Queue) MoveToBack(e *list.Element) {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	q.list.MoveToBack(e)
}
