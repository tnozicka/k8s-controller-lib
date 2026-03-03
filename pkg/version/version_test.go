package version

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	apimachineryversion "k8s.io/apimachinery/pkg/version"
)

func TestGet(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		expected apimachineryversion.Info
	}{
		{
			name: "default values",
			expected: apimachineryversion.Info{
				Major:        "",
				Minor:        "",
				GitCommit:    GitCommit,
				GitVersion:   Version,
				GitTreeState: gitTreeState,
				BuildDate:    BuildDate,
				GoVersion:    runtime.Version(),
				Compiler:     runtime.Compiler,
				Platform: fmt.Sprintf(
					"%s/%s", runtime.GOOS, runtime.GOARCH,
				),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Get()
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}
