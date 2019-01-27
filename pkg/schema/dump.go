package schema

import (
	"encoding/json"

	"github.com/alecthomas/jsonschema"
)

// DumpSchema return the schema as a JSON marshalled buffer
func DumpSchema(v interface{}) ([]byte, error) {
	reflector := &jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true, // include the required fields block
		ExpandedStruct:             true, // remove the top level struct joining the schema together
	}
	actualSchema := reflector.Reflect(v)
	return json.MarshalIndent(actualSchema, "", "  ")
}
