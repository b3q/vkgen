package schema

import (
	"fmt"

	"github.com/tidwall/gjson"
)

type ObjectDefinition struct {
	Name string
	Expr ObjectExpr
}

type ObjectExpr struct {
	Type        string
	Description *string
	Ref         *ObjectDefinition
	RefTo       *string
	Properties  []ObjectDefinition
	AllOf       []ObjectExpr
	OneOf       []ObjectExpr
	Enum        []interface{}
	EnumNames   []string
	ArrayOf     *ObjectExpr
	IsBaseType  bool
	IsReference bool
	IsAllOf     bool
	IsOneOf     bool
	IsEnum      bool
	//IsArray     bool
}

func ParseObjects(schema []byte) ([]ObjectDefinition, error) {
	var defs []ObjectDefinition
	var err error
	gjson.ParseBytes(schema).Get("definitions").ForEach(func(objName, objData gjson.Result) bool {
		expr, parseErr := parseObjectExpression(objData)
		if parseErr != nil {
			err = parseErr
			return false
		}

		defs = append(defs, ObjectDefinition{
			Name: objName.String(),
			Expr: expr,
		})
		return true
	})
	return defs, err
}

func parseObjectExpression(obj gjson.Result) (ObjectExpr, error) {
	var expr ObjectExpr

	if desc := obj.Get("description"); desc.Exists() {
		d := desc.String()
		expr.Description = &d
	}

	var err error
	if props := obj.Get("properties"); props.Exists() {
		props.ForEach(func(propName, propData gjson.Result) bool {
			propObj, parseErr := parseObjectExpression(propData)
			if parseErr != nil {
				err = parseErr
				return false
			}
			expr.Properties = append(expr.Properties, ObjectDefinition{
				Name: propName.String(),
				Expr: propObj,
			})
			return true
		})
	}

	if err != nil {
		return expr, err
	}

	if ref := obj.Get("$ref"); ref.Exists() {
		s := resolveReferenceName(ref.String())
		expr.RefTo = &s
		expr.IsReference = true
		return expr, nil
	}

	// проверка на allOf перед проверкой существования типа, потому что
	// newsfeed_getSuggestedSources_response
	if allof := obj.Get("allOf"); allof.Exists() && allof.IsArray() {
		for _, item := range allof.Array() {
			itemObjExpr, parseErr := parseObjectExpression(item)
			if parseErr != nil {
				return expr, parseErr
			}

			expr.AllOf = append(expr.AllOf, itemObjExpr)
		}
		expr.IsAllOf = true
		return expr, nil
	}

	typ := obj.Get("type")
	if !typ.Exists() {
		//pp.Println(obj)
		return expr, nil
		//return expr, fmt.Errorf("undefined type")
	}
	expr.Type = typ.String()

	if enum := obj.Get("enum"); enum.Exists() && enum.IsArray() {
		for _, item := range enum.Array() {
			switch typ.String() {
			case "string":
				expr.Enum = append(expr.Enum, item.String())
			case "number":
				expr.Enum = append(expr.Enum, item.Float())
			case "integer":
				expr.Enum = append(expr.Enum, item.Int())
			default:
				return expr, fmt.Errorf("unsupported enum type: %s", typ.String())
			}
		}

		if enumNames := obj.Get("enumNames"); enumNames.Exists() && enumNames.IsArray() {
			for _, name := range enumNames.Array() {
				expr.EnumNames = append(expr.EnumNames, name.String())
			}
		}

		expr.IsEnum = true
		return expr, nil
	}

	switch typ.String() {
	case "integer":
		fallthrough
	case "number":
		fallthrough
	case "string":
		fallthrough
	case "boolean":
		expr.IsBaseType = true
	case "object":
		if oneof := obj.Get("oneOf"); oneof.Exists() && oneof.IsArray() {
			for _, item := range oneof.Array() {
				itemObjExpr, parseErr := parseObjectExpression(item)
				if parseErr != nil {
					return expr, parseErr
				}

				expr.OneOf = append(expr.OneOf, itemObjExpr)
			}
			expr.IsOneOf = true
			return expr, nil
		}
	case "array":
		items := obj.Get("items")
		if !items.Exists() {
			return expr, fmt.Errorf("array must have items field")
		}

		arrayType, parseErr := parseObjectExpression(items)
		if parseErr != nil {
			return expr, parseErr
		}
		expr.IsBaseType = true
		//expr.IsArray = true
		expr.ArrayOf = &arrayType
		return expr, nil
	default:
		//pp.Println("undefined type", obj)
		//panic("unimplemented")
	}

	return expr, nil
}
