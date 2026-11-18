package lib

type SessionContext struct {
	AC *ApplicationContext

	Session Session
}

type Session interface {
	GetSessionContext() *SessionContext
}

////////////////////////////////////////////////////////////////////////////////

type DefaultSession struct {
	context *SessionContext
}

// GetSessionContext implements [Session].
func (inst *DefaultSession) GetSessionContext() *SessionContext {
	return inst.context
}

// GetLifeCycle implements [Life].
func (inst *DefaultSession) GetLifeCycle() *LifeCycle {
	lc := &LifeCycle{
		OnInit: inst.onInit,
	}
	return lc
}

func (inst *DefaultSession) onInit(ac *ApplicationContext) error {

	sc := ac.SessionContext

	sc.Session = inst
	ac.Session = inst
	inst.context = sc

	return nil
}

func (inst *DefaultSession) _impl() (Session, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
