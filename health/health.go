package health

type Check struct {
	Name    string
	Healthy bool
	Message string
}

type Status struct {
	Live   bool
	Ready  bool
	Checks []Check
}

