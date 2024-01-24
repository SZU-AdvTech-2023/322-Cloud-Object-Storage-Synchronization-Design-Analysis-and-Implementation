package main

type synchronizeEventEmitter struct {
	//manage the parameter of the event
	observers []synchronizeEventHandler
	taskIndex int
	fromPath  string
	toPath    string
	kwargs    map[string]string
}

func (see *synchronizeEventEmitter) init() {
	see.observers = make([]synchronizeEventHandler, 0)
	see.taskIndex = 0
	see.fromPath = ""
	see.toPath = ""
	see.kwargs = map[string]string{}
}

func (see *synchronizeEventEmitter) register(observer synchronizeEventHandler) {
	//register the observer to observers
	for _, ob := range see.observers {
		if SEHeq(ob, observer) {
			return
		}
	}
	see.observers = append(see.observers, observer)
}

func (see *synchronizeEventEmitter) unregister(observer synchronizeEventHandler) {
	//unregister the observer from observers
	for index, ob := range see.observers {
		if SEHeq(ob, observer) {
			see.observers = append(see.observers[:index], see.observers[index+1:]...)
			break
		}
	}
}

func (see *synchronizeEventEmitter) notify() {
	//note the observer to update itself
	for _, observer := range see.observers {
		observer.update(*see)
	}
}

func (see *synchronizeEventEmitter) setData(taskIndex int, args []string, kwargs map[string]string) {
	see.taskIndex = taskIndex
	see.fromPath = args[0]
	if len(args) > 1 {
		see.toPath = args[1]
	} else {
		see.toPath = ""
	}
	see.kwargs = kwargs
	see.notify()
}
