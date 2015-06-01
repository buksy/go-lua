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
	Map map[string]int
}

func (t *TestStruct) C(a int, b int) string{
	t.Gihan = "c"+ strconv.Itoa(a * b)
	return t.Gihan
	
}

func (t *TestStruct) B() string{
	t.Gihan = "B"
	return "B"
}

func (t *TestStruct) D() *TestStruct{
	s := new (TestStruct)
	s.Gihan = "Me"
	return s
}

func (t *TestStruct) E(a *TestStruct) {
	print (a.Gihan)
}

// testing an exported function 
type PrintFunc struct {
	
} 

func (f *PrintFunc) Name() string{
	return "myPrint"
}

func (f *PrintFunc) Invoke(L * lua.State) int{
//	print ("print from go " + strconv.Itoa(L.GetTop())+"  "+L.ToString(1))
	v := L.ToInterface(1)
	t := (v).(*TestStruct)
	print (t.Gihan)
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
	code := `
	json = require("json")
	function test(p)
a = myModule.myAdd(3, 3) 
p.B(4, 3) 
p.Map["test2"] = 2 
print((#p.Map)) 
for i, v in pairs(p) do
      print(i, v)
end
local x = p.D() 
print(p.E(x)) 
p.Test = "hello" 
return p.D() 
end`
	
	L, err := lua.NewState (true)
	if (err == nil) {
//		err = L.LoadCodeString ("local a = 10; return a + 20")
//		err = L.LoadCodeString ("function test(n) return n*n*n end")
		L.ExportGoFunction (new (PrintFunc))
		L.ExportGoModule (new (MyModule))
		
		// myPrint(p.E(p.D(),1))
		
		// print((#p.Map))
		// print(p.E(x)) p.Test = \"hello\" return p.D()
		err = L.LoadCodeString (code)
		L.SetTop(0)
		
		if (err == nil) {
			L.GetGlobal ("test")
			var t* TestStruct
			t = new(TestStruct)
			t.Gihan = "Hello"
			
			m := make (map[string]int)
			m["test1"] = 1
//			m["test2"] = 2
//			m["test3"] = 4
			t.Map = m
			
			L.PushInterface(t)
			
			print(L.ToInterface(-1).(*TestStruct).Gihan+"\n")
			err = L.PCall (1, 1)
			
			if err != nil {
				print (err.Error())
			}else {
				a := L.ToInterface(-1).(* TestStruct)
				//print ("hello")
				fmt.Printf("%s : %s : %s %d\n", a.Gihan , t.Test, t.Gihan, t.Map["test2"])	
			}
			defer L.Close()
		}else {
			print (err.Error())
		}
	}else {
		print (err)
	}
}