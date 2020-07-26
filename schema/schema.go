package schema

import (
	"io/ioutil"
	"net/http"

	"github.com/tidwall/gjson"
)

type SchemaType string

const (
	MethodsSchema   SchemaType = "methods.json"
	ObjectsSchema   SchemaType = "objects.json"
	ResponsesSchema SchemaType = "responses.json"
	UnknownSchema   SchemaType = "unknown"
	repoMasterURL              = "https://github.com/VKCOM/vk-api-schema/blob/master/"
)

func DetectSchemaType(val gjson.Result) SchemaType {
	if m := val.Get("methods"); m.Exists() && m.IsArray() {
		return MethodsSchema
	}

	if t := val.Get("title"); t.Exists() {
		if t.String() == "objects" {
			return ObjectsSchema
		}

		if t.String() == "responses" {
			return ResponsesSchema
		}
	}

	return UnknownSchema
}

func downloadSchemeFromURL(url string) (gjson.Result, error) {
	resp, err := http.Get(url)
	if err != nil {
		return gjson.Result{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.ParseBytes(body), nil
}
