// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package plugin

import (
	"sync"

	"github.com/azure/kaito/pkg/inference"
)

type Registration struct {
	Name     string
	Instance inference.Model
}

type ModelRegister struct {
	sync.RWMutex
	models map[string]*Registration
}

var KaitoModelRegister ModelRegister

// Register allows model to be added
func (reg *ModelRegister) Register(r *Registration) {
	reg.Lock()
	defer reg.Unlock()
	if r.Name == "" {
		panic("model name is not specified")
	}

	if reg.models == nil {
		reg.models = make(map[string]*Registration)
	}

	reg.models[r.Name] = r
}

func (reg *ModelRegister) MustGet(name string) inference.Model {
	reg.Lock()
	defer reg.Unlock()
	if _, ok := reg.models[name]; ok {
		return reg.models[name].Instance
	}
	panic("model is not registered")
}

func (reg *ModelRegister) ListModelNames() []string {
	reg.Lock()
	defer reg.Unlock()
	n := []string{}
	for k, _ := range reg.models {
		n = append(n, k)
	}
	return n
}
