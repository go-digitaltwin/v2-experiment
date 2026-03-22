package coverage

// Task exercises fields with sibling types: types defined in other files within
// the same package. The generated code should reference them by their
// unqualified name and emit no imports, regardless of underlying type.
type Task struct {
	ID         string
	Status     Status
	Label      Token
	OnComplete Signal
}

// TaskContainers exercises container fields whose element or key types are
// sibling types defined in other files within the same package.
type TaskContainers struct {
	ID             string
	StatusLog      []Status
	NamesByStatus  map[Status]string
	StatusesByName map[string]Status
	Handled        map[Signal]Token
	Signals        [3]Signal
}
