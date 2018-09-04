// +build windows

package main

import (
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func main() {
	ole.CoInitialize(0)
	unknown, _ := oleutil.CreateObject("InternetExplorer.Application")
	ie, _ := unknown.QueryInterface(ole.IID_IDispatch)
	oleutil.CallMethod(ie, "Navigate", "http://www.google.com")
	oleutil.PutProperty(ie, "Visible", true)
	for {
		if oleutil.MustGetProperty(ie, "Busy").Val == 0 {
			break
		}
	}

	time.Sleep(1e9)

	document := oleutil.MustGetProperty(ie, "document").ToIDispatch()
	window := oleutil.MustGetProperty(document, "parentWindow").ToIDispatch()
	// set 'golang' to text box.
	oleutil.MustCallMethod(window, "eval", "document.getElementsByName('q')[0].value = 'golang'")
	// click btnG.
	btnG := oleutil.MustCallMethod(window, "eval", "document.getElementsByName('btnG')[0]").ToIDispatch()
	oleutil.MustCallMethod(btnG, "click")
}
