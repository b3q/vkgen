package schema

import (
	"strings"

	"github.com/tidwall/gjson"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

type Definition struct {
	Name        *string
	Description *string
	DataType    DataType
	RefTo       *string
	Array       *Array
	Enum        *Enum
	OneOf       *OneOf
	AllOf       *AllOf
	Properties  []*Definition
	IsOptional  bool
}

type DataType string

const (
	UnsupportedType DataType = "unsupported"
	ReferenceType   DataType = "$ref"
	ArrayType       DataType = "array"
	EnumType        DataType = "enum"
	OneofType       DataType = "oneOf"
	AllofType       DataType = "allOf"
	ObjectType      DataType = "object"
	BooleanType     DataType = "boolean"
	IntegerType     DataType = "integer"
	NumberType      DataType = "number"
	StringType      DataType = "string"
)

func getDataType(name string) DataType {
	switch name {
	case "array":
		return ArrayType
	case "boolean":
		return BooleanType
	case "integer":
		return IntegerType
	case "number":
		return NumberType
	case "string":
		return StringType
	case "object":
		return ObjectType
	default:
		return UnsupportedType
	}
}

func (p *Parser) ParseDefinition(name string, definition gjson.Result) *Definition {
	def := new(Definition)
	def.Name = &name
	if ref := definition.Get("$ref"); ref.Exists() {
		referenceName := p.resolveReferenceName(ref.String())
		def.RefTo = &referenceName
		def.DataType = ReferenceType
	}

	if desc := definition.Get("description"); desc.Exists() {
		d := desc.String()
		def.Description = &d
	}

	if allof := definition.Get("allOf"); allof.Exists() {
		a := p.parseAllOf(name, allof)
		def.AllOf = &a
		//DataType: allofType
	}

	if oneof := definition.Get("oneOf"); oneof.Exists() && oneof.IsArray() {
		o := p.parseOneOf(name, oneof)
		def.OneOf = &o
		//DataType: oneofType
	}

	if enum := definition.Get("enum"); enum.Exists() {
		e := p.parseEnum(name, definition)
		def.Enum = &e
		//DataType: oneofType
	}

	if props := definition.Get("properties"); props.Exists() {
		parsedProps := p.ParseProperties(props)
		def.Properties = parsedProps
		//DataType:   objectType
	}

	dtype := definition.Get("type")
	if dtype.Exists() {
		switch dtype.String() {
		case "array":
			arr := p.parseArray(definition)
			def.Array = &arr
			def.DataType = ArrayType
		case "string":
			def.DataType = StringType
		case "integer":
			def.DataType = IntegerType
		case "number":
			def.DataType = NumberType
		case "boolean":
			def.DataType = BooleanType
		case "object":
			def.DataType = ObjectType
		default:
			def.DataType = UnsupportedType
			//panic("unsupported property type " + dtype.String())
		}
	}

	return def
}

func (p *Parser) ParseProperties(properties gjson.Result) []*Definition {
	var props []*Definition
	properties.ForEach(func(pname, pdata gjson.Result) bool {
		pnameString := pname.String()
		def := p.ParseDefinition(pnameString, pdata)
		def.Name = &pnameString
		props = append(props, def)
		return true
	})
	return props
}

type Array struct {
	DataType DataType
	Ref      *Definition
	AllOf    *AllOf
}

func (p *Parser) parseArray(array gjson.Result) Array {
	items := array.Get("items")
	if !items.Exists() {
		panic("array must have items field")
	}

	if ref := items.Get("$ref"); ref.Exists() {
		def := p.resolveReferenceName(ref.String())
		return Array{
			DataType: ReferenceType,
			Ref: &Definition{
				DataType: ReferenceType,
				Name:     &def,
				//RefTo: &def,
			},
		}
	}

	if allof := items.Get("allOf"); allof.Exists() {
		a := p.parseAllOf("", allof)
		return Array{
			DataType: AllofType,
			AllOf:    &a,
		}
	}

	arrayDataType := items.Get("type")
	if !arrayDataType.Exists() {
		panic("undefined array data type")
	}

	if getDataType(arrayDataType.String()) == ArrayType {
		subArray := p.parseArray(items)
		return Array{
			DataType: ArrayType,
			Ref: &Definition{
				DataType: ArrayType,
				Array:    &subArray,
			},
		}
	}

	return Array{
		DataType: getDataType(arrayDataType.String()),
	}
}

type AllOf struct {
	Name           string
	PossibleValues []*Definition
}

func (p *Parser) parseAllOf(name string, allof gjson.Result) AllOf {
	if !allof.IsArray() {
		panic("allof must be json array")
	}

	var defs []*Definition
	for _, item := range allof.Array() {
		if ref := item.Get("$ref"); ref.Exists() {
			def := p.resolveReferenceName(ref.String())
			defs = append(defs, &Definition{
				DataType: ReferenceType,
				Name:     &def,
				//RefTo: &def,
			})
		}

		if props := item.Get("properties"); props.Exists() {
			parsedProps := p.ParseProperties(props)
			defs = append(defs, &Definition{
				Properties: parsedProps,
			})
		}
	}

	return AllOf{name, defs}
}

type Enum struct {
	Name          string
	Description   *string
	DataType      DataType
	StringValues  []string
	NumericValues []int64
	BooleanValues []bool
	ObjectValues  []*Definition
	Names         []string
}

func (p *Parser) parseEnum(name string, enum gjson.Result) Enum {
	enumDataType := enum.Get("type")
	if !enumDataType.Exists() {
		panic("undefined enum data type")
	}

	genum := Enum{
		Name:     name,
		DataType: getDataType(enumDataType.String()),
	}
	if desc := enum.Get("description"); desc.Exists() {
		d := desc.String()
		genum.Description = &d
	}

	//dtype := getDataType(enumDataType.String())
	values := enum.Get("enum")
	if !values.Exists() || !values.IsArray() {
		panic("enum must be array")
	}

	switch enumDataType.String() {
	case "integer":
		fallthrough
	case "number":
		for _, value := range values.Array() {
			genum.NumericValues = append(genum.NumericValues, value.Int())
		}
	case "boolean":
		for _, value := range values.Array() {
			genum.BooleanValues = append(genum.BooleanValues, value.Bool())
		}
	case "string":
		for _, value := range values.Array() {
			genum.StringValues = append(genum.StringValues, value.String())
		}
	case "object":
		for _, value := range values.Array() {
			genum.ObjectValues = append(genum.ObjectValues, p.ParseDefinition(name, value))
		}
	default:
		panic("unsupported enum data type: " + enumDataType.String())
	}

	if enames := enum.Get("enumNames"); enames.Exists() && enames.IsArray() {
		for _, name := range enames.Array() {
			genum.Names = append(genum.Names, name.String())
		}
	}

	return genum
}

type OneOf struct {
	Name           string
	PossibleValues []*Definition
}

func (p *Parser) parseOneOf(name string, oneof gjson.Result) OneOf {
	// o := oneof.Get("oneof")
	// if !o.Exists() || !o.IsArray() {
	// 	panic("oneof must be array")
	// }

	var defs []*Definition
	for _, item := range oneof.Array() {
		if ref := item.Get("$ref"); ref.Exists() {
			def := p.resolveReferenceName(ref.String())
			defs = append(defs, &Definition{
				DataType: ReferenceType,
				Name:     &def,
				//RefTo: &def,
			})
		}

		// ???
		// if typ := item.Get("type"); typ.Exists(){

		// }
	}

	return OneOf{name, defs}
}

func (p *Parser) resolveReferenceName(refpath string) string {
	objectNameIndex := strings.LastIndex(refpath, `/`)
	objectName := refpath[objectNameIndex+1:]
	return objectName
}

// func (p *Parser) resolveReference(refpath string) *Definition {
// 	filenamePrefixIndex := strings.Index(refpath, `/`)
// 	filename := refpath[filenamePrefixIndex+1:]

// 	objectNameIndex := strings.LastIndex(refpath, `/`)
// 	objectName := refpath[objectNameIndex+1:]

// 	gjsonPath := refpath[filenamePrefixIndex+1:]
// 	gjsonPath = strings.ReplaceAll(gjsonPath, `/`, `.`)

// 	var js gjson.Result
// 	switch filename {
// 	case "objects.json":
// 		fallthrough
// 		//js = p.objects.Get(gjsonPath)
// 	default:
// 		panic("unsupported resolving file: " + filename)
// 	}

// 	def := p.ParseDefinition(objectName, js)
// 	def.Name = &objectName
// 	return def
// }
