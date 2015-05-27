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
	return t.Gihan
	
}

func (t *TestStruct) B() string{
	t.Gihan = "B"
	return "B"
}

func main() {
	L, err := lua.NewState (true)
	if (err == nil) {
//		err = L.LoadCodeString ("local a = 10; return a + 20")
//		err = L.LoadCodeString ("function test(n) return n*n*n end")
		err = L.LoadCodeString ("function test(p)  a = p.C(3, 3)  p.B(4, 3) p.Test = \"hello\" return a end")
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
				fmt.Printf("%s : %s : %s", a, t.Test, t.Gihan)	
			}
			defer L.Close()
		}else {
			print (err.Error())
		}
	}else {
		print (err)
	}
}