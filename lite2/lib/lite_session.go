package lib

type SessionContext struct {
	AC *ApplicationContext

	Session Session
}

type Session interface {
	GetSessionContext() *SessionContext
}
