package taskmodel

type SetParamValTask struct {
	Name  string
	Value string
	Type  string
}

type GetParamValTask struct {
	Name []string
}

type GetParamNamesTask struct {
	ParameterPath string
	NextLevel     int
}

type GetParamAttrTask struct {
	Name []string
}

type AddTask struct {
	Name string
}

type DeleteTask struct {
	Name string
}

type SoapTask interface {
	AddTask | DeleteTask | GetParamValTask | SetParamValTask | []SetParamValTask
}
