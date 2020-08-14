package schema

import "strings"

func resolveReferenceName(refpath string) string {
	objectNameIndex := strings.LastIndex(refpath, `/`)
	objectName := refpath[objectNameIndex+1:]
	return objectName
}
