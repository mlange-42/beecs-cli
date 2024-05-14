package util

import (
	"reflect"

	"github.com/mlange-42/beecs/obs"
)

var observerRegistry map[string]reflect.Type

func init() {
	observerRegistry = map[string]reflect.Type{}

	registerObserver[obs.WorkerCohorts]()
	registerObserver[obs.ForagingPeriod]()
	registerObserver[obs.Stores]()
}

func registerObserver[T any]() {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	observerRegistry[tp.String()] = tp
}

func GetObserver(name string) (reflect.Type, bool) {
	t, ok := observerRegistry[name]
	return t, ok
}
