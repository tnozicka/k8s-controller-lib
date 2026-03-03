package assets

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func TestMarshalYAML(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name        string
		input       any
		expected    string
		expectedErr error
	}{
		{
			name:        "simple string value",
			input:       "hello",
			expected:    "hello",
			expectedErr: nil,
		},

		{
			name:        "map",
			input:       map[string]string{"foo": "bar"},
			expected:    "foo: bar",
			expectedErr: nil,
		},

		{
			name: "struct",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "alice", Age: 30,
			},
			expected:    "age: 30\nname: alice",
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := MarshalYAML(tc.input)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}

			if got != tc.expected {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestIndent(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name       string
		spaceCount int
		input      string
		expected   string
	}{
		{
			name:       "single line",
			spaceCount: 4,
			input:      "hello",
			expected:   "    hello",
		},

		{
			name:       "multi-line",
			spaceCount: 2,
			input:      "line1\nline2\nline3",
			expected:   "  line1\n  line2\n  line3",
		},

		{
			name:       "0 spaces",
			spaceCount: 0,
			input:      "hello\nworld",
			expected:   "hello\nworld",
		},

		{
			name:       "empty string",
			spaceCount: 4,
			input:      "",
			expected:   "    ",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Indent(tc.spaceCount, tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestNIndent(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name       string
		spaceCount int
		input      string
		expected   string
	}{
		{
			name:       "prepends newline then indents",
			spaceCount: 2,
			input:      "hello",
			expected:   "\n  hello",
		},

		{
			name:       "multi-line",
			spaceCount: 3,
			input:      "a\nb",
			expected:   "\n   a\n   b",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NIndent(tc.spaceCount, tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestIndentNext(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name       string
		spaceCount int
		input      string
		expected   string
	}{
		{
			name:       "single line no newline unchanged",
			spaceCount: 4,
			input:      "hello",
			expected:   "hello",
		},

		{
			name:       "two lines second indented",
			spaceCount: 2,
			input:      "first\nsecond\nthird",
			expected:   "first\n  second\n  third",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := IndentNext(tc.spaceCount, tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestRepeat(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		input    string
		count    int
		expected string
	}{
		{
			name:     "n=3",
			input:    "ab",
			count:    3,
			expected: "ababab",
		},

		{
			name:     "n=0",
			input:    "ab",
			count:    0,
			expected: "",
		},

		{
			name:     "empty string",
			input:    "",
			count:    5,
			expected: "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Repeat(tc.input, tc.count)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestToBytes(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		input    string
		expected []byte
	}{
		{
			name:     "known input",
			input:    "hello",
			expected: []byte("hello"),
		},

		{
			name:     "empty",
			input:    "",
			expected: []byte(""),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ToBytes(tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestToBase64(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "known input",
			input:    []byte("hello"),
			expected: base64.StdEncoding.EncodeToString([]byte("hello")),
		},

		{
			name:     "empty",
			input:    []byte{},
			expected: "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ToBase64(tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestIsTrue(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		input    *bool
		expected bool
	}{
		{
			name:     "nil returns false",
			input:    nil,
			expected: false,
		},

		{
			name:     "ptr to true returns true",
			input:    new(true),
			expected: true,
		},

		{
			name:     "ptr to false returns false",
			input:    new(false),
			expected: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := IsTrue(tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestMakeMap(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name        string
		args        []any
		expected    map[any]any
		expectedErr error
	}{
		{
			name:        "even args",
			args:        []any{"a", 1, "b", 2},
			expected:    map[any]any{"a": 1, "b": 2},
			expectedErr: nil,
		},

		{
			name:        "odd args returns error",
			args:        []any{"a", 1, "b"},
			expected:    nil,
			expectedErr: fmt.Errorf("map length 3 isn't divisible into tuples"),
		},

		{
			name:        "zero args returns empty map",
			args:        []any{},
			expected:    map[any]any{},
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := MakeMap(tc.args...)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}

			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	t.Parallel()

	missingKeyTmpl := template.Must(template.New("test").Parse("Hello, {{.Missing}}!"))
	missingKeyTmpl.Option("missingkey=error")
	var missingKeyBuf bytes.Buffer
	missingKeyExecErr := missingKeyTmpl.Execute(&missingKeyBuf, map[string]string{"Name": "World"})

	tt := []struct {
		name        string
		tmpl        string
		inputs      any
		expected    string
		expectedErr error
	}{
		{
			name:        "valid template",
			tmpl:        "Hello, {{.Name}}!",
			inputs:      map[string]string{"Name": "World"},
			expected:    "Hello, World!",
			expectedErr: nil,
		},

		{
			name:        "missing key returns error",
			tmpl:        "Hello, {{.Missing}}!",
			inputs:      map[string]string{"Name": "World"},
			expected:    "",
			expectedErr: fmt.Errorf("can't execute template %q: %w", "test", missingKeyExecErr),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpl := template.Must(template.New("test").Parse(tc.tmpl))
			got, err := RenderTemplate(tmpl, tc.inputs)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}

			gotStr := string(got)
			if gotStr != tc.expected {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, gotStr))
			}
		})
	}
}

func TestDecode(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	cf := serializer.NewCodecFactory(scheme)
	decoder := cf.UniversalDeserializer()

	validConfigMapYAML := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n  namespace: default\ndata:\n  key: value\n"

	_, invalidYAMLErr := Decode[*corev1.ConfigMap]([]byte("this is not valid yaml for k8s"), decoder)
	_, castErr := Decode[*corev1.ConfigMap]([]byte("apiVersion: v1\nkind: Secret\nmetadata:\n  name: test\n"), decoder)

	tt := []struct {
		name              string
		data              []byte
		expectedName      string
		expectedNamespace string
		expectedDataKey   string
		expectedErr       error
	}{
		{
			name:              "valid K8s YAML returns typed object",
			data:              []byte(validConfigMapYAML),
			expectedName:      "test",
			expectedNamespace: "default",
			expectedDataKey:   "value",
			expectedErr:       nil,
		},

		{
			name:              "invalid YAML returns error",
			data:              []byte("this is not valid yaml for k8s"),
			expectedName:      "",
			expectedNamespace: "",
			expectedDataKey:   "",
			expectedErr:       invalidYAMLErr,
		},

		{
			name:              "valid YAML but wrong type returns cast error",
			data:              []byte("apiVersion: v1\nkind: Secret\nmetadata:\n  name: test\n"),
			expectedName:      "",
			expectedNamespace: "",
			expectedDataKey:   "",
			expectedErr:       castErr,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			obj, err := Decode[*corev1.ConfigMap](tc.data, decoder)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}
			if err != nil {
				return
			}
			if obj.Name != tc.expectedName {
				t.Errorf("expected and got name differ:\n%s", cmp.Diff(tc.expectedName, obj.Name))
			}
			if obj.Namespace != tc.expectedNamespace {
				t.Errorf("expected and got namespace differ:\n%s", cmp.Diff(tc.expectedNamespace, obj.Namespace))
			}
			if obj.Data["key"] != tc.expectedDataKey {
				t.Errorf("expected and got data differ:\n%s", cmp.Diff(tc.expectedDataKey, obj.Data["key"]))
			}
		})
	}
}

func TestRenderAndDecode_ConfigMap(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	cf := serializer.NewCodecFactory(scheme)
	decoder := cf.UniversalDeserializer()

	validConfigMapYAML := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n  namespace: default\ndata:\n  key: value\n"

	t.Run("valid ConfigMap template returns object", func(t *testing.T) {
		t.Parallel()

		tmpl := template.Must(template.New("cm").Parse(validConfigMapYAML))
		obj, rendered, err := RenderAndDecode[*corev1.ConfigMap](tmpl, nil, decoder)
		if err != nil {
			t.Fatalf("RenderAndDecode() unexpected error: %v", err)
		}
		if !reflect.DeepEqual(obj.Name, "test") {
			t.Errorf("expected and got name differ:\n%s", cmp.Diff("test", obj.Name))
		}
		if rendered == "" {
			t.Error("RenderAndDecode() expected non-empty rendered string")
		}
	})

	t.Run("invalid template returns render error", func(t *testing.T) {
		t.Parallel()

		tmpl := template.Must(template.New("bad").Parse("{{.Missing}}"))
		_, _, err := RenderAndDecode[*corev1.ConfigMap](tmpl, map[string]string{}, decoder)
		if err == nil {
			t.Fatal("RenderAndDecode() expected error for invalid template, got nil")
		}
	})

	t.Run("valid render but bad YAML returns decode error", func(t *testing.T) {
		t.Parallel()

		tmpl := template.Must(template.New("notk8s").Parse("this is not k8s yaml"))
		_, rendered, err := RenderAndDecode[*corev1.ConfigMap](tmpl, nil, decoder)
		if err == nil {
			t.Fatal("RenderAndDecode() expected decode error, got nil")
		}
		if !reflect.DeepEqual(rendered, "this is not k8s yaml") {
			t.Errorf("expected and got rendered differ:\n%s", cmp.Diff("this is not k8s yaml", rendered))
		}
	})
}

func TestRenderAndDecode_Secret(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	cf := serializer.NewCodecFactory(scheme)
	decoder := cf.UniversalDeserializer()

	t.Run("valid render but bad YAML returns decode error with redaction", func(t *testing.T) {
		t.Parallel()

		tmpl := template.Must(template.New("badsecret").Parse("this is not k8s yaml"))
		_, rendered, err := RenderAndDecode[*corev1.Secret](tmpl, nil, decoder)
		if err == nil {
			t.Fatal("RenderAndDecode() expected decode error, got nil")
		}
		if !reflect.DeepEqual(rendered, "this is not k8s yaml") {
			t.Errorf("expected and got rendered differ:\n%s", cmp.Diff("this is not k8s yaml", rendered))
		}
	})
}

func TestParseObjectTemplate(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	cf := serializer.NewCodecFactory(scheme)
	decoder := cf.UniversalDeserializer()

	validConfigMapYAML := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n  namespace: default\ndata:\n  key: value\n"

	_, parseErr := template.New("bad").Funcs(nil).Parse("{{.Unclosed")

	tt := []struct {
		name        string
		tmplName    string
		tmplString  string
		expectedErr error
	}{
		{
			name:        "valid template",
			tmplName:    "test",
			tmplString:  validConfigMapYAML,
			expectedErr: nil,
		},

		{
			name:        "invalid template syntax",
			tmplName:    "bad",
			tmplString:  "{{.Unclosed",
			expectedErr: fmt.Errorf("can't parse template %q: %w", "bad", parseErr),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseObjectTemplate[*corev1.ConfigMap](tc.tmplName, tc.tmplString, nil, decoder)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}
		})
	}
}

func TestParseObjectTemplateOrDie(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	cf := serializer.NewCodecFactory(scheme)
	decoder := cf.UniversalDeserializer()

	validConfigMapYAML := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n  namespace: default\ndata:\n  key: value\n"

	t.Run("valid template does not panic", func(t *testing.T) {
		t.Parallel()

		ot := ParseObjectTemplateOrDie[*corev1.ConfigMap]("test", validConfigMapYAML, nil, decoder)
		_ = ot
	})

	t.Run("invalid template panics", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Error("ParseObjectTemplateOrDie() did not panic for invalid template")
			}
		}()
		ParseObjectTemplateOrDie[*corev1.ConfigMap]("bad", "{{.Unclosed", nil, decoder)
		t.Error("ParseObjectTemplateOrDie() should have panicked but did not")
	})
}

func TestObjectTemplateRenderObject(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	cf := serializer.NewCodecFactory(scheme)
	decoder := cf.UniversalDeserializer()

	validConfigMapYAML := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n  namespace: default\ndata:\n  key: value\n"

	t.Run("valid returns object and string", func(t *testing.T) {
		t.Parallel()

		ot, err := ParseObjectTemplate[*corev1.ConfigMap]("cm", validConfigMapYAML, nil, decoder)
		if err != nil {
			t.Fatalf("ParseObjectTemplate() unexpected error: %v", err)
		}
		obj, rendered, err := ot.RenderObject(nil)
		if err != nil {
			t.Fatalf("RenderObject() unexpected error: %v", err)
		}
		if !reflect.DeepEqual(obj.Name, "test") {
			t.Errorf("expected and got name differ:\n%s", cmp.Diff("test", obj.Name))
		}
		if rendered == "" {
			t.Error("RenderObject() expected non-empty rendered string")
		}
	})

	t.Run("invalid data returns error", func(t *testing.T) {
		t.Parallel()

		ot, err := ParseObjectTemplate[*corev1.ConfigMap]("tmpl", "Hello {{.Missing}}", nil, decoder)
		if err != nil {
			t.Fatalf("ParseObjectTemplate() unexpected error: %v", err)
		}
		_, _, err = ot.RenderObject(map[string]string{})
		if err == nil {
			t.Fatal("RenderObject() expected error for invalid data, got nil")
		}
	})
}

func TestObjectTemplateItemRender(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	cf := serializer.NewCodecFactory(scheme)
	decoder := cf.UniversalDeserializer()

	t.Run("delegates correctly", func(t *testing.T) {
		t.Parallel()

		tmplStr := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: {{.Name}}\n  namespace: default\ndata:\n  key: value\n"
		ot, err := ParseObjectTemplate[*corev1.ConfigMap]("item", tmplStr, nil, decoder)
		if err != nil {
			t.Fatalf("ParseObjectTemplate() unexpected error: %v", err)
		}
		item := &ObjectTemplateItem[*corev1.ConfigMap]{
			Template: ot,
			Data:     map[string]any{"Name": "my-cm"},
		}
		obj, rendered, err := item.Render()
		if err != nil {
			t.Fatalf("Render() unexpected error: %v", err)
		}
		if !reflect.DeepEqual(obj.Name, "my-cm") {
			t.Errorf("expected and got name differ:\n%s", cmp.Diff("my-cm", obj.Name))
		}
		if rendered == "" {
			t.Error("Render() expected non-empty rendered string")
		}
	})
}

func TestBytesDecode(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	cf := serializer.NewCodecFactory(scheme)

	validConfigMapYAML := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n  namespace: default\ndata:\n  key: value\n"

	_, invalidBytesErr := Bytes[*corev1.ConfigMap]("not valid k8s yaml at all").Decode(cf)

	tt := []struct {
		name              string
		data              Bytes[*corev1.ConfigMap]
		expectedName      string
		expectedNamespace string
		expectedDataKey   string
		expectedErr       error
	}{
		{
			name:              "valid YAML bytes returns typed ConfigMap",
			data:              Bytes[*corev1.ConfigMap](validConfigMapYAML),
			expectedName:      "test",
			expectedNamespace: "default",
			expectedDataKey:   "value",
			expectedErr:       nil,
		},

		{
			name:              "invalid bytes returns error",
			data:              Bytes[*corev1.ConfigMap]("not valid k8s yaml at all"),
			expectedName:      "",
			expectedNamespace: "",
			expectedDataKey:   "",
			expectedErr:       invalidBytesErr,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			obj, err := tc.data.Decode(cf)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}
			if err != nil {
				return
			}
			if obj.Name != tc.expectedName {
				t.Errorf("expected and got name differ:\n%s", cmp.Diff(tc.expectedName, obj.Name))
			}
			if obj.Namespace != tc.expectedNamespace {
				t.Errorf("expected and got namespace differ:\n%s", cmp.Diff(tc.expectedNamespace, obj.Namespace))
			}
			if obj.Data["key"] != tc.expectedDataKey {
				t.Errorf("expected and got data differ:\n%s", cmp.Diff(tc.expectedDataKey, obj.Data["key"]))
			}
		})
	}
}
