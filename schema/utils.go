package schema

import "strings"

func resolveReferenceName(refpath string) string {
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
