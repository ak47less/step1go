package common

import "fmt"

////////////////////////////////////////////////////////////////////////////////
// 电脑按键编码

type ComputerKeyCode int

const (
	CKeyCodeMin ComputerKeyCode = iota

	// 可显示字符保持与 ASCII 一致。
	CKeyCodeGrave ComputerKeyCode = '`'

	CKeyCode1 ComputerKeyCode = '1'
	CKeyCode2 ComputerKeyCode = '2'
	CKeyCode3 ComputerKeyCode = '3'
	CKeyCode4 ComputerKeyCode = '4'
	CKeyCode5 ComputerKeyCode = '5'
	CKeyCode6 ComputerKeyCode = '6'
	CKeyCode7 ComputerKeyCode = '7'
	CKeyCode8 ComputerKeyCode = '8'
	CKeyCode9 ComputerKeyCode = '9'
	CKeyCode0 ComputerKeyCode = '0'

	CKeyCodeMinus ComputerKeyCode = '-'
	CKeyCodeEqual ComputerKeyCode = '='

	CKeyCodeQ ComputerKeyCode = 'q'
	CKeyCodeW ComputerKeyCode = 'w'
	CKeyCodeE ComputerKeyCode = 'e'
	CKeyCodeR ComputerKeyCode = 'r'
	CKeyCodeT ComputerKeyCode = 't'
	CKeyCodeY ComputerKeyCode = 'y'
	CKeyCodeU ComputerKeyCode = 'u'
	CKeyCodeI ComputerKeyCode = 'i'
	CKeyCodeO ComputerKeyCode = 'o'
	CKeyCodeP ComputerKeyCode = 'p'

	CKeyCodeLeftBracket  ComputerKeyCode = '['
	CKeyCodeRightBracket ComputerKeyCode = ']'
	CKeyCodeBackslash    ComputerKeyCode = '\\'

	CKeyCodeA ComputerKeyCode = 'a'
	CKeyCodeS ComputerKeyCode = 's'
	CKeyCodeD ComputerKeyCode = 'd'
	CKeyCodeF ComputerKeyCode = 'f'
	CKeyCodeG ComputerKeyCode = 'g'
	CKeyCodeH ComputerKeyCode = 'h'
	CKeyCodeJ ComputerKeyCode = 'j'
	CKeyCodeK ComputerKeyCode = 'k'
	CKeyCodeL ComputerKeyCode = 'l'

	CKeyCodeSemicolon  ComputerKeyCode = ';'
	CKeyCodeApostrophe ComputerKeyCode = '\''

	CKeyCodeZ ComputerKeyCode = 'z'
	CKeyCodeX ComputerKeyCode = 'x'
	CKeyCodeC ComputerKeyCode = 'c'
	CKeyCodeV ComputerKeyCode = 'v'
	CKeyCodeB ComputerKeyCode = 'b'
	CKeyCodeN ComputerKeyCode = 'n'
	CKeyCodeM ComputerKeyCode = 'm'

	CKeyCodeComma  ComputerKeyCode = ','
	CKeyCodePeriod ComputerKeyCode = '.'
	CKeyCodeSlash  ComputerKeyCode = '/'

	CKeyCodeSpace ComputerKeyCode = ' '
)

// 功能键使用 ASCII 可显示范围之外的独立编码。
const (
	CKeyCodeEscape ComputerKeyCode = 128 + iota
	CKeyCodeF1
	CKeyCodeF2
	CKeyCodeF3
	CKeyCodeF4
	CKeyCodeF5
	CKeyCodeF6
	CKeyCodeF7
	CKeyCodeF8
	CKeyCodeF9
	CKeyCodeF10
	CKeyCodeF11
	CKeyCodeF12

	CKeyCodePrintScreen
	CKeyCodeScrollLock
	CKeyCodePause

	CKeyCodeTab
	CKeyCodeBackspace
	CKeyCodeCapsLock
	CKeyCodeEnter

	CKeyCodeLeftShift
	CKeyCodeRightShift
	CKeyCodeLeftCtrl
	CKeyCodeRightCtrl
	CKeyCodeLeftAlt
	CKeyCodeRightAlt
	CKeyCodeLeftWin
	CKeyCodeRightWin
	CKeyCodeMenu

	CKeyCodeInsert
	CKeyCodeDelete
	CKeyCodeHome
	CKeyCodeEnd
	CKeyCodePageUp
	CKeyCodePageDown
	CKeyCodeUp
	CKeyCodeDown
	CKeyCodeLeft
	CKeyCodeRight

	CKeyCodeNumLock
	CKeyCodeNumpadDivide
	CKeyCodeNumpadMultiply
	CKeyCodeNumpadSubtract
	CKeyCodeNumpad7
	CKeyCodeNumpad8
	CKeyCodeNumpad9
	CKeyCodeNumpadAdd
	CKeyCodeNumpad4
	CKeyCodeNumpad5
	CKeyCodeNumpad6
	CKeyCodeNumpad1
	CKeyCodeNumpad2
	CKeyCodeNumpad3
	CKeyCodeNumpadEnter
	CKeyCodeNumpad0
	CKeyCodeNumpadDecimal

	CKeyCodeMax
)

// methods

func (code ComputerKeyCode) Name() ComputerKeyName {
	if name, ok := theComputerKeyNameByCode[code]; ok {
		return name
	}
	return CKeyNameMin
}

func (code ComputerKeyCode) String() string {
	name := code.Name()
	num := int(code)
	return fmt.Sprintf("%s(%d)", name, num)
}

////////////////////////////////////////////////////////////////////////////////
// 电脑按键名称

type ComputerKeyName string

const (
	CKeyNameMin ComputerKeyName = "min"

	CKeyNameGrave ComputerKeyName = "`"

	CKeyName1 ComputerKeyName = "1"
	CKeyName2 ComputerKeyName = "2"
	CKeyName3 ComputerKeyName = "3"
	CKeyName4 ComputerKeyName = "4"
	CKeyName5 ComputerKeyName = "5"
	CKeyName6 ComputerKeyName = "6"
	CKeyName7 ComputerKeyName = "7"
	CKeyName8 ComputerKeyName = "8"
	CKeyName9 ComputerKeyName = "9"
	CKeyName0 ComputerKeyName = "0"

	CKeyNameMinus ComputerKeyName = "-"
	CKeyNameEqual ComputerKeyName = "="

	CKeyNameQ ComputerKeyName = "q"
	CKeyNameW ComputerKeyName = "w"
	CKeyNameE ComputerKeyName = "e"
	CKeyNameR ComputerKeyName = "r"
	CKeyNameT ComputerKeyName = "t"
	CKeyNameY ComputerKeyName = "y"
	CKeyNameU ComputerKeyName = "u"
	CKeyNameI ComputerKeyName = "i"
	CKeyNameO ComputerKeyName = "o"
	CKeyNameP ComputerKeyName = "p"

	CKeyNameLeftBracket  ComputerKeyName = "["
	CKeyNameRightBracket ComputerKeyName = "]"
	CKeyNameBackslash    ComputerKeyName = "\\"

	CKeyNameA ComputerKeyName = "a"
	CKeyNameS ComputerKeyName = "s"
	CKeyNameD ComputerKeyName = "d"
	CKeyNameF ComputerKeyName = "f"
	CKeyNameG ComputerKeyName = "g"
	CKeyNameH ComputerKeyName = "h"
	CKeyNameJ ComputerKeyName = "j"
	CKeyNameK ComputerKeyName = "k"
	CKeyNameL ComputerKeyName = "l"

	CKeyNameSemicolon  ComputerKeyName = ";"
	CKeyNameApostrophe ComputerKeyName = "'"

	CKeyNameZ ComputerKeyName = "z"
	CKeyNameX ComputerKeyName = "x"
	CKeyNameC ComputerKeyName = "c"
	CKeyNameV ComputerKeyName = "v"
	CKeyNameB ComputerKeyName = "b"
	CKeyNameN ComputerKeyName = "n"
	CKeyNameM ComputerKeyName = "m"

	CKeyNameComma  ComputerKeyName = ","
	CKeyNamePeriod ComputerKeyName = "."
	CKeyNameSlash  ComputerKeyName = "/"

	CKeyNameSpace ComputerKeyName = "space"

	CKeyNameEscape      ComputerKeyName = "escape"
	CKeyNameF1          ComputerKeyName = "f1"
	CKeyNameF2          ComputerKeyName = "f2"
	CKeyNameF3          ComputerKeyName = "f3"
	CKeyNameF4          ComputerKeyName = "f4"
	CKeyNameF5          ComputerKeyName = "f5"
	CKeyNameF6          ComputerKeyName = "f6"
	CKeyNameF7          ComputerKeyName = "f7"
	CKeyNameF8          ComputerKeyName = "f8"
	CKeyNameF9          ComputerKeyName = "f9"
	CKeyNameF10         ComputerKeyName = "f10"
	CKeyNameF11         ComputerKeyName = "f11"
	CKeyNameF12         ComputerKeyName = "f12"
	CKeyNamePrintScreen ComputerKeyName = "print_screen"
	CKeyNameScrollLock  ComputerKeyName = "scroll_lock"
	CKeyNamePause       ComputerKeyName = "pause"
	CKeyNameTab         ComputerKeyName = "tab"
	CKeyNameBackspace   ComputerKeyName = "backspace"
	CKeyNameCapsLock    ComputerKeyName = "caps_lock"
	CKeyNameEnter       ComputerKeyName = "enter"

	CKeyNameLeftShift  ComputerKeyName = "left_shift"
	CKeyNameRightShift ComputerKeyName = "right_shift"
	CKeyNameLeftCtrl   ComputerKeyName = "left_ctrl"
	CKeyNameRightCtrl  ComputerKeyName = "right_ctrl"
	CKeyNameLeftAlt    ComputerKeyName = "left_alt"
	CKeyNameRightAlt   ComputerKeyName = "right_alt"
	CKeyNameLeftWin    ComputerKeyName = "left_win"
	CKeyNameRightWin   ComputerKeyName = "right_win"
	CKeyNameMenu       ComputerKeyName = "menu"

	CKeyNameInsert   ComputerKeyName = "insert"
	CKeyNameDelete   ComputerKeyName = "delete"
	CKeyNameHome     ComputerKeyName = "home"
	CKeyNameEnd      ComputerKeyName = "end"
	CKeyNamePageUp   ComputerKeyName = "page_up"
	CKeyNamePageDown ComputerKeyName = "page_down"
	CKeyNameUp       ComputerKeyName = "up"
	CKeyNameDown     ComputerKeyName = "down"
	CKeyNameLeft     ComputerKeyName = "left"
	CKeyNameRight    ComputerKeyName = "right"

	CKeyNameNumLock        ComputerKeyName = "num_lock"
	CKeyNameNumpadDivide   ComputerKeyName = "numpad_divide"
	CKeyNameNumpadMultiply ComputerKeyName = "numpad_multiply"
	CKeyNameNumpadSubtract ComputerKeyName = "numpad_subtract"
	CKeyNameNumpad7        ComputerKeyName = "numpad_7"
	CKeyNameNumpad8        ComputerKeyName = "numpad_8"
	CKeyNameNumpad9        ComputerKeyName = "numpad_9"
	CKeyNameNumpadAdd      ComputerKeyName = "numpad_add"
	CKeyNameNumpad4        ComputerKeyName = "numpad_4"
	CKeyNameNumpad5        ComputerKeyName = "numpad_5"
	CKeyNameNumpad6        ComputerKeyName = "numpad_6"
	CKeyNameNumpad1        ComputerKeyName = "numpad_1"
	CKeyNameNumpad2        ComputerKeyName = "numpad_2"
	CKeyNameNumpad3        ComputerKeyName = "numpad_3"
	CKeyNameNumpadEnter    ComputerKeyName = "numpad_enter"
	CKeyNameNumpad0        ComputerKeyName = "numpad_0"
	CKeyNameNumpadDecimal  ComputerKeyName = "numpad_decimal"

	CKeyNameMax ComputerKeyName = "max"
)

// methods

func (name ComputerKeyName) Code() ComputerKeyCode {
	if code, ok := theComputerKeyCodeByName[name]; ok {
		return code
	}
	return CKeyCodeMin
}

func (name ComputerKeyName) String() string {
	return string(name)
}

////////////////////////////////////////////////////////////////////////////////

type ComputerKeyInfo struct {
	Code ComputerKeyCode

	Name ComputerKeyName
}

// 按标准 104 键布局从左到右、从上到下排列。
var theComputerKeys = [...]ComputerKeyInfo{
	{CKeyCodeEscape, CKeyNameEscape},
	{CKeyCodeF1, CKeyNameF1},
	{CKeyCodeF2, CKeyNameF2},
	{CKeyCodeF3, CKeyNameF3},
	{CKeyCodeF4, CKeyNameF4},
	{CKeyCodeF5, CKeyNameF5},
	{CKeyCodeF6, CKeyNameF6},
	{CKeyCodeF7, CKeyNameF7},
	{CKeyCodeF8, CKeyNameF8},
	{CKeyCodeF9, CKeyNameF9},
	{CKeyCodeF10, CKeyNameF10},
	{CKeyCodeF11, CKeyNameF11},
	{CKeyCodeF12, CKeyNameF12},
	{CKeyCodePrintScreen, CKeyNamePrintScreen},
	{CKeyCodeScrollLock, CKeyNameScrollLock},
	{CKeyCodePause, CKeyNamePause},

	{CKeyCodeGrave, CKeyNameGrave},
	{CKeyCode1, CKeyName1},
	{CKeyCode2, CKeyName2},
	{CKeyCode3, CKeyName3},
	{CKeyCode4, CKeyName4},
	{CKeyCode5, CKeyName5},
	{CKeyCode6, CKeyName6},
	{CKeyCode7, CKeyName7},
	{CKeyCode8, CKeyName8},
	{CKeyCode9, CKeyName9},
	{CKeyCode0, CKeyName0},
	{CKeyCodeMinus, CKeyNameMinus},
	{CKeyCodeEqual, CKeyNameEqual},
	{CKeyCodeBackspace, CKeyNameBackspace},

	{CKeyCodeTab, CKeyNameTab},
	{CKeyCodeQ, CKeyNameQ},
	{CKeyCodeW, CKeyNameW},
	{CKeyCodeE, CKeyNameE},
	{CKeyCodeR, CKeyNameR},
	{CKeyCodeT, CKeyNameT},
	{CKeyCodeY, CKeyNameY},
	{CKeyCodeU, CKeyNameU},
	{CKeyCodeI, CKeyNameI},
	{CKeyCodeO, CKeyNameO},
	{CKeyCodeP, CKeyNameP},
	{CKeyCodeLeftBracket, CKeyNameLeftBracket},
	{CKeyCodeRightBracket, CKeyNameRightBracket},
	{CKeyCodeBackslash, CKeyNameBackslash},

	{CKeyCodeCapsLock, CKeyNameCapsLock},
	{CKeyCodeA, CKeyNameA},
	{CKeyCodeS, CKeyNameS},
	{CKeyCodeD, CKeyNameD},
	{CKeyCodeF, CKeyNameF},
	{CKeyCodeG, CKeyNameG},
	{CKeyCodeH, CKeyNameH},
	{CKeyCodeJ, CKeyNameJ},
	{CKeyCodeK, CKeyNameK},
	{CKeyCodeL, CKeyNameL},
	{CKeyCodeSemicolon, CKeyNameSemicolon},
	{CKeyCodeApostrophe, CKeyNameApostrophe},
	{CKeyCodeEnter, CKeyNameEnter},

	{CKeyCodeLeftShift, CKeyNameLeftShift},
	{CKeyCodeZ, CKeyNameZ},
	{CKeyCodeX, CKeyNameX},
	{CKeyCodeC, CKeyNameC},
	{CKeyCodeV, CKeyNameV},
	{CKeyCodeB, CKeyNameB},
	{CKeyCodeN, CKeyNameN},
	{CKeyCodeM, CKeyNameM},
	{CKeyCodeComma, CKeyNameComma},
	{CKeyCodePeriod, CKeyNamePeriod},
	{CKeyCodeSlash, CKeyNameSlash},
	{CKeyCodeRightShift, CKeyNameRightShift},

	{CKeyCodeLeftCtrl, CKeyNameLeftCtrl},
	{CKeyCodeLeftWin, CKeyNameLeftWin},
	{CKeyCodeLeftAlt, CKeyNameLeftAlt},
	{CKeyCodeSpace, CKeyNameSpace},
	{CKeyCodeRightAlt, CKeyNameRightAlt},
	{CKeyCodeRightWin, CKeyNameRightWin},
	{CKeyCodeMenu, CKeyNameMenu},
	{CKeyCodeRightCtrl, CKeyNameRightCtrl},

	{CKeyCodeInsert, CKeyNameInsert},
	{CKeyCodeHome, CKeyNameHome},
	{CKeyCodePageUp, CKeyNamePageUp},
	{CKeyCodeDelete, CKeyNameDelete},
	{CKeyCodeEnd, CKeyNameEnd},
	{CKeyCodePageDown, CKeyNamePageDown},
	{CKeyCodeLeft, CKeyNameLeft},
	{CKeyCodeUp, CKeyNameUp},
	{CKeyCodeDown, CKeyNameDown},
	{CKeyCodeRight, CKeyNameRight},

	{CKeyCodeNumLock, CKeyNameNumLock},
	{CKeyCodeNumpadDivide, CKeyNameNumpadDivide},
	{CKeyCodeNumpadMultiply, CKeyNameNumpadMultiply},
	{CKeyCodeNumpadSubtract, CKeyNameNumpadSubtract},
	{CKeyCodeNumpad7, CKeyNameNumpad7},
	{CKeyCodeNumpad8, CKeyNameNumpad8},
	{CKeyCodeNumpad9, CKeyNameNumpad9},
	{CKeyCodeNumpadAdd, CKeyNameNumpadAdd},
	{CKeyCodeNumpad4, CKeyNameNumpad4},
	{CKeyCodeNumpad5, CKeyNameNumpad5},
	{CKeyCodeNumpad6, CKeyNameNumpad6},
	{CKeyCodeNumpad1, CKeyNameNumpad1},
	{CKeyCodeNumpad2, CKeyNameNumpad2},
	{CKeyCodeNumpad3, CKeyNameNumpad3},
	{CKeyCodeNumpadEnter, CKeyNameNumpadEnter},
	{CKeyCodeNumpad0, CKeyNameNumpad0},
	{CKeyCodeNumpadDecimal, CKeyNameNumpadDecimal},
}

var theComputerKeyNameByCode, theComputerKeyCodeByName = buildComputerKeyIndexes()

func buildComputerKeyIndexes() (map[ComputerKeyCode]ComputerKeyName, map[ComputerKeyName]ComputerKeyCode) {
	nameByCode := make(map[ComputerKeyCode]ComputerKeyName, len(theComputerKeys))
	codeByName := make(map[ComputerKeyName]ComputerKeyCode, len(theComputerKeys))
	for _, key := range theComputerKeys {
		nameByCode[key.Code] = key.Name
		codeByName[key.Name] = key.Code
	}
	return nameByCode, codeByName
}

////////////////////////////////////////////////////////////////////////////////
// 电脑键盘对象

type ComputerKeyboard struct {

	// 表示pc上常用的104键盘

}

func (inst *ComputerKeyboard) ListCodes() []ComputerKeyCode {
	items := make([]ComputerKeyCode, len(theComputerKeys))
	for i, key := range theComputerKeys {
		items[i] = key.Code
	}
	return items
}

func (inst *ComputerKeyboard) ListNames() []ComputerKeyName {
	items := make([]ComputerKeyName, len(theComputerKeys))
	for i, key := range theComputerKeys {
		items[i] = key.Name
	}
	return items
}

func (inst *ComputerKeyboard) ListAll() []*ComputerKeyInfo {
	items := make([]*ComputerKeyInfo, len(theComputerKeys))
	for i, key := range theComputerKeys {
		items[i] = &ComputerKeyInfo{
			Code: key.Code,
			Name: key.Name,
		}
	}
	return items
}

////////////////////////////////////////////////////////////////////////////////
// EOF
