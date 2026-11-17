package lib

type SessionContext struct {
	Session Session
}

type Session interface {
	GetSessionContext() *SessionContext
}
