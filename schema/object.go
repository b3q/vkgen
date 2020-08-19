package schema

import (
	"fmt"
	"reflect"

	"github.com/tidwall/gjson"
)

type ObjectDefinition struct {
	Name string
	Expr ObjectExpr
}

func (def ObjectDefinition) Equal(another ObjectDefinition) bool {
	if def.Name != another.Name {
		return false
	}

	return def.Expr.EqualType(another.Expr)
}

type ObjectType byte

const (
	Base ObjectType = 1 << iota
	Ref
	AllOf
	OneOf
	Enum
	Array
)

type ObjectExpr struct {
	objtype     ObjectType
	Type        string
	Description *string
	Ref         func() (ObjectDefinition, error)
	Properties  []ObjectDefinition
	AllOf       []ObjectExpr
	OneOf       []ObjectExpr
	Enum        []interface{}
	EnumNames   []string
	ArrayOf     *ObjectExpr
}

func (expr ObjectExpr) Is(typ ObjectType) bool {
	return expr.objtype&typ != 0
}

func (expr ObjectExpr) EqualType(another ObjectExpr) bool {

	if expr.Type != another.Type {
		return false
	}

	if len(expr.Properties) != len(another.Properties) {
		return false
	}

	for i := 0; i < len(expr.Properties); i++ {
		def1, def2 := expr.Properties[i], another.Properties[i]
		if !def1.Equal(def2) {
			return false
		}
	}

	if expr.Is(Ref) {
		r1, err := expr.Ref()
		if err != nil {
			panic(err)
		}

		r2, err := another.Ref()
		if err != nil {
			panic(err)
		}

		if !r1.Equal(r2) {
			return false
		}
	}

	if expr.Is(AllOf) {
		if len(expr.AllOf) != len(another.AllOf) {
			return false
		}

		for i := 0; i < len(expr.AllOf); i++ {
			item1, item2 := expr.AllOf[i], another.AllOf[i]
			if !item1.EqualType(item2) {
				return false
			}
		}
	}

	if expr.Is(OneOf) {
		if len(expr.OneOf) != len(another.OneOf) {
			return false
		}

		for i := 0; i < len(expr.OneOf); i++ {
			item1, item2 := expr.OneOf[i], another.OneOf[i]
			if !item1.EqualType(item2) {
				return false
			}
		}
	}

	if expr.Is(Enum) {
		if len(expr.Enum) != len(another.Enum) ||
			len(expr.EnumNames) != len(another.EnumNames) {
			return false
		}

		for i := 0; i < len(expr.Enum); i++ {
			e1, e2 := expr.Enum[i], another.Enum[i]
			if !reflect.DeepEqual(e1, e2) {
				return false
			}
		}

		for i := 0; i < len(expr.EnumNames); i++ {
			n1, n2 := expr.EnumNames[i], another.EnumNames[i]
			if n1 != n2 {
				return false
			}
		}
	}

	if expr.Is(Array) {
		if !expr.ArrayOf.EqualType(*another.ArrayOf) {
			return false
		}
	}

	return true
}

func (p *Parser) ParseObjects(schema []byte) ([]ObjectDefinition, error) {
	var defs []ObjectDefinition
	var err error
	gjson.ParseBytes(schema).Get("definitions").ForEach(func(objName, objData gjson.Result) bool {
		expr, parseErr := p.parseObjectExpression(objData)
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

func (p *Parser) parseObjectExpression(obj gjson.Result) (ObjectExpr, error) {
	var expr ObjectExpr

	if desc := obj.Get("description"); desc.Exists() {
		d := desc.String()
		expr.Description = &d
	}

	var err error
	if props := obj.Get("properties"); props.Exists() {
		props.ForEach(func(propName, propData gjson.Result) bool {
			propObj, parseErr := p.parseObjectExpression(propData)
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
		refFn := func() (ObjectDefinition, error) {
			return p.resolveReference(ref.String())
		}
		expr.Ref = refFn
		expr.objtype = Ref
		return expr, nil
	}

	// проверка на allOf перед проверкой существования типа, потому что
	// newsfeed_getSuggestedSources_response
	if allof := obj.Get("allOf"); allof.Exists() && allof.IsArray() {
		for _, item := range allof.Array() {
			itemObjExpr, parseErr := p.parseObjectExpression(item)
			if parseErr != nil {
				return expr, parseErr
			}

			expr.AllOf = append(expr.AllOf, itemObjExpr)
		}
		expr.objtype = AllOf
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

		expr.objtype = Enum
		return expr, nil
	}

	switch typ.String() {
	case "integer", "number", "string", "boolean":
		expr.objtype = Base
	case "object":
		if oneof := obj.Get("oneOf"); oneof.Exists() && oneof.IsArray() {
			for _, item := range oneof.Array() {
				itemObjExpr, parseErr := p.parseObjectExpression(item)
				if parseErr != nil {
					return expr, parseErr
				}

				expr.OneOf = append(expr.OneOf, itemObjExpr)
			}
			expr.objtype = OneOf
			return expr, nil
		}
	case "array":
		items := obj.Get("items")
		if !items.Exists() {
			return expr, fmt.Errorf("array must have items field")
		}

		arrayType, parseErr := p.parseObjectExpression(items)
		if parseErr != nil {
			return expr, parseErr
		}
		expr.objtype = Array
		//expr.IsArray = true
		expr.ArrayOf = &arrayType
		return expr, nil
	default:
		//pp.Println("undefined type", obj)
		//panic("unimplemented")
	}

	return expr, nil
}
