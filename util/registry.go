package util

import (
	"reflect"

	"github.com/mlange-42/beecs/obs"
)

var observerRegistry map[string]reflect.Type
var resourcesRegistry map[string]reflect.Type

func init() {
	observerRegistry = map[string]reflect.Type{}
	RegisterObserver[obs.WorkerCohorts]()
	RegisterObserver[obs.ForagingPeriod]()
	RegisterObserver[obs.Stores]()
	RegisterObserver[obs.PatchNectar]()
	RegisterObserver[obs.PatchPollen]()

	resourcesRegistry = map[string]reflect.Type{}
	//registerResource[comp.Age]()
}

func RegisterObserver[T any]() {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	observerRegistry[tp.String()] = tp
}

func RegisterResource[T any]() {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	resourcesRegistry[tp.String()] = tp
}

func GetObserver(name string) (reflect.Type, bool) {
	t, ok := observerRegistry[name]
	return t, ok
}

func GetResource(name string) (reflect.Type, bool) {
	t, ok := resourcesRegistry[name]
	return t, ok
}
