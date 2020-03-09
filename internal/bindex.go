package internal

import "sync"

type set = map[interface{}]struct{}

type BIndex struct {
	mu sync.RWMutex
	userToTagSet map[interface{}]set // map[key] map[key2]struct{}
	tagToUserSet map[interface{}]set // map[key2] map[key]struct{}
}

func (bi *BIndex) AddUserTag(user, tag interface{}) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	// 给用户user添加标签tag

	// 正向索引
	tagSet, ok := bi.userToTagSet[user]
	if !ok {
		// user不存在，创建新tagSet并加入
		tagSet = make(set)
		bi.userToTagSet[user] = tagSet
	}
	tagSet[tag] = struct{}{}

	// 反向索引
	userSet, ok := bi.tagToUserSet[tag]
	if !ok {
		// tag不存在，创建新userSet并加入
		userSet = make(set)
		bi.tagToUserSet[tag] = userSet
	}
	userSet[user] = struct{}{}
}

func (bi *BIndex) RemoveUserTag(user, tag interface{}) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	// 正向索引
	tagSet, ok := bi.userToTagSet[user]
	if !ok {
		// user不存在
		return
	}
	delete(tagSet, tag)
	if len(tagSet) == 0 {
		delete(bi.userToTagSet, user)
	}

	// 反向索引
	userSet, ok := bi.tagToUserSet[tag]
	if !ok {
		// tag不存在，创建新userSet并加入
		return
	}
	delete(userSet, user)
	if len(userSet) == 0 {
		delete(bi.tagToUserSet, tag)
	}
}

func (bi *BIndex) Tags(user interface{}, output interface{}) bool {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	// 正向索引
	tagSet, ok := bi.userToTagSet[user]
	if !ok {
		// user不存在
		return false
	}

	switch output.(type) {
	case *[]interface{}:
		tags := output.(*[]interface{})
		if size := len(tagSet); cap(*tags) < size {
			*tags = make([]interface{}, 0, size)
		} else {
			*tags = (*tags)[:0]
		}
		for tag, _ := range tagSet {
			*tags = append(*tags, tag)
		}
		return true
	default:
		return false
	}
}

func (bi *BIndex) Users(tag interface{}, output interface{}) bool {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	// 反向索引
	userSet, ok := bi.tagToUserSet[tag]
	if !ok {
		// tag不存在
		return false
	}

	switch output.(type) {
	case *[]interface{}:
		users := output.(*[]interface{})
		if size := len(userSet); cap(*users) < size {
			*users = make([]interface{}, 0, size)
		} else {
			*users = (*users)[:0]
		}
		for user, _ := range userSet {
			*users = append(*users, user)
		}
		return true
	default:
		return false
	}
}

func (bi *BIndex) AllTags(output interface{}) bool {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	switch output.(type) {
	case *[]interface{}:
		tags := output.(*[]interface{})
		if size := len(bi.tagToUserSet); cap(*tags) < size {
			*tags = make([]interface{}, 0, size)
		} else {
			*tags = (*tags)[:0]
		}
		for tag, _ := range bi.tagToUserSet {
			*tags = append(*tags, tag)
		}
		return true
	default:
		return false
	}
}

func (bi *BIndex) AllUsers(output interface{}) bool {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	switch output.(type) {
	case *[]interface{}:
		users := output.(*[]interface{})
		if size := len(bi.userToTagSet); cap(*users) < size {
			*users = make([]interface{}, 0, size)
		} else {
			*users = (*users)[:0]
		}
		for user, _ := range bi.userToTagSet {
			*users = append(*users, user)
		}
		return true
	default:
		return false
	}
}

func (bi *BIndex) RemoveUser(user interface{}) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	// 正向索引
	tagSet, ok := bi.userToTagSet[user]
	if !ok {
		// user不存在
		return
	}

	for tag, _ := range tagSet {
		userSet := bi.tagToUserSet[tag]
		delete(userSet, user)
		if len(userSet) == 0 {
			delete(bi.tagToUserSet, tag)
		}
	}

	delete(bi.userToTagSet, user)
}

func (bi *BIndex) RemoveTag(tag interface{}) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	// 反向索引
	userSet, ok := bi.tagToUserSet[tag]
	if !ok {
		// user不存在
		return
	}

	for user, _ := range userSet {
		tagSet := bi.userToTagSet[user]
		delete(tagSet, tag)
		if len(tagSet) == 0 {
			delete(bi.userToTagSet, user)
		}
	}

	delete(bi.tagToUserSet, tag)
}

func NewBIndex() *BIndex {
	bi := &BIndex{
		userToTagSet: make(map[interface{}]set),
		tagToUserSet: make(map[interface{}]set),
	}
	return bi
}
