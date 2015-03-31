package schemas

import (
	"encoding/json"
	"github.com/mtabini/gojsonschema"
)

var Schemas = map[string]*gojsonschema.JsonSchemaDocument{}

func resolveReferencesSlice(schema []interface{}) []interface{} {
	for key, value := range schema {
		switch value.(type) {
		case map[string]interface{}:
			schema[key] = resolveReferencesMap(value.(map[string]interface{}))

		case []interface{}:
			schema[key] = resolveReferencesSlice(value.([]interface{}))
		}
	}

	return schema
}

func resolveReferencesMap(schema map[string]interface{}) map[string]interface{} {
	ref, hasRef := schema["$ref"]
	if hasRef {
		res, _ := resolveSchema(ref.(string))
		for key, value := range schema {
			if key != "$ref" {
				res[key] = value
			}
		}
		return res
	}

	for key, value := range schema {
		switch value.(type) {
		case map[string]interface{}:
			schema[key] = resolveReferencesMap(value.(map[string]interface{}))

		case []interface{}:
			schema[key] = resolveReferencesSlice(value.([]interface{}))
		}
	}

	return schema
}

func resolveSchema(name string) (res map[string]interface{}, raw []byte) {
	raw, err := Asset(name)

	if err != nil {
		panic(err.Error())
	}

	err = json.Unmarshal(raw, &res)

	if err != nil {
		panic(err.Error())
	}

	resolveReferencesMap(res)

	raw, err = json.Marshal(res)

	if err != nil {
		panic(err)
	}

	return
}

func loadSchema(name string) *gojsonschema.JsonSchemaDocument {
	schemaMap, rawSchema := resolveSchema(name)

	schema, err := gojsonschema.NewJsonSchemaDocument(schemaMap)

	if err != nil {
		println(">>>>>>>>>>>>>>>>>>>>>>>>>>", name)
		jj, _ := json.MarshalIndent(rawSchema, "", "  ")
		println(string(jj))
		panic(err.Error())
	}

	return schema
}

func init() {
	Schemas["$add"] = loadSchema("json/add.json")
}
