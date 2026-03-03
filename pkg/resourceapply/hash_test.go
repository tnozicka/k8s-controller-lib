package resourceapply

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
)

func Test_hashAnnotator_SetHashAnnotationWithCleanup(t *testing.T) {
	t.Parallel()

	const hashAnnotationKey = "test/managed-hash"

	newObject := func() kubetypes.Object {
		return &objectType{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"test": "test",
				},
			},
		}
	}

	newObjectWithExplicitHash := func(hash string) kubetypes.Object {
		obj := newObject()

		a := newObject().GetAnnotations()
		if a == nil {
			a = map[string]string{}
		}
		a[hashAnnotationKey] = hash
		obj.SetAnnotations(a)

		return obj
	}

	newObjectWithCorrectHash := func() kubetypes.Object {
		return newObjectWithExplicitHash("LzBYWLIorAv8RbPCO0ogp8mFy8tpOtOgMI5qx+WME7lwQEm7kkby4QBb1dgEdQbnVwwjd8p543TJZC36dM2yow==")
	}

	tt := []struct {
		name        string
		obj         kubetypes.Object
		expectedObj kubetypes.Object
	}{
		{
			name:        "object without a managed annotation",
			obj:         newObject(),
			expectedObj: newObjectWithCorrectHash(),
		},
		{
			name:        "object with correct hash",
			obj:         newObjectWithCorrectHash(),
			expectedObj: newObjectWithCorrectHash(),
		},
		{
			name:        "object with incorrect hash gets updated",
			obj:         newObjectWithExplicitHash("foo=="),
			expectedObj: newObjectWithCorrectHash(),
		},
		{
			name: "resource version doesn't change the hash but is retained in the object",
			obj: func() kubetypes.Object {
				obj := newObjectWithCorrectHash()
				obj.SetResourceVersion("42")
				return obj
			}(),
			expectedObj: func() kubetypes.Object {
				obj := newObjectWithCorrectHash()
				obj.SetResourceVersion("42")
				return obj
			}(),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ha := NewHashAnnotator(hashAnnotationKey)

			got := tc.expectedObj.DeepCopyObject().(kubetypes.Object)
			err := ha.SetHashAnnotationWithCleanup(got)
			if err != nil {
				t.Fatal(err)
			}

			if !apiequality.Semantic.DeepEqual(got, tc.expectedObj) {
				t.Errorf("expected and got objects differ:\n%s", cmp.Diff(tc.expectedObj, got))
			}
		})
	}
}
