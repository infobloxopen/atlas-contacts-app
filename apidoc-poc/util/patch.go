package main

import (
	"encoding/json"
	"github.com/go-openapi/spec"

	"flag"
	"fmt"
	"os"
	"strings"

	"io/ioutil"
)

var (
	swaggerFile string
	sw          spec.Swagger
	seenRefs    = map[string]bool{}
)

func main() {
	b, err := ioutil.ReadFile(swaggerFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file %q: %v", swaggerFile, err)
		os.Exit(1)
	}

	if err := json.Unmarshal(b, &sw); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing file %q: %v", swaggerFile, err)
		os.Exit(1)
	}

	// Fix collection operators and IDs and gather refs along Paths.

	refs := []spec.Ref{}
	fixedPaths := map[string]spec.PathItem{}

	for pn, pi := range sw.Paths.Paths {
		pnElements := []string{}
		for _, v := range strings.Split(pn, "/") {
			if strings.HasSuffix(v, "id.resource_id}") {
				pnElements = append(pnElements, "{id}")
			} else {
				pnElements = append(pnElements, v)
			}
		}
		pn := strings.Join(pnElements, "/")
		for on, op := range pathItemAsMap(pi) {
			if op == nil {
				continue
			}
			fixedParams := []spec.Parameter{}
			for _, param := range op.Parameters {

				// Fix Collection Operators
				if strings.HasPrefix(param.Description, "atlas.api.") {
					switch strings.TrimPrefix(param.Description, "atlas.api.") {
					case "filtering":
						fixedParams = append(fixedParams, *(spec.QueryParam("_filter")).WithDescription(
							"A collection of response resources can be filtered by a logical expression " +
								"string that includes JSON tag references to values in each resource, literal " +
								"values, and logical operators. If a resource does not have the specified tag, " +
								"its value is assumed to be null." +
								"\n" +
								"Literal values include numbers (integer and floating-point), and quoted " +
								"(both single- or double-quoted) literal strings, and “null”. The following " +
								"operators are commonly used in filter expressions:\n" +
								" | Op | Description |\n" +
								" | -- | ----------- |\n" +
								" | == | Equal |\n" +
								" | != | Not Equal |\n" +
								" | > | Greater Than |\n" +
								" |  >= | Greater Than or Equal To |\n" +
								" | < | Less Than |\n" +
								" | <= | Less Than or Equal To |\n" +
								" | and | Logical AND |\n" +
								" | ~ | Matches Regex |\n" +
								" | !~ | Does Not Match Regex |\n" +
								" | or | Logical OR |\n" +
								" | not | Logical NOT |\n" +
								" | () | Groupping Operators |\n",
						))
					case "sorting":
						fixedParams = append(fixedParams, *(spec.QueryParam("_order_by")).WithDescription(
							"A collection of response resources can be sorted by their JSON tags. For a " +
								"“flat” resource, the tag name is straightforward. If sorting is allowed on " +
								"non-flat hierarchical resources, the service should implement a qualified " +
								"naming scheme such as dot-qualification to reference data down the hierarchy. " +
								"If a resource does not have the specified tag, its value is assumed to be null.)" +
								"\n\n" +
								"Specify this parameter as a comma-separated list of JSON tag names. The sort " +
								"direction can be specified by a suffix separated by whitespace before the tag " +
								"name. The suffix “asc” sorts the data in ascending order. The suffix “desc” " +
								"sorts the data in descending order. If no suffix is specified the data is sorted " +
								"in ascending order.",
						))
					case "field_selection":
						fixedParams = append(fixedParams, *(spec.QueryParam("_fields")).WithDescription(

							"A collection of response resources can be transformed by specifying a set of JSON " +
								"tags to be returned. For a “flat” resource, the tag name is straightforward. If " +
								"field selection is allowed on non-flat hierarchical resources, the service should " +
								"implement a qualified naming scheme such as dot-qualification to reference data down " +
								"the hierarchy. If a resource does not have the specified tag, the tag does not appear " +
								"in the output resource." +
								"\n\n" +
								"Specify this parameter as a comma-separated list of JSON tag names.",
						))
					case "paging":
						fixedParams = append(
							fixedParams,
							*(spec.QueryParam("_offset")).WithDescription(
								"The integer index (zero-origin) of the offset into a collection of resources. " +
									"If omitted or null the value is assumed to be “0”.",
							),
							*(spec.QueryParam("_limit")).WithDescription(
								"The integer number of resources to be returned in the response. The " +
									"service may impose maximum value. If omitted the service may impose " +
									"a default value.",
							),
							*(spec.QueryParam("_page_token")).WithDescription(
								"The service-defined string used to identify a page of resources. A null value " +
									"indicates the first page.",
							),
						)
					// Skip ID
					default:
					}
					// Replace resource_id with id
				} else if strings.HasSuffix(param.Name, "id.resource_id") {
					param.Name = "id"
					fixedParams = append(fixedParams, param)
				} else {
					// Gather ref in body.
					if param.In == "body" && param.Schema != nil {
						refs = append(refs, param.Schema.Ref)
					}
					fixedParams = append(fixedParams, param)
				}
			}
			op.Parameters = fixedParams

			// Wrap responses
			if op.Responses.StatusCodeResponses != nil {
				rsp := op.Responses.StatusCodeResponses[200]
				if !isNilRef(rsp.Schema.Ref) {
					s, _, err := rsp.Schema.Ref.GetPointer().Get(sw)
					if err != nil {
						panic(err)
					}

					schema, ok := s.(spec.Schema)
					if !ok {
						panic("cannot cast interface to spec.Schema")
					}

					if schema.Properties == nil {
						schema.Properties = map[string]spec.Schema{}
					}

					schema.Properties["success"] = *(&spec.Schema{}).WithProperties(
						map[string]spec.Schema{
							"status":  *spec.StringProperty().WithExample(opToStatus(on)),
							"code":    *spec.Int32Property().WithExample(opToCode(on)),
							"message": *spec.StringProperty().WithExample("<response message>"),
						})

					sw.Definitions[trim(rsp.Schema.Ref)] = schema

					refs = append(refs, rsp.Schema.Ref)

					delete(op.Responses.StatusCodeResponses, 200)

					op.Responses.StatusCodeResponses[opToCode(on)] = rsp
				}
			}
		}

		fixedPaths[pn] = pi
	}

	sw.Paths.Paths = fixedPaths

	// Break recursive rules introduced by many-to-many.
	for _, r := range refs {
		seenRefs[trim(r)] = true
		s, _, err := r.GetPointer().Get(sw)
		if err != nil {
			panic(err)
		}
		checkRecursion(s.(spec.Schema), r, []string{})
	}

	// Cleanup unused definitions.
	for dn, v := range sw.Definitions {
		// hidden definitions should become explicit.
		if strings.HasPrefix(dn, "_") {
			sw.Definitions[strings.TrimPrefix(dn, "_")] = v
			delete(sw.Definitions, dn)
			seenRefs[dn] = true
		}

		if seenRefs[dn] == false {
			delete(sw.Definitions, dn)
		}
	}

	bOut, err := json.MarshalIndent(sw, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling result: %v", err)
		os.Exit(1)
	}

	fmt.Printf("%s", bOut)
}

func getPropRef(p spec.Schema) spec.Ref {
	if len(p.Type) == 1 && p.Type[0] == "array" {
		return p.Items.Schema.Ref
	}

	return p.Ref
}

func setPropRef(p *spec.Schema, r spec.Ref) {
	if len(p.Type) == 1 && p.Type[0] == "array" {
		p.Items.Schema.Ref = r
	} else {
		p.Ref = r
	}
}

func checkRecursion(s spec.Schema, r spec.Ref, path []string) spec.Ref {
	var newRefLength int
	var newRefName string

	var newProps = map[string]spec.Schema{}

	npath := path[:]
	npath = append(npath, trim(r))

	newProps = map[string]spec.Schema{}
	for np, p := range s.Properties {
		if p.Description == "atlas.api.identifier" {
			p.Description = "The resource identifier."
			if np == "id" {
				p.ReadOnly = true
			}
		}

		newProps[np] = p

		sr := getPropRef(p)

		if isNilRef(sr) {
			continue
		}

		for i, prefs := range npath {
			if trim(sr) == prefs {
				delete(newProps, np)
				if newRefLength < len(npath)-i {
					newRefName = strings.Join(reverse(npath[i:]), "_In_")
					newRefLength = len(npath) - i
				}
			}
		}

		if _, ok := newProps[np]; !ok {
			continue
		}

		ss, _, _ := sr.GetPointer().Get(sw)
		if _, ok := ss.(spec.Schema); !ok {
			continue
		}

		nr := checkRecursion(ss.(spec.Schema), sr, npath)

		if trim(nr) != trim(sr) {
			if newRefName == "" {
				newRefName = strings.TrimPrefix(trim(nr), trim(sr)+"_In_")
			}

			delete(newProps, np)

			if len(p.Type) == 1 && p.Type[0] == "array" {
				newProps[np] = *spec.ArrayProperty(spec.RefProperty(nr.String()))
			} else {
				newProps[np] = *spec.RefProperty(nr.String())
			}
		} else {
			seenRefs[trim(sr)] = true
		}
	}

	if newRefName != "" {
		seenRefs[newRefName] = true
		// underscore hides definitions from following along recursive path.
		sw.Definitions["_"+newRefName] = *(&spec.Schema{}).WithProperties(newProps)
		return spec.MustCreateRef("#/definitions/" + newRefName)
	} else {
		s.Properties = newProps
		sw.Definitions[trim(r)] = s
	}

	return r
}

func trim(r spec.Ref) string {
	return strings.TrimPrefix(r.String(), "#/definitions/")
}

func isNilRef(r spec.Ref) bool {
	return r.String() == ""
}

func reverse(s []string) []string {
	news := make([]string, len(s))
	for i := len(s) - 1; i >= 0; i-- {
		news[i] = s[len(s)-1-i]
	}

	return news
}

func pathItemAsMap(pi spec.PathItem) map[string]*spec.Operation {
	return map[string]*spec.Operation{
		"GET":    pi.Get,
		"POST":   pi.Post,
		"PUT":    pi.Put,
		"DELETE": pi.Delete,
	}
}

func opToCode(on string) int {
	return map[string]int{
		"GET":    200,
		"POST":   201,
		"PUT":    202,
		"DELETE": 203,
	}[on]
}

func opToStatus(on string) string {
	return map[string]string{
		"GET":    "OK",
		"POST":   "CREATED",
		"PUT":    "UPDATED",
		"DELETE": "DELETED",
	}[on]
}

func init() {
	var input = flag.String("input", "", "input swagger file")

	flag.Parse()

	if *input == "" {
		panic("missing input file")
	}

	swaggerFile = *input
}
