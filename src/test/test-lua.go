package main

import (
	"lua"
	"fmt"
	"strconv"
)


type TestStruct struct {
	Gihan string
	Test string
}

func (t *TestStruct) C(a int, b int) string{
	t.Gihan = "c"+ strconv.Itoa(a * b)
	return "c"
	
}

func main() {
	L, err := lua.NewState (true)
	if (err == nil) {
//		err = L.LoadCodeString ("local a = 10; return a + 20")
//		err = L.LoadCodeString ("function test(n) return n*n*n end")
		err = L.LoadCodeString ("function test(p) print (p.Gihan) p.C()  a = p.Gihan p.Test = \"hello\" return a end")
		L.SetTop(0)
		if (err == nil) {
			L.GetGlobal ("test")
			var t* TestStruct
			t = new(TestStruct)
			t.Gihan = "Hello"
			L.PushInterface(t)
			err = L.PCall (1, 1)
			
			if err != nil {
				print (err.Error())
			}else {
				a := L.ToString(-1)
				fmt.Printf("%s : %s", a, t.Test)	
			}
			defer L.Close()
		}else {
			print (err.Error())
		}
	}else {
		print (err)
	}
}