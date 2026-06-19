package constant

type AppState int

const (
	AppStateConfig AppState = iota
	AppStatePassword
	AppStateDBSelect
	AppStateMain
)

type Focus int

const (
	FocusTable Focus = iota
	FocusColumn
	FocusEditor
)
