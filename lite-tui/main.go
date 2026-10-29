package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/ak47less/step1go/lite-tui/gui"
)

func main() {

	fmt.Println("hello, gui")
	text := "aaa\nbbb\n\tccc"
	label := widget.NewLabel(text)
	myApp := app.New()
	myWindow := myApp.NewWindow("step1go/lite/gui")
	ctx := new(gui.Context)
	km := new(KeyMapper)

	ctx.Init()
	km.init()

	err := ctx.GetServer().Start()
	if err != nil {
		panic(err)
	}

	size := fyne.Size{
		Width:  480,
		Height: 320,
	}

	myWindow.SetContent(label)
	myWindow.Resize(size)
	myWindow.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {

		code := km.Map(ke.Name)
		logger := ctx.GetLogger()
		stat := ctx.GetStatus()

		stat.PressKeyCode = code
		stat.PressKeyCount++

		text = stat.String()

		ses, err := ctx.GetSession()
		if err == nil && ses != nil {
			err = ses.DispatchKeyPress(code)
		}
		ctx.HandleError(err)

		label.SetText(text)
		logger.Info("status: %s", text)
	})

	text = ctx.GetStatus().String()
	fmt.Print(text)
	label.SetText(text)

	myWindow.ShowAndRun()

}

////////////////////////////////////////////////////////////////////////////////
// KeyMapper

type KeyMapper struct {
	table map[fyne.KeyName]byte
}

func (inst *KeyMapper) init() {

	table := make(map[fyne.KeyName]byte)

	const (
		theCodeBackSpace = 0x08
		theCodeEnterR    = 0x0d
		theCodeEnterL    = 0x0d
		theCodeEscape    = 0x1b
		theCodeSpace     = 0x20
		theCodeDelete    = 0x7f

		theCodeF1  = 'a'
		theCodeF2  = 'b'
		theCodeF3  = 'c'
		theCodeF4  = 'd'
		theCodeF5  = 'e'
		theCodeF6  = 'f'
		theCodeF7  = 'g'
		theCodeF8  = 'h'
		theCodeF9  = 'i'
		theCodeF10 = 'j'
		theCodeF11 = 'k'
		theCodeF12 = 'l'
	)

	table[fyne.KeyBackspace] = theCodeBackSpace
	table[fyne.KeyEnter] = theCodeEnterR
	table[fyne.KeyReturn] = theCodeEnterL
	table[fyne.KeyEscape] = theCodeEscape
	table[fyne.KeySpace] = theCodeSpace
	table[fyne.KeyDelete] = theCodeDelete

	table[fyne.KeyF1] = theCodeF1
	table[fyne.KeyF2] = theCodeF2
	table[fyne.KeyF3] = theCodeF3
	table[fyne.KeyF4] = theCodeF4
	table[fyne.KeyF5] = theCodeF5
	table[fyne.KeyF6] = theCodeF6
	table[fyne.KeyF7] = theCodeF7
	table[fyne.KeyF8] = theCodeF8
	table[fyne.KeyF9] = theCodeF9
	table[fyne.KeyF10] = theCodeF10
	table[fyne.KeyF11] = theCodeF11
	table[fyne.KeyF12] = theCodeF12

	inst.table = table
}

func (inst *KeyMapper) Map(name fyne.KeyName) byte {
	size := len(name)
	if size == 1 {
		b := byte(name[0])
		return b
	}
	return inst.table[name]
}

////////////////////////////////////////////////////////////////////////////////
// EOF
