package ajson

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// JSON from example https://goessner.net/articles/JsonPath/index.html#e3
var jsonPathTestData = []byte(`{ "store": {
    "book": [ 
      { "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      { "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      },
      { "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99
      },
      { "category": "fiction",
        "author": "J. R. R. Tolkien",
        "title": "The Lord of the Rings",
        "isbn": "0-395-19395-8",
        "price": 22.99
      }
    ],
    "bicycle": {
      "color": "red",
      "price": 19.95
    }
  }
}`)

func fullPath(array []*Node) string {
	return sliceString(Paths(array))
}

func sliceString(array []string) string {
	return "[" + strings.Join(array, ", ") + "]"
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestJsonPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
		wantErr  bool
	}{
		{name: "root", path: "$", expected: "[$]"},
		{name: "roots", path: "$.", expected: "[$]"},
		{name: "all objects", path: "$..", expected: "[$, $['store'], $['store']['bicycle'], $['store']['book'], $['store']['book'][0], $['store']['book'][1], $['store']['book'][2], $['store']['book'][3]]"},
		{name: "only children", path: "$.*", expected: "[$['store']]"},

		{name: "by key", path: "$.store.bicycle", expected: "[$['store']['bicycle']]"},
		{name: "all key 1", path: "$..bicycle", expected: "[$['store']['bicycle']]"},
		{name: "all key 2", path: "$..price", expected: "[$['store']['bicycle']['price'], $['store']['book'][0]['price'], $['store']['book'][1]['price'], $['store']['book'][2]['price'], $['store']['book'][3]['price']]"},
		{name: "all key bracket", path: "$..['price']", expected: "[$['store']['bicycle']['price'], $['store']['book'][0]['price'], $['store']['book'][1]['price'], $['store']['book'][2]['price'], $['store']['book'][3]['price']]"},
		{name: "all fields", path: "$['store']['book'][1].*", expected: "[$['store']['book'][1]['author'], $['store']['book'][1]['category'], $['store']['book'][1]['price'], $['store']['book'][1]['title']]"},

		{name: "union fields", path: "$['store']['book'][2]['author','price','title']", expected: "[$['store']['book'][2]['author'], $['store']['book'][2]['price'], $['store']['book'][2]['title']]"},
		{name: "union indexes", path: "$['store']['book'][1,2]", expected: "[$['store']['book'][1], $['store']['book'][2]]"},
		{name: "union indexes calculate", path: "$['store']['book'][-2,(@.length-1)]", expected: "[$['store']['book'][2], $['store']['book'][3]]"},
		{name: "union indexes position", path: "$['store']['book'][-1,-3]", expected: "[$['store']['book'][3], $['store']['book'][1]]"},

		{name: "slices 1", path: "$..[1:4]", expected: "[$['store']['book'][1], $['store']['book'][2], $['store']['book'][3]]"},
		{name: "slices 2", path: "$..[1:4:]", expected: "[$['store']['book'][1], $['store']['book'][2], $['store']['book'][3]]"},
		{name: "slices 3", path: "$..[1:4:1]", expected: "[$['store']['book'][1], $['store']['book'][2], $['store']['book'][3]]"},
		{name: "slices 4", path: "$..[1:]", expected: "[$['store']['book'][1], $['store']['book'][2], $['store']['book'][3]]"},
		{name: "slices 5", path: "$..[:2]", expected: "[$['store']['book'][0], $['store']['book'][1]]"},
		{name: "slices 6", path: "$..[:4:2]", expected: "[$['store']['book'][0], $['store']['book'][2]]"},
		{name: "slices 7", path: "$..[:4:]", expected: "[$['store']['book'][0], $['store']['book'][1], $['store']['book'][2], $['store']['book'][3]]"},
		{name: "slices 8", path: "$..[::]", expected: "[$['store']['book'][0], $['store']['book'][1], $['store']['book'][2], $['store']['book'][3]]"},
		{name: "slices 9", path: "$['store']['book'][1:4:2]", expected: "[$['store']['book'][1], $['store']['book'][3]]"},
		{name: "slices 10", path: "$['store']['book'][1:4:3]", expected: "[$['store']['book'][1]]"},
		{name: "slices 11", path: "$['store']['book'][:-1]", expected: "[$['store']['book'][0], $['store']['book'][1], $['store']['book'][2]]"},
		{name: "slices 12", path: "$['store']['book'][-1:]", expected: "[$['store']['book'][3]]"},
		{name: "slices 13", path: "$..[::-1]", expected: "[$['store']['book'][3], $['store']['book'][2], $['store']['book'][1], $['store']['book'][0]]"},
		{name: "slices 14", path: "$..[::-2]", expected: "[$['store']['book'][3], $['store']['book'][1]]"},
		{name: "slices 15", path: "$..[::2]", expected: "[$['store']['book'][0], $['store']['book'][2]]"},
		{name: "slices 16", path: "$..[-3:(@.length)]", expected: "[$['store']['book'][1], $['store']['book'][2], $['store']['book'][3]]"},
		{name: "slices 17", path: "$..[(-3*@.length + 1):(@.length - 1)]", expected: "[$['store']['book'][1], $['store']['book'][2]]"},
		{name: "slices 18", path: "$..[(foobar(@.length))::]", wantErr: true},
		{name: "slices 19", path: "$..[::0]", wantErr: true},
		{name: "slices 20", path: "$..[:(1/0):]", wantErr: true},
		{name: "slices 21", path: "$..[:(1/2):]", wantErr: true},
		{name: "slices 22", path: "$..[:0.5:]", wantErr: true},

		{name: "calculated 1", path: "$['store']['book'][(@.length-1)]", expected: "[$['store']['book'][3]]"},
		{name: "calculated 2", path: "$['store']['book'][(3.5 - 3/2)]", expected: "[$['store']['book'][2]]"},
		{name: "calculated 3", path: "$..book[?(@.isbn)]", expected: "[$['store']['book'][2], $['store']['book'][3]]"},
		{name: "calculated 4", path: "$..[?(@.price < factorial(3) + 3)]", expected: "[$['store']['book'][0], $['store']['book'][2]]"},
		{name: "calculated 5", path: "$..[(1/0)]", wantErr: true},

		{name: "$.store.book[*].author", path: "$.store.book[*].author", expected: "[$['store']['book'][0]['author'], $['store']['book'][1]['author'], $['store']['book'][2]['author'], $['store']['book'][3]['author']]"},
		{name: "$..author", path: "$..author", expected: "[$['store']['book'][0]['author'], $['store']['book'][1]['author'], $['store']['book'][2]['author'], $['store']['book'][3]['author']]"},
		{name: "$.store..price", path: "$.store..price", expected: "[$['store']['bicycle']['price'], $['store']['book'][0]['price'], $['store']['book'][1]['price'], $['store']['book'][2]['price'], $['store']['book'][3]['price']]"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := JSONPath(jsonPathTestData, test.path)
			if (err != nil) != test.wantErr {
				t.Errorf("JSONPath() error = %v, wantErr %v. got = %v", err, test.wantErr, result)
				return
			}
			if test.wantErr {
				return
			}
			if fullPath(result) != test.expected {
				t.Errorf("Error on JsonPath(json, %s) as %s: path doesn't match\nExpected: %s\nActual:   %s", test.path, test.name, test.expected, fullPath(result))
			}
		})
	}
}

func TestJsonPath_value(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected interface{}
	}{
		{name: "length", path: "$['store']['book'].length", expected: float64(4)},
		{name: "price", path: "$['store']['book'][?(@.price + 0.05 == 9)].price", expected: float64(8.95)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := JSONPath(jsonPathTestData, test.path)
			if err != nil {
				t.Errorf("Error on JsonPath(json, %s) as %s: %s", test.path, test.name, err.Error())
			} else if len(result) != 1 {
				t.Errorf("Error on JsonPath(json, %s) as %s: path to long, expected only value\nActual: %s", test.path, test.name, fullPath(result))
			} else {
				val, err := result[0].Value()
				if err != nil {
					t.Errorf("Error on JsonPath(json, %s): error %s", test.path, err.Error())
				} else {
					switch {
					case result[0].IsNumeric():
						if val.(float64) != test.expected.(float64) {
							t.Errorf("Error on JsonPath(json, %s): value doesn't match\nExpected: %v\nActual:   %v", test.path, test.expected, val)
						}
					case result[0].IsString():
						if val.(string) != test.expected.(string) {
							t.Errorf("Error on JsonPath(json, %s): value doesn't match\nExpected: %v\nActual:   %v", test.path, test.expected, val)
						}
					case result[0].IsBool():
						if val.(bool) != test.expected.(bool) {
							t.Errorf("Error on JsonPath(json, %s): value doesn't match\nExpected: %v\nActual:   %v", test.path, test.expected, val)
						}
					default:
						t.Errorf("Error on JsonPath(json, %s): unsupported type found", test.path)
					}
				}
			}
		})
	}
}

func TestParseJSONPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{name: "root", path: "$", expected: []string{"$"}},
		{name: "roots", path: "$.", expected: []string{"$"}},
		{name: "all objects", path: "$..", expected: []string{"$", ".."}},
		{name: "only children", path: "$.*", expected: []string{"$", "*"}},
		{name: "all objects children", path: "$..*", expected: []string{"$", "..", "*"}},
		{name: "path dot:simple", path: "$.root.element", expected: []string{"$", "root", "element"}},
		{name: "path dot:combined", path: "$.root.*.element", expected: []string{"$", "root", "*", "element"}},
		{name: "path bracket:simple", path: "$['root']['element']", expected: []string{"$", "'root'", "'element'"}},
		{name: "path bracket:combined", path: "$['root'][*]['element']", expected: []string{"$", "'root'", "*", "'element'"}},
		{name: "path bracket:int", path: "$['store']['book'][0]['title']", expected: []string{"$", "'store'", "'book'", "0", "'title'"}},
		{name: "path combined:simple", path: "$['root'].*['element']", expected: []string{"$", "'root'", "*", "'element'"}},
		{name: "path combined:dotted", path: "$.['root'].*.['element']", expected: []string{"$", "'root'", "*", "'element'"}},
		{name: "path combined:dotted small", path: "$['root'].*.['element']", expected: []string{"$", "'root'", "*", "'element'"}},
		{name: "phoneNumbers", path: "$.phoneNumbers[*].type", expected: []string{"$", "phoneNumbers", "*", "type"}},
		{name: "filtered", path: "$.store.book[?(@.price < 10)].title", expected: []string{"$", "store", "book", "?(@.price < 10)", "title"}},
		{name: "formula", path: "$..phoneNumbers..('ty' + 'pe')", expected: []string{"$", "..", "phoneNumbers", "..", "('ty' + 'pe')"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseJSONPath(test.path)
			if err != nil {
				t.Errorf("Error on parseJsonPath(json, %s) as %s: %s", test.path, test.name, err.Error())
			} else if !sliceEqual(result, test.expected) {
				t.Errorf("Error on parseJsonPath(%s) as %s: path doesn't match\nExpected: %s\nActual: %s", test.path, test.name, sliceString(test.expected), sliceString(result))
			}
		})
	}
}

// Test suites from cburgmer/json-path-comparison
func TestJSONPath_suite(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		path     string
		expected []interface{}
		wantErr  bool
	}{
		{
			name:     "Bracket notation with double quotes",
			input:    `{"key": "value"}`,
			path:     `$["key"]`,
			expected: []interface{}{"value"}, // ["value"]
		},
		{
			name:     "Filter expression with bracket notation",
			input:    `[{"key": 0}, {"key": 42}, {"key": -1}, {"key": 41}, {"key": 43}, {"key": 42.0001}, {"key": 41.9999}, {"key": 100}, {"some": "value"}]`,
			path:     `$[?(@['key']==42)]`,
			expected: []interface{}{map[string]interface{}{"key": float64(42)}}, // [{"key": 42}]
		},
		{
			name:     "Filter expression with equals string with dot literal",
			input:    `[{"key": "some"}, {"key": "value"}, {"key": "some.value"}]`,
			path:     `$[?(@.key=="some.value")]`,
			expected: []interface{}{map[string]interface{}{"key": "some.value"}}, // [{"key": "some.value"}]
		},
		{
			name:     "Array slice with negative step only",
			input:    `["first", "second", "third", "forth", "fifth"]`,
			path:     `$[::-2]`,
			expected: []interface{}{"fifth", "third", "first"}, // ["fifth", "third", "first"]
		},
		{
			name:     "Filter expression with bracket notation with -1",
			input:    `[[2, 3], ["a"], [0, 2], [2]]`,
			path:     `$[?(@[-1]==2)]`,
			expected: []interface{}{[]interface{}{float64(0), float64(2)}, []interface{}{float64(2)}}, // [[0, 2], [2]]
		},
		{
			name:     "Filter expression with bracket notation with number",
			input:    `[["a", "b"], ["x", "y"]]`,
			path:     `$[?(@[1]=='b')]`,
			expected: []interface{}{[]interface{}{"a", "b"}}, // [["a", "b"]]
		},
		{
			name:     "Filter expression with equals string with current object literal",
			input:    `[{"key": "some"}, {"key": "value"}, {"key": "hi@example.com"}]`,
			path:     `$[?(@.key=="hi@example.com")]`,
			expected: []interface{}{map[string]interface{}{"key": "hi@example.com"}}, // [{"key": "hi@example.com"}]
		},
		// 		{
		// 			name:     "Filter expression with negation and equals",
		// 			input:    `[
		//     {"key": 0},
		//     {"key": 42},
		//     {"key": -1},
		//     {"key": 41},
		//     {"key": 43},
		//     {"key": 42.0001},
		//     {"key": 41.9999},
		//     {"key": 100},
		//     {"key": "43"},
		//     {"key": "42"},
		//     {"key": "41"},
		//     {"key": "value"},
		//     {"some": "value"}
		// ]`,
		// 			path:     `$[?(!(@.key==42))]`,
		// 			expected: []interface{}{
		// 				map[string]interface{}{"key": float64(0)},
		// 				map[string]interface{}{"key": float64(-1)},
		// 				map[string]interface{}{"key": float64(41)},
		// 				map[string]interface{}{"key": float64(43)},
		// 				map[string]interface{}{"key": float64(42.0001)},
		// 				map[string]interface{}{"key": float64(41.9999)},
		// 				map[string]interface{}{"key": float64(100)},
		// 				map[string]interface{}{"key": "43"},
		// 				map[string]interface{}{"key": "42"},
		// 				map[string]interface{}{"key": "41"},
		// 				map[string]interface{}{"key": "value"},
		// 				map[string]interface{}{"some": "value"},
		// 			},
		// 		},
		{
			name:     "Filter expression with bracket notation with number on object",
			input:    `{"1": ["a", "b"], "2": ["x", "y"]}`,
			path:     `$[?(@[1]=='b')]`,
			expected: []interface{}{[]interface{}{"a", "b"}}, // [["a", "b"]]
		},
		// {
		// 	name:     "Dot notation with single quotes and dot",
		// 	input:    `{"some.key": 42, "some": {"key": "value"}}`,
		// 	path:     `$.'some.key'`,
		// 	expected: []interface{}{float64(42)}, // [42]
		// },
		{
			name:    "Array slice with step 0",
			input:   `["first", "second", "third", "forth", "fifth"]`,
			path:    `$[0:3:0]`,
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nodes, err := JSONPath([]byte(test.input), test.path)
			if (err != nil) != test.wantErr {
				t.Errorf("JSONPath() error = %v, wantErr %v. got = %v", err, test.wantErr, nodes)
				return
			}
			if test.wantErr {
				return
			}

			results := make([]interface{}, 0)
			for _, node := range nodes {
				value, err := node.Unpack()
				if err != nil {
					t.Errorf("node.Unpack(): unexpected error: %v", err)
					return
				}
				results = append(results, value)
			}

			if !reflect.DeepEqual(results, test.expected) {
				t.Errorf("JSONPath(): wrong result:\nExpected: %#+v\nActual:   %#+v", test.expected, results)
			}
		})
	}
}

func ExampleJSONPath() {
	json := []byte(`{ "store": {
    "book": [ 
      { "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      { "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      },
      { "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99
      },
      { "category": "fiction",
        "author": "J. R. R. Tolkien",
        "title": "The Lord of the Rings",
        "isbn": "0-395-19395-8",
        "price": 22.99
      }
    ],
    "bicycle": {
      "color": "red",
      "price": 19.95
    }
  }
}`)
	authors, err := JSONPath(json, "$.store.book[*].author")
	if err != nil {
		panic(err)
	}
	for _, author := range authors {
		fmt.Println(author.MustString())
	}
	// Output:
	// Nigel Rees
	// Evelyn Waugh
	// Herman Melville
	// J. R. R. Tolkien
}

func ExampleJSONPath_array() {
	json := []byte(`{ "store": {
    "book": [ 
      { "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      { "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      },
      { "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99
      },
      { "category": "fiction",
        "author": "J. R. R. Tolkien",
        "title": "The Lord of the Rings",
        "isbn": "0-395-19395-8",
        "price": 22.99
      }
    ],
    "bicycle": {
      "color": "red",
      "price": 19.95
    }
  }
}`)
	authors, err := JSONPath(json, "$.store.book[*].author")
	if err != nil {
		panic(err)
	}
	result, err := Marshal(ArrayNode("", authors))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(result))
	// Output:
	// ["Nigel Rees","Evelyn Waugh","Herman Melville","J. R. R. Tolkien"]
}

func ExampleEval() {
	json := []byte(`{ "store": {
    "book": [ 
      { "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      { "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      },
      { "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99
      },
      { "category": "fiction",
        "author": "J. R. R. Tolkien",
        "title": "The Lord of the Rings",
        "isbn": "0-395-19395-8",
        "price": 22.99
      }
    ],
    "bicycle": [
      {
        "color": "red",
        "price": 19.95
      }
    ]
  }
}`)
	root, err := Unmarshal(json)
	if err != nil {
		panic(err)
	}
	result, err := Eval(root, "avg($..price)")
	if err != nil {
		panic(err)
	}
	fmt.Print(result.MustNumeric())
	// Output:
	// 14.774000000000001
}

func TestEval(t *testing.T) {
	json := []byte(`{ "store": {
    "book": [ 
      { "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      { "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      },
      { "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99
      },
      { "category": "fiction",
        "author": "J. R. R. Tolkien",
        "title": "The Lord of the Rings",
        "isbn": "0-395-19395-8",
        "price": 22.99
      }
    ],
    "bicycle": [
      {
        "color": "red",
        "price": 19.95
      }
    ]
  }
}`)
	tests := []struct {
		name     string
		root     *Node
		eval     string
		expected *Node
		wantErr  bool
	}{
		{
			name:     "avg($..price)",
			root:     Must(Unmarshal(json)),
			eval:     "avg($..price)",
			expected: NumericNode("", 14.774000000000001),
			wantErr:  false,
		},
		{
			name:     "avg($..price)",
			root:     Must(Unmarshal(json)),
			eval:     "avg()",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "avg($..price)",
			root:     Must(Unmarshal(json)),
			eval:     "($..price+)",
			expected: nil,
			wantErr:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Eval(test.root, test.eval)
			if (err != nil) != test.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v. got = %v", err, test.wantErr, result)
				return
			}
			if test.wantErr {
				return
			}
			if result == nil {
				t.Errorf("Eval() result in nil")
				return
			}

			if ok, err := result.Eq(test.expected); !ok {
				t.Errorf("result.Eq(): wrong result:\nExpected: %#+v\nActual: %#+v", test.expected, result.value.Load())
			} else if err != nil {
				t.Errorf("result.Eq() error = %v", err)
			}

		})
	}
}

func BenchmarkJSONPath_all_prices(b *testing.B) {
	var err error
	for i := 0; i < b.N; i++ {
		_, err = JSONPath(jsonPathTestData, "$.store..price")
		if err != nil {
			b.Error()
		}
	}
}
