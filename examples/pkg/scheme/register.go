package scheme

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	Scheme = runtime.NewScheme()

	localSchemeBuilder = runtime.SchemeBuilder{
		kscheme.AddToScheme,
	}

	Install = localSchemeBuilder.AddToScheme
)

func init() {
	utilruntime.Must(Install(Scheme))
}
