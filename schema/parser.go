package schema

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

type Parser struct {
	objects gjson.Result
}

func NewParser(objectsSchema []byte) *Parser {
	return &Parser{
		objects: gjson.ParseBytes(objectsSchema),
	}
}

func (p *Parser) resolveReference(refpath string) (ObjectDefinition, error) {
	filenamePrefixIndex := strings.Index(refpath, `/`)
	filename := refpath[:filenamePrefixIndex-1]

	objectNameIndex := strings.LastIndex(refpath, `/`)
	objectName := refpath[objectNameIndex+1:]

	gjsonPath := refpath[filenamePrefixIndex+1:]
	gjsonPath = strings.ReplaceAll(gjsonPath, `/`, `.`)

	var js gjson.Result
	switch filename {
	case "objects.json":
		js = p.objects.Get(gjsonPath)
	case "responses.json":
		return ObjectDefinition{
			Name: objectName,
		}, nil
	default:
		fmt.Println(refpath)
		panic("unsupported resolving file: " + filename)
	}

	expr, err := p.parseObjectExpression(js)
	return ObjectDefinition{
		Name: objectName,
		Expr: expr,
	}, err
}
