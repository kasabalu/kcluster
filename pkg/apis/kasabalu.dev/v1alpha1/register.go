package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var SchemeGroupVersion = schema.GroupVersion{Group: "kasabalu.dev", Version: "v1alpha1"}

// As soon as this PKG loaded , we have to register type to K8s

var (
	SchemeBuilder runtime.SchemeBuilder
)

func init() {
	//this gets called as soon as pkg gets loaded
	//SchemeBuilder.Register expects Func as perameter that will register k8s
	SchemeBuilder.Register(addKnownTypes)

}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, &Kluster{}, &KlusterList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
