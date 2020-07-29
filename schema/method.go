package schema

import "github.com/tidwall/gjson"

type MethodDefinition struct {
	Name        string
	Description *string
	AccessType  []string
	Parameters  []MethodParam
	Responses   []ObjectDefinition
}

type MethodParam struct {
	Name string
	ObjectExpr
}

func ParseMethods(schema []byte) ([]MethodDefinition, error) {
	var defs []MethodDefinition
	for _, method := range gjson.ParseBytes(schema).Get("methods").Array() {
		def, err := parseMethod(method)
		if err != nil {
			return nil, err
		}
		defs = append(defs, def)
	}

	return defs, nil
}

func parseMethod(method gjson.Result) (MethodDefinition, error) {
	var mdef MethodDefinition
	mdef.Name = method.Get("name").String()
	if desc := method.Get("description"); desc.Exists() {
		d := desc.String()
		mdef.Description = &d
	}
	var access []string
	for _, acctype := range method.Get("access_token_type").Array() {
		access = append(access, acctype.String())
	}
	mdef.AccessType = access

	for _, param := range method.Get("parameters").Array() {
		paramExpr, err := parseObjectExpression(param)
		if err != nil {
			return mdef, err
		}
		mdef.Parameters = append(mdef.Parameters, MethodParam{
			Name:       param.Get("name").String(),
			ObjectExpr: paramExpr,
		})
	}

	var err error
	method.Get("responses").ForEach(func(respName, respData gjson.Result) bool {
		expr, parseErr := parseObjectExpression(respData)
		if parseErr != nil {
			err = parseErr
			return false
		}
		mdef.Responses = append(mdef.Responses, ObjectDefinition{
			Name: respName.String(),
			Expr: expr,
		})
		return true
	})

	return mdef, err
}
