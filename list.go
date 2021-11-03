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

// ConcurrentList goroutine safe for list
type ConcurrentList struct {
	list   list.List
	rwLock sync.RWMutex
}

func (q *ConcurrentList) Len() int { return q.list.Len() }

func (q *ConcurrentList) Front() *list.Element {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	return q.list.Front()
}

func (q *ConcurrentList) Back() *list.Element {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	return q.list.Back()
}

func (q *ConcurrentList) Remove(e *list.Element) interface{} {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	return q.list.Remove(e)
}

func (q *ConcurrentList) RemoveByValue(v interface{}) interface{} {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	for ele := q.list.Front(); ele != nil; ele = ele.Next() {
		if ele.Value == v {
			return q.list.Remove(ele)
		}
	}
	return nil
}

func (q *ConcurrentList) PushFront(v interface{}) *list.Element {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	return q.list.PushFront(v)
}

func (q *ConcurrentList) PushBack(v interface{}) *list.Element {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	return q.list.PushBack(v)
}
func (q *ConcurrentList) MoveToFront(e *list.Element) {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	q.list.MoveToFront(e)
}

func (q *ConcurrentList) MoveToBack(e *list.Element) {
	q.rwLock.Lock()
	defer q.rwLock.Unlock()
	q.list.MoveToBack(e)
}
