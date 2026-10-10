package core

import "github.com/ak47less/step1go/step1core"

type DefaultLifeManager struct {
	_main step1core.LifeAPI

	buffer []*step1core.Life // 未经加工的 items
	cached []*step1core.Life // 经过排序 & 排重的 items (按照 item.order 递增 )
}

// AddAPI implements [step1core.LifeManager].
func (inst *DefaultLifeManager) AddAPI(api step1core.LifeAPI) {
	if inst == nil || api == nil {
		return
	}
	life := api.GetLife()
	inst.innerAdd(life)
}

// AddElement implements [step1core.LifeManager].
func (inst *DefaultLifeManager) AddElement(el step1core.Element) {

	api, ok := el.(step1core.LifeAPI)
	if !ok {
		return
	}
	if api == nil {
		return
	}
	life := api.GetLife()
	inst.innerAdd(life)
}

// AddLife implements [step1core.LifeManager].
func (inst *DefaultLifeManager) AddLife(life *step1core.Life) {
	inst.innerAdd(life)
}

func (inst *DefaultLifeManager) innerAdd(life *step1core.Life) {
	if life == nil || inst == nil {
		return
	}
	inst.buffer = append(inst.buffer, life)
	inst.cached = nil
}

// GetMain implements [step1core.LifeManager].
func (inst *DefaultLifeManager) GetMain() step1core.LifeAPI {
	m := inst._main
	if m == nil {
		m = inst.innerLoadMain()
		inst._main = m
	}
	return m
}

func (inst *DefaultLifeManager) innerLoadMain() step1core.LifeAPI {
	panic("unimplemented")
}

// ListAll implements [step1core.LifeManager].
func (inst *DefaultLifeManager) ListAll() []*step1core.Life {
	panic("unimplemented")
}

func (inst *DefaultLifeManager) _impl() step1core.LifeManager {
	return inst
}
