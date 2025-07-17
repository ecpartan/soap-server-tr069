package methods

import (
	"context"
)

type MFunc func(ctx context.Context, req map[string]any) ([]byte, error)

type ReqMap map[string]map[string]MFunc

type FuncMap struct {
	ReqMap
}

func New() *FuncMap {
	return &FuncMap{
		ReqMap: make(ReqMap),
	}
}

func (m *ReqMap) Add(method, methodArgs string, funcM MFunc) {
	if _, ok := (*m)[method]; !ok {
		(*m)[method] = make(map[string]MFunc)
	}
	(*m)[method][methodArgs] = funcM
}
