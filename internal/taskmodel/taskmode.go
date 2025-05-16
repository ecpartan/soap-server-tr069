package taskmodel

type SetParamTask struct {
	Name  string
	Value string
	Type  string
}

type GetParamTask struct {
	Name []string
}

type AddTask struct {
	Name string
}

type DeleteTask struct {
	Name string
}

type SoapTask interface {
	AddTask | DeleteTask | GetParamTask | SetParamTask | []SetParamTask
}
