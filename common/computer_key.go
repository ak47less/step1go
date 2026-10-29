package common

////////////////////////////////////////////////////////////////////////////////
// 电脑按键编码

type ComputerKeyCode int

const (
	CKeyCodeMin ComputerKeyCode = iota

	//todo: full list of 104-keys

	CKeyCodeMax
)

func (inst *ComputerKeyCode) Name() ComputerKeyName {

	//todo

}

////////////////////////////////////////////////////////////////////////////////
// 电脑按键名称

type ComputerKeyName string

const (
	CKeyNameMin ComputerKeyName = "min"

	//todo: full list of 104-keys

	CKeyNameMax ComputerKeyName = "max"
)

func (inst *ComputerKeyName) Code() ComputerKeyCode {

	//todo
}

////////////////////////////////////////////////////////////////////////////////

type ComputerKeyInfo struct {
	Code ComputerKeyCode

	Name ComputerKeyName
}

////////////////////////////////////////////////////////////////////////////////
// 电脑键盘对象

type ComputerKeyboard struct {

	// 表示pc上常用的104键盘

}

func (inst *ComputerKeyboard) ListCodes() []ComputerKeyCode {

	// todo

}

func (inst *ComputerKeyboard) ListNames() []ComputerKeyName {
	// todo
}

func (inst *ComputerKeyboard) ListAll() []*ComputerKeyInfo {
	// todo
}

////////////////////////////////////////////////////////////////////////////////
// EOF
