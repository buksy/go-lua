package main

import (
	"lua"
	"fmt"
	"strconv"
	"time"
)


//Test Struct and method invocation
type TestStruct struct {
	Gihan string
	Test string
	A 	 string
	Map map[string]int
	TS *TestStruct
}

func (t *TestStruct) C(a int, b int) string{
	t.Gihan = "c"+ strconv.Itoa(a * b)
	return t.Gihan
	
}

func (t *TestStruct) b() string{
	t.Gihan = "B"
	return "B"
}

func (t *TestStruct) D() *TestStruct{
	s := new (TestStruct)
	s.Gihan = "Me"
	return s
}

func (t *TestStruct) E(a *TestStruct, b string) {
	print (a.Gihan)
}

type Test struct {
	test string
	yo  string
}

func (t *Test) FromLUATable ( L *lua.State) error {
	L.PushString("test");
	L.GetTable(-2);
	if (L.IsString(-1)) {
		t.test = L.ToString(-1)
	}
	L.Pop(1);
	
	L.PushString("yo");
	L.GetTable(-2);
	if (L.IsString(-1)) {
		t.yo = L.ToString(-1)
	}
	L.Pop(1);
	
	return nil
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
var code = `
	--json = require("json")
	function test(p)
a = myModule.myAdd(3, 3) 
p.Map["test2"] = 2 
print((#p.Map)) 
for i, v in pairs(p) do
      print(i, v)
end
print(p.Map.test1)
p.A = "egf"
local x = p.d() 
print(p.TS.E(x,"hello")) 
x = p.d() 
x = nil
p.Test = "hello" 
return {test="hello " .. p.Map.test1, yo="hi"} 
end`
func main() {
	
	defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered in f", r)
        }
    }()
	for i:=0; i < 10000; i++ {
		a := i
		go do_work(a)
		time.Sleep(time.Second * 1)
//	}(i)
	}
	
	time.Sleep(time.Hour)
}


func do_work(a int) {

			time.Sleep(time.Second)
			L, err := lua.NewState (true)
			defer L.Close()
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
					t.A = "abc"
					m := make (map[string]int)
					m["test1"] = a
		//			m["test2"] = 2
		//			m["test3"] = 4
					t.Map = m
					
					t.TS = new (TestStruct)
					
					L.PushInterface(t)
					
					//print(L.ToInterface(-1).(*TestStruct).Gihan+"\n")
					err = L.PCall (1, 1)
					
					if err != nil {
						print (err.Error())
					}else {
						out := new (Test)
						L.ReadFormTable(out, -1)
						
						//a := L.ToInterface(-1).(* TestStruct)
						//print ("hello")
						fmt.Printf(out.test)	
					}
					
				}else {
					print (err.Error())
				}
			}else {
				print (err)
			}
//			L.Close()
}
