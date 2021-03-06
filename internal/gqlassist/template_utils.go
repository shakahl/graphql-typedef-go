package gqlassist

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/shurcooL/graphql/ident"

	"github.com/shakahl/gqlassist/internal/utils"
)

func ParseTemplate(name, tpl string) *template.Template {
	funcMap := makeTemplateFuncMap()
	return template.Must(template.New(name).Funcs(sprig.TxtFuncMap()).Funcs(funcMap).Parse(tpl))
}

func makeTemplateFuncMap() template.FuncMap {
	feature := func(field string) bool {
		switch field {
		case "use_integer_enums":
			return true
		default:
			return false
		}
	}
	isExcluded := func(s string) bool {
		return false
	}
	clean := func(s string) string {
		return strings.Join(strings.Fields(s), " ")
	}
	quote := func(s string) string {
		return strconv.Quote(s)
	}
	formatDescription := func(s string) string {
		s = strings.ToLower(s[0:1]) + s[1:]
		if !strings.HasSuffix(s, ".") {
			s += "."
		}
		return s
	}
	join := func(elems []string, sep string) string {
		return strings.Join(elems, sep)
	}
	sortByName := func(types []interface{}) []interface{} {
		sort.Slice(types, func(i, j int) bool {
			ni := types[i].(map[string]interface{})["name"].(string)
			nj := types[j].(map[string]interface{})["name"].(string)
			return ni < nj
		})
		return types
	}
	sortByNameRev := func(types []interface{}) []interface{} {
		sort.Slice(types, func(i, j int) bool {
			ni := types[i].(map[string]interface{})["name"].(string)
			nj := types[j].(map[string]interface{})["name"].(string)
			return ni > nj
		})
		return types
	}
	filterBy := func(field string, kind string, types []interface{}) []interface{} {
		var filtered = []interface{}{}
		for _, t := range types {
			if val, ok := t.(map[string]interface{})[field]; ok && val.(string) == kind && !isExcluded(val.(string)) {
				filtered = append(filtered, t)
			}
			continue
		}
		return filtered
	}
	extractField := func(field string, types []interface{}) []string {
		var values = []string{}
		for _, t := range types {
			if val, ok := t.(map[string]interface{})[field]; ok {
				values = append(values, val.(string))
			}
			continue
		}
		return values
	}
	ucFirst := func(s string) string {
		return utils.StringUpperCaseFirst(s)
	}
	lcFirst := func(s string) string {
		return utils.StringLowerCaseFirst(s)
	}
	identifier := func(s string) string {
		s = strings.TrimLeft(s, "_")
		return ident.ParseScreamingSnakeCase(s).ToMixedCaps()
	}
	scalarIdentifier := func(s string) string {
		return ident.ParseScreamingSnakeCase(s).ToMixedCaps()
	}
	enumIdentifierValueSuffix := func(s string) string {
		// return s
		// return ident.ParseScreamingSnakeCase(name).ToMixedCaps()
		return "_" + strings.ToUpper(s)
	}
	enumTypeString := func(s string) string {
		return identifier(s)
	}
	enumAllValuesIdentifier := func(s string) string {
		return enumTypeString(s) + "__LIST"
	}
	scalarTypeString := func(gqltype string) string {
		tnorm := strings.ToLower(gqltype)
		switch tnorm {
		case "order_by":
			return "string"
		case "time":
			return "time.Time"
		case "timestamp":
			return "time.Time"
		case "timestamptz":
			return "time.Time"
		case "date":
			return "time.Time"
		case "datetime":
			return "time.Time"
		case "uuid":
			return "uuid.UUID"
		case "id":
			return "string"
		case "string":
			return "string"
		case "boolean":
			return "bool"
		case "float":
			return "float64"
		case "integer":
			return "time.Time"
		case "int":
			return "int32"
		case "json":
			return "map[string]interface{}"
		case "jsonb":
			return "map[string]interface{}"
		default:
			return scalarIdentifier(gqltype)
		}
	}

	// typeString returns a string representation of GraphQL type t.
	var typeString func(t map[string]interface{}) string
	typeString = func(t map[string]interface{}) string {
		switch t["kind"] {
		case "SCALAR":
			return "*" + scalarTypeString(t["name"].(string))
		case "NON_NULL":
			s := typeString(t["ofType"].(map[string]interface{}))
			if !strings.HasPrefix(s, "*") {
				panic(fmt.Errorf("nullable type %q doesn't begin with '*'", s))
			}
			return s[1:] // Strip "*" from nullable type to make it non-null.
		case "LIST":
			return "*[]" + typeString(t["ofType"].(map[string]interface{}))
		case "ENUM":
			return "*" + enumTypeString(t["name"].(string))
		case "INPUT_OBJECT", "OBJECT":
			break
		default:
			break
		}
		return "*" + identifier(t["name"].(string))
	}

	inputObjects := func(types []interface{}) []string {
		var names []string
		for _, t := range types {
			t := t.(map[string]interface{})
			if t["kind"].(string) != "INPUT_OBJECT" {
				continue
			}
			names = append(names, t["name"].(string))
		}
		sort.Strings(names)
		return names
	}

	objects := func(types []interface{}) []string {
		var names []string
		for _, t := range types {
			t := t.(map[string]interface{})
			if t["kind"].(string) != "OBJECT" {
				continue
			}
			names = append(names, t["name"].(string))
		}
		sort.Strings(names)
		return names
	}
	first := func(x int, a interface{}) bool {
		return x == 0
	}
	last := func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	}
	toUpper := func(s string) string {
		return strings.ToUpper(s)
	}
	toLower := func(s string) string {
		return strings.ToLower(s)
	}
	isGraphQLMeta := func(s string) bool {
		return false
		// return strings.HasPrefix(s, "__")
	}

	// settings["integer_enums"] = true

	return template.FuncMap{
		"ucFirst":                   ucFirst,
		"lcFirst":                   lcFirst,
		"feature":                   feature,
		"first":                     first,
		"last":                      last,
		"toUpper":                   toUpper,
		"toLower":                   toLower,
		"isGraphQLMeta":             isGraphQLMeta,
		"isExcluded":                isExcluded,
		"quote":                     quote,
		"join":                      join,
		"sortByName":                sortByName,
		"sortByNameRev":             sortByNameRev,
		"filterBy":                  filterBy,
		"extractField":              extractField,
		"inputObjects":              inputObjects,
		"objects":                   objects,
		"identifier":                identifier,
		"type":                      typeString,
		"enumType":                  enumTypeString,
		"enumIdentifierValueSuffix": enumIdentifierValueSuffix,
		"enumAllValuesIdentifier":   enumAllValuesIdentifier,
		"scalarIdentifier":          scalarIdentifier,
		"scalarType":                scalarTypeString,
		"clean":                     clean,
		"formatDescription":         formatDescription,
	}
}
