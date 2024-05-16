package util

import (
	"reflect"

	"github.com/mlange-42/beecs/obs"
	"github.com/mlange-42/beecs/sys"
)

var observerRegistry map[string]reflect.Type
var resourcesRegistry map[string]reflect.Type
var systemsRegistry map[string]reflect.Type

func init() {
	observerRegistry = map[string]reflect.Type{}
	RegisterObserver[obs.WorkerCohorts]()
	RegisterObserver[obs.ForagingPeriod]()
	RegisterObserver[obs.Stores]()
	RegisterObserver[obs.PatchNectar]()
	RegisterObserver[obs.PatchPollen]()

	resourcesRegistry = map[string]reflect.Type{}
	//RegisterResource[comp.Age]()

	systemsRegistry = map[string]reflect.Type{}
	RegisterSystem[sys.InitStore]()
	RegisterSystem[sys.InitCohorts]()
	RegisterSystem[sys.InitPopulation]()
	RegisterSystem[sys.InitPatchesList]()

	RegisterSystem[sys.CalcAff]()
	RegisterSystem[sys.CalcForagingPeriod]()
	RegisterSystem[sys.ReplenishPatches]()
	RegisterSystem[sys.BroodCare]()
	RegisterSystem[sys.AgeCohorts]()
	RegisterSystem[sys.TransitionForagers]()
	RegisterSystem[sys.EggLaying]()

	RegisterSystem[sys.MortalityCohorts]()
	RegisterSystem[sys.MortalityForagers]()

	RegisterSystem[sys.Foraging]()
	RegisterSystem[sys.HoneyConsumption]()
	RegisterSystem[sys.PollenConsumption]()

	RegisterSystem[sys.CountPopulation]()
	RegisterSystem[sys.FixedTermination]()
}

func RegisterObserver[T any]() {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	observerRegistry[tp.String()] = tp
}

func RegisterResource[T any]() {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	resourcesRegistry[tp.String()] = tp
}

func RegisterSystem[T any]() {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	systemsRegistry[tp.String()] = tp
}

func GetObserver(name string) (reflect.Type, bool) {
	t, ok := observerRegistry[name]
	return t, ok
}

func GetResource(name string) (reflect.Type, bool) {
	t, ok := resourcesRegistry[name]
	return t, ok
}

func GetSystem(name string) (reflect.Type, bool) {
	t, ok := systemsRegistry[name]
	return t, ok
}
