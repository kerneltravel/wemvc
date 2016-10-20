package wemvc

// CtxItems the context item struct
type CtxItems struct {
	items map[string]interface{}
}

// Get the the data from the context item map
func (ci *CtxItems) Get(key string) interface{} {
	return ci.items[key]
}

// Set add data to the context item map
func (ci *CtxItems) Set(key string, data interface{}) {
	ci.items[key] = data
}

// Clear clear the context item map
func (ci *CtxItems) Clear() {
	ci.items = nil
}

// Delete delete data from the context item map and return the deleted data
func (ci *CtxItems) Delete(key string) interface{} {
	data, ok := ci.items[key]
	if ok {
		delete(ci.items, key)
	}
	return data
}
