package main

import (
	"lua"
	"fmt"
	"strconv"
)


//Test Struct and method invocation
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

// testing an exported function 
type PrintFunc struct {
	
} 

func (f *PrintFunc) Name() string{
	return "myPrint"
}

func (f *PrintFunc) Invoke(L * lua.State) int{
	print ("print from go" +L.ToString(1) +" \n")
	return 0
}

type AddFnc struct {
	
} 

func (f *AddFnc) Name() string{
	return "myAdd"
}

func (f *AddFnc) Invoke(L * lua.State) int{
	a := L.ToInteger(1) + L.ToInteger(2)
	L.PushInteger(a)
	return 1
}

// testing exported modules

type MyModule struct {
	
} 

func (f *MyModule) Name() string{
	return "myModule"
}


func (f *MyModule) ExportedFunctions() []lua.GoExportedFunction{
	v := []lua.GoExportedFunction{new(PrintFunc), new (AddFnc)}
	return v
}

func main() {
	L, err := lua.NewState (true)
	if (err == nil) {
//		err = L.LoadCodeString ("local a = 10; return a + 20")
//		err = L.LoadCodeString ("function test(n) return n*n*n end")
		L.ExportGoFunction (new (PrintFunc))
		L.ExportGoModule (new (MyModule))
		err = L.LoadCodeString ("function test(p) myPrint(\"begin\") a = myModule.myAdd(3, 3)  p.B(4, 3) p.Test = \"hello\" return a end")
		L.SetTop(0)
		if (err == nil) {
			L.GetGlobal ("test")
			var t* TestStruct
			t = new(TestStruct)
			t.Gihan = "Hello"
			L.PushInterface(t)
			err = L.PCall (1, 1)
			
			if err != nil {
				print ("hello e")
				print (err.Error())
			}else {
				a := L.ToInteger(-1)
				print ("hello")
				fmt.Printf("%d : %s : %s", a, t.Test, t.Gihan)	
			}
			defer L.Close()
		}else {
			print (err.Error())
		}
	}else {
		print (err)
	}
}