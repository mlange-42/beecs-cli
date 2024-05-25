package registry

import (
	"fmt"
	"reflect"

	"github.com/mlange-42/arche-pixel/plot"
	"github.com/mlange-42/beecs-cli/view"
	"github.com/mlange-42/beecs/obs"
	"github.com/mlange-42/beecs/sys"
)

var observerRegistry = map[string]reflect.Type{}
var drawersRegistry = map[string]reflect.Type{}
var resourcesRegistry = map[string]reflect.Type{}
var systemsRegistry = map[string]reflect.Type{}

func init() {
	RegisterObserver[obs.WorkerCohorts]()
	RegisterObserver[obs.ForagingPeriod]()
	RegisterObserver[obs.Stores]()
	RegisterObserver[obs.PatchNectar]()
	RegisterObserver[obs.PatchPollen]()
	RegisterObserver[obs.NectarVisits]()
	RegisterObserver[obs.PollenVisits]()

	RegisterObserver[obs.AgeStructure]()
	RegisterObserver[obs.ForagingStats]()

	RegisterDrawer[plot.Monitor]()
	RegisterDrawer[plot.Resources]()
	RegisterDrawer[plot.Systems]()
	RegisterDrawer[view.Foraging]()

	//RegisterResource[...]()

	RegisterSystem[sys.InitStore]()
	RegisterSystem[sys.InitCohorts]()
	RegisterSystem[sys.InitPopulation]()
	RegisterSystem[sys.InitPatchesList]()
	RegisterSystem[sys.InitForagingPeriod]()

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
	if _, ok := observerRegistry[tp.String()]; ok {
		panic(fmt.Sprintf("there is already an observer with type name '%s' registered", tp.String()))
	}
	observerRegistry[tp.String()] = tp
}

func RegisterDrawer[T any]() {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	if _, ok := drawersRegistry[tp.String()]; ok {
		panic(fmt.Sprintf("there is already a drawer with type name '%s' registered", tp.String()))
	}
	drawersRegistry[tp.String()] = tp
}

func RegisterResource[T any]() {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	if _, ok := resourcesRegistry[tp.String()]; ok {
		panic(fmt.Sprintf("there is already a resource with type name '%s' registered", tp.String()))
	}
	resourcesRegistry[tp.String()] = tp
}

func RegisterSystem[T any]() {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	if _, ok := systemsRegistry[tp.String()]; ok {
		panic(fmt.Sprintf("there is already a system with type name '%s' registered", tp.String()))
	}
	systemsRegistry[tp.String()] = tp
}

func GetObserver(name string) (reflect.Type, bool) {
	t, ok := observerRegistry[name]
	return t, ok
}

func GetDrawer(name string) (reflect.Type, bool) {
	t, ok := drawersRegistry[name]
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
