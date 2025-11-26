package pg

import (
	"fmt"
	"strings"
)

// identifierNameSql returns the object name ready to be used in a sql query as an object name (e.a. Select * from %s).
func identifierNameSQL(objectName string) (escaped string) {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(objectName, `"`, `""`))
}

// stringValueSql uses proper quoting for values in SQL queries.
// func stringValueSql(stringValue string) (escaped string) {
// 	return fmt.Sprintf("'%s'", strings.Replace(stringValue, "'", "''", -1))
// }

// connectStringValue uses proper quoting for connect string values.
func connectStringValue(objectName string) (escaped string) {
	return fmt.Sprintf(`'%s'`, strings.ReplaceAll(objectName, `'`, `\'`))
}
