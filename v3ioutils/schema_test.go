package v3ioutils

import (
	"fmt"
	"testing"
)

const schemaTst = `
{"fields":[{"name":"age","type":"long","nullable":true},{"name":"job","type":"string","nullable":true},{"name":"marital","type":"string","nullable":true},{"name":"education","type":"string","nullable":true},{"name":"default","type":"string","nullable":true},{"name":"balance","type":"long","nullable":true},{"name":"housing","type":"string","nullable":true},{"name":"loan","type":"string","nullable":true},{"name":"contact","type":"string","nullable":true},{"name":"day","type":"long","nullable":true},{"name":"month","type":"string","nullable":true},{"name":"duration","type":"long","nullable":true},{"name":"campaign","type":"long","nullable":true},{"name":"pdays","type":"long","nullable":true},{"name":"previous","type":"long","nullable":true},{"name":"poutcome","type":"string","nullable":true},{"name":"y","type":"string","nullable":true},{"name":"id","type":"long","nullable":false}],"key":"id","hashingBucketNum":0}
`

func TestNewSchema(t *testing.T) {
	schema, err := SchemaFromJson([]byte(schemaTst))
	fmt.Println(err, schema)
}
