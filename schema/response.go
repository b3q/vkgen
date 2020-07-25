package schema

import (
	"fmt"

	"github.com/tidwall/gjson"
)

type ResponseDefinition struct {
	Name string
	Expr ResponseExpr
}

type ResponseExpr struct {
	ObjectExpr
	Required []string
}

func ParseResponses(responses gjson.Result) ([]ResponseDefinition, error) {
	var defs []ResponseDefinition
	var err error
	responses.Get("definitions").ForEach(func(respName, respData gjson.Result) bool {
		expr, parseErr := ParseResponseExpression(respData)
		if parseErr != nil {
			err = parseErr
			return false
		}

		defs = append(defs, ResponseDefinition{
			Name: respName.String(),
			Expr: expr,
		})
		return true
	})
	return defs, err
}

func ParseResponseExpression(resp gjson.Result) (ResponseExpr, error) {
	var expr ResponseExpr
	r := resp.Get("properties.response")
	if !r.Exists() {
		return expr, fmt.Errorf("properties.response field does not exists")
	}

	objExpr, err := ParseObjectExpression(r)
	if err != nil {
		return expr, err
	}

	expr.ObjectExpr = objExpr
	for _, req := range r.Get("required").Array() {
		expr.Required = append(expr.Required, req.String())
	}
	return expr, nil
}
