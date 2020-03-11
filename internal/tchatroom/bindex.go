package tchatroom

import (
	"sync"
)

type set = map[interface{}]struct{}

type BIndex struct {
	mu           sync.RWMutex
	userToTagSet map[interface{}]set // map[key] map[key2]struct{}
	tagToUserSet map[interface{}]set // map[key2] map[key]struct{}
}

func (bi *BIndex) AddUserTag(user interface{}, tags ...interface{}) {
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

	for _, tag := range tags {
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
}

func (bi *BIndex) RemoveUserTag(user interface{}, tags ...interface{}) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	// 正向索引
	tagSet, ok := bi.userToTagSet[user]
	if !ok {
		// user不存在
		return
	}

	for _, tag := range tags {
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

func (bi *BIndex) SelectUsers(tags []interface{}, output interface{}) {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	var users *[]interface{}
	switch output.(type) {
	case *[]interface{}:
		users = output.(*[]interface{})
		*users = (*users)[:0]
	default:
		return
	}

	for _, tag := range tags {
		// 反向索引
		userSet, ok := bi.tagToUserSet[tag]
		if ok {
			for user, _ := range userSet {
				*users = append(*users, user)
			}
		}
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

type Index struct {
	mu           sync.RWMutex
	userToTagSet map[interface{}]set // map[key] map[key2]struct{}
	tagToUser    map[interface{}]interface{}
}

func (i *Index) AddUserTag(user interface{}, tags ...interface{}) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// 给用户user添加标签tag

	// 正向索引
	tagSet, ok := i.userToTagSet[user]
	if !ok {
		// user不存在，创建新tagSet并加入
		tagSet = make(set)
		i.userToTagSet[user] = tagSet
	}

	for _, tag := range tags {
		tagSet[tag] = struct{}{}
		if i.tagToUser != nil {
			i.tagToUser[tag] = user
		}
	}
}

func (i *Index) RemoveUserTag(user interface{}, tags ...interface{}) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// 正向索引
	tagSet, ok := i.userToTagSet[user]
	if !ok {
		// user不存在
		return
	}

	for _, tag := range tags {
		delete(tagSet, tag)
		if len(tagSet) == 0 {
			delete(i.userToTagSet, user)
		}
		if i.tagToUser != nil {
			delete(i.tagToUser, tag)
		}
	}
}

func (i *Index) Tags(user interface{}, output interface{}) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// 正向索引
	tagSet, ok := i.userToTagSet[user]
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

func (i *Index) SelectTags(users []interface{}, output interface{}) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	var tags *[]interface{}
	switch output.(type) {
	case *[]interface{}:
		tags = output.(*[]interface{})
		*tags = (*tags)[:0]
	default:
		return
	}

	for _, user := range users {
		tagSet, ok := i.userToTagSet[user]
		if ok {
			for tag, _ := range tagSet {
				*tags = append(*tags, tag)
			}
		}
	}
}

func (i *Index) AllUsers(output interface{}) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	switch output.(type) {
	case *[]interface{}:
		users := output.(*[]interface{})
		if size := len(i.userToTagSet); cap(*users) < size {
			*users = make([]interface{}, 0, size)
		} else {
			*users = (*users)[:0]
		}
		for user, _ := range i.userToTagSet {
			*users = append(*users, user)
		}
		return true
	default:
		return false
	}
}

func (i *Index) RemoveUser(user interface{}) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.tagToUser != nil {
		// 正向索引
		tagSet, ok := i.userToTagSet[user]
		if !ok {
			// user不存在
			return
		}

		for tag, _ := range tagSet {
			delete(i.tagToUser, tag)
		}
	}
	delete(i.userToTagSet, user)
}

func (i *Index) RemoveTag(tag interface{}) {
	if i.tagToUser == nil {
		return
	}

	if user, ok := i.tagToUser[tag]; ok {
		tagSet := i.userToTagSet[user]
		delete(tagSet, tag)
		if len(tagSet) == 0 {
			delete(i.userToTagSet, user)
		}
	}
}

func (i *Index) User(tag interface{}) (interface{}, bool) {
	if i.tagToUser == nil {
		return nil, false
	}

	if user, ok := i.tagToUser[tag]; ok {
		return user, true
	} else {
		return nil, false
	}
}

func NewIndex(reverse bool) *Index {
	i := &Index{
		userToTagSet: make(map[interface{}]set),
	}
	if reverse {
		i.tagToUser = make(map[interface{}]interface{})
	}
	return i
}
