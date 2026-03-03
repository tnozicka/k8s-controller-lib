package controllerhelpers

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClusterKey(t *testing.T) {
	t.Parallel()

	got := MakeClusterKey()
	if !reflect.DeepEqual(got, ClusterKey{}) {
		t.Errorf("expected and got differ:\n%s", cmp.Diff(ClusterKey{}, got))
	}

	gotName := got.ObjectName()
	if gotName != "cluster" {
		t.Errorf("expected and got ObjectName differ:\n%s", cmp.Diff("cluster", gotName))
	}
}
