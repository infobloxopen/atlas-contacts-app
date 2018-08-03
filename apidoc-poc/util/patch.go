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
				if rsp.Schema != nil {
					if rsp.Schema.Properties == nil {
						rsp.Schema.Properties = map[string]spec.Schema{}
					}
					rsp.Schema.Properties["success"] = *(&spec.Schema{}).WithProperties(
						map[string]spec.Schema{
							"status":  *spec.StringProperty().WithExample(opToStatus(on)),
							"code":    *spec.Int32Property().WithExample(opToCode(on)),
							"message": *spec.StringProperty().WithExample("<response message>"),
						})

					if !isNilRef(rsp.Schema.Ref) {
						refs = append(refs, rsp.Schema.Ref)
						rsp.Schema.Properties["results"] = spec.Schema{SchemaProps: spec.SchemaProps{Ref: rsp.Schema.Ref}}
						rsp.Schema.Ref = spec.Ref{}
					}

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
		recurseCheck(s.(spec.Schema), r, []string{})
	}

	// Cleanup unused definitions.
	for dn, _ := range sw.Definitions {
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

func recurseCheck(s spec.Schema, r spec.Ref, path []string) spec.Ref {
	var patchedName string

	npath := path[:]
	newProps := map[string]spec.Schema{}

	if s.Properties != nil {
		npath = append(npath, trim(r))
		for np, p := range s.Properties {
			isArray := len(p.Type) == 1 && p.Type[0] == "array"

			var sref spec.Ref
			if isArray {
				sref = p.Items.Schema.Ref
			} else {
				sref = p.Ref
			}
			var hasRecursion bool
			if !isNilRef(sref) {
				srefName := trim(sref)
				for i, np := range npath {
					if np == srefName {
						patchedName = strings.Join(reverse(npath[i:]), "In")
						hasRecursion = true
					}
				}

				if !hasRecursion {
					s, _, err := sref.GetPointer().Get(sw)
					if err != nil {
						panic(err)
					}

					// If we cannot resolve, it is a new ref.
					if _, ok := s.(spec.Schema); !ok {
						s = sw.Definitions[srefName]
					}

					newRef := recurseCheck(s.(spec.Schema), sref, npath)
					seenRefs[trim(newRef)] = true

					if isArray {
						p.Items.Schema.Ref = newRef
					} else {
						p.Ref = newRef
					}
				}

			}

			// Skip over recursive property.
			if hasRecursion {
				continue
			}

			if p.Description == "atlas.api.identifier" {
				p.Description = "The resource identifier."
				p.ReadOnly = true
			}

			newProps[np] = p
		}
	}

	// Recursion has been detected and Properties should be patched with new properties.
	if len(s.Properties) != len(newProps) {
		sw.Definitions[patchedName] = *(&spec.Schema{}).WithProperties(newProps)
		return spec.MustCreateRef("#/definitions/" + patchedName)
		// No recursion occured, just update Properties in case if any changes to fields were
		// performed during recursion check.
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
