package codemanager

type JsonObject map[string]interface{}

func (parent JsonObject) GetArray(key string) []interface{} {
	return parent[key].([]interface{})
}

func (parent JsonObject) GetObject(key string) JsonObject {
	return JsonObject(parent[key].(map[string]interface{}))
}
