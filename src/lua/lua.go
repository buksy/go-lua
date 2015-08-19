/**
 * Copyright [2015] [Gihan Munasinghe ayeshka@gmail.com ]
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */
package lua

//#cgo LDFLAGS: -llua -lm -ldl
/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "luanative.h"
*/
import "C"


import (
	"fmt"
	"reflect"
	"unsafe"
	"unicode"
	"strings"
	"sync"
	//	"strconv"
)

func debug(str interface{}) {
//		print(str)
//		print ("\n")
}

var global_field_map = make(map[string]map[string]string)
var global_method_map = make (map[string]map[string]string)

type Lib int

const (
	OS      Lib = 5
	MATH    Lib = 8
	PACKAGE Lib = 1
	IO      Lib = 4
	TABLE   Lib = 3
	STRING  Lib = 6
	DEBUG   Lib = 9
	BASE    Lib = 0
)

type LuaTableReader interface {
	FromLUATable ( *State) error
}

type LuaTableWriter interface {
	ToLUATable ( *State) error
}

func is_valid_name(name string) bool {
	r := rune(name[0])
	return !unicode.IsLower(r)
}

func build_struct_map(t reflect.Type) {
//	debug ("Building struct map for "+t.Name())
	m := global_field_map[t.Name()]
	if m == nil {
		m := make(map[string]string)
		build_map_recursive(t, m)
		global_field_map[t.Name()] = m
//		fmt.Println(global_field_map)
	}
	method_map := global_method_map[t.Name()]
	if (method_map == nil) {
		method_map := make (map[string]string)
		for i := 0; i < t.NumMethod() ; i++ {
			method := t.Method(i)
			m_name := method.Name
			if (is_valid_name(m_name)) {
				method_map[strings.ToUpper(m_name)] = m_name
			}
			
		}
		global_method_map[t.Name()] = method_map
	}
}


func get_field_name(type_name string, fname string) string {
	var ret string
	ret = ""
//	fmt.Println ("Looking for "+type_name)
//	fmt.Println(global_field_map)
	m := global_field_map[type_name]
	if (m != nil) {
		ret = m[fname]
//		fmt.Println ("Looking for Field"+fname)
		if 	ret == "" {
			ret = m[strings.ToUpper(fname)]
//			fmt.Println ("Value for Field is "+ret)
		}
	}
	
	return ret
}

func get_method_name(type_name string, mname string) string {
	if !is_valid_name(mname) {
		tem := mname[1:len(mname)]
		c  := strings.ToUpper(string(mname[0]))
		return c+tem
	}
	return mname
//	m := global_method_map[type_name]
//	if (m != nil) {
//		return m[strings.ToUpper(mname)]	
//	}
//	return ""
}



func build_map_recursive(t reflect.Type, field_map map[string]string ) {
	for i := 0; i < t.NumField() ; i++ {
		field := t.Field(i)
		field_name := field.Name
//		debug("Filed is "+field_name+"\n")
		if (is_valid_name(field_name)) {
			kind := field.Type.Kind()
			if(kind == reflect.Struct && field.Type.Name() == field_name){
				build_map_recursive (field.Type, field_map)
			}else {
				// add the value in
//				debug("Filed added "+field_name+"\n")
				field_map[strings.ToUpper(field_name)] = field_name
				
				// Now support the JOSN tag format
				jsonTag := field.Tag.Get("json")
				if jsonTag != "" {			
					jsonName := strings.Split(jsonTag,",")[0]
					if jsonName != "" {
						field_map[jsonName] = field_name
					}
				}
			}
		}
	}
	
}

type State struct {
	s *C.lua_State
	// Any object are kept here till teh lua script is finished otherwise go will be garbadge collecting.
	// Go can not see what is coging on in the C side so Go will happly grabadge collect if the values is not refered
	obj_table map[int64]*wrapper
	curr_id int64
	lock *sync.Mutex
}

type GOLuaFunction interface {
	Invoke(L *State) int
}

//Implement this interface on a struct to export a go written function
type GoExportedFunction interface {
	GOLuaFunction
	Name() string
}

//Implement this interface on a struct to export a module witten in go on to lua
type GoExportedModule interface {
	ExportedFunctions() []GoExportedFunction
	Name() string
}

type wrapper struct {
	v          interface{}
	obj_type   reflect.Kind
	pointer    int
	isFunction int
	name       string
	id 		   int64
}

func (L *State) newWrapper() *wrapper{
	w := new(wrapper)
	L.lock.Lock()
	w.id = L.curr_id
	L.curr_id ++
	L.lock.Unlock()
	L.obj_table[w.id] = w
	//L.obj_table = append(L.obj_table, w)
	return w
}

type luaError struct {
	errStr string
}

type luaErrorI interface {
	error
}

type methodInvoker struct {
	method string
	value  interface{}
}

func (m *methodInvoker) Invoke(L *State) int {
	method, _ := get_method(m.value, m.method)
	in := make([]reflect.Value, method.Type().NumIn())
	var ret int
	//fmt.Printf("Get Top %s %d len %d\n",m.method , L.GetTop(), len(in))
	if L.GetTop() >= len(in) {
		funcT := method.Type()
		for j := 0; j < len(in); j++ {
			t := funcT.In(j)
			d := reflect.New(t)
			luaToGo(L, d.Elem(), (j + 1))
			in[j] = d.Elem()
		}
		debug(method.String())
		out := method.Call(in)
		ret = len(out)
		L.SetTop(0)
		for i := 0; i < ret; i++ {
			goToLua(L, out[i])
		}
	} else {
		str := fmt.Sprintf("Not enough arguments for call %s, require %d parameters call only supplied %d ", m.method, len(in), L.GetTop())
		L.Error(str)
//		debug("method Invoked Failed not enough params")
		return 0
	}

	return ret
}

func (err *luaError) Error() string {
	return err.errStr
}

func NewState(loadDefaultLibs bool) (*State, error) {
	L := new(State)
	L.obj_table  = make(map[int64]*wrapper)
	L.curr_id = 0
	L.lock = new(sync.Mutex)
	L.s = C.luaL_newstate()
	if L.s != nil {
		C.initNewState(L.s, unsafe.Pointer(reflect.ValueOf(L).Pointer()))
		if loadDefaultLibs {
			L.OpenLib(BASE)
			L.OpenLib(OS)
			L.OpenLib(IO)
			L.OpenLib(PACKAGE)
			L.OpenLib(MATH)
			L.OpenLib(STRING)
			L.OpenLib(TABLE)
			L.OpenLib(DEBUG)
		}
		return L, nil
	}
	err := new(luaError)
	err.errStr = "Could not initialize the lua environment"
	return nil, err
}

func (L *State) Close() {
	C.deinitState(L.s)
	C.lua_close(L.s)
//	L.s = nil
	L.obj_table = nil
}

func (L *State) OpenLib(l Lib) {
	C.openDefaultLib(L.s, C.int(int(l)))
}

func (L *State) LoadExternalModule(name string) error{
	L.GetGlobal("require")
	L.PushString(name)
	return L.PCall(1,1)
}

// Check methods
func (L *State) IsNil(index int) bool {
	return int(C.lua_type(L.s, C.int(index))) == C.LUA_TNIL
}

func (L *State) IsNone(index int) bool {
	return int(C.lua_type(L.s, C.int(index))) == C.LUA_TNONE
}

func (L *State) IsNoneOrNil(index int) bool {
	return int(C.lua_type(L.s, C.int(index))) <= 0
}

func (L *State) IsNumber(index int) bool {
	return C.lua_isnumber(L.s, C.int(index)) == 1
}

func (L *State) IsString(index int) bool {
	return C.lua_isstring(L.s, C.int(index)) == 1
}

func (L *State) IsTable(index int) bool {
	return int(C.lua_type(L.s, C.int(index))) == C.LUA_TTABLE
}

func (L *State) IsUserdata(index int) bool {
	return C.lua_isuserdata(L.s, C.int(index)) == 1
}

func (L *State) IsBoolean(index int) bool {
	return int(C.lua_type(L.s, C.int(index))) == C.LUA_TBOOLEAN
}

// Push methods
func (L *State) PushBoolean(b bool) {
	var bint int
	if b {
		bint = 1
	} else {
		bint = 0
	}
	C.lua_pushboolean(L.s, C.int(bint))
}

func (L *State) PushString(str string) {
	cstr := C.CString(str)
	C.lua_pushstring(L.s, cstr)
	C.free(unsafe.Pointer(cstr))
}

func (L *State) PushInteger(n int) {
	C.lua_pushinteger(L.s, C.lua_Integer(C.int(n)))
}

func (L *State) PushNil() {
	C.lua_pushnil(L.s)
}

func (L *State) PushNumber(n float64) {
	C.lua_pushnumber(L.s, C.lua_Number(n))
}

func (L *State) PushValue(index int) {
	C.lua_pushvalue(L.s, C.int(index))
}

func (L *State) PushInterface(val interface{}) {
	
	str, str_ok := val.(string)
	in, int_ok := val.(int64)
	fl, float_ok := val.(float64)
	if str_ok {
		L.PushString (str)
	}else if int_ok {
		L.PushInteger (int(in))
	}else if float_ok {
		L.PushNumber (fl)
	}else {
		w := L.newWrapper()
		w.v = val
		w.pointer = 0
		
		if reflect.ValueOf(val).Kind() == reflect.Ptr {
			w.pointer = 1
		}
		//	debug (val)
		//	w.name = "Test"
		w.isFunction = 0
		var k reflect.Kind
		var sType reflect.Type
		
		if w.pointer == 1 {
			//		debug ("Pointer ....")
			k = reflect.ValueOf(val).Elem().Kind()
			sType = reflect.ValueOf(val).Elem().Type()
		} else {
			//		debug ("Not Pointer")
			k = reflect.ValueOf(val).Kind()
			sType = reflect.ValueOf(val).Type()
		}
		w.obj_type = k
		debug(" Kind is " + k.String())
		if (k != reflect.Slice) && (k != reflect.Struct) && (k != reflect.Map) {
			panic("The pushed interface can only be a Slice, Map or Struct")
		}
		//	debug ("---------------")
		//	debug (reflect.ValueOf(&w).Pointer())
		//	C.pushObject(L.s, unsafe.Pointer(&w))
		if (k == reflect.Struct ) {
			w.name = sType.Name()
			build_struct_map (sType)
		}
	//	L.obj_table = append(L.obj_table, val)
		C.pushObject(L.s,  C.longlong(w.id), 1)
	//	debug (w.pointer)
	}
}

func (L *State) pushFunction(f GOLuaFunction) {
	w := L.newWrapper()
	w.v = f
	w.isFunction = 1
	w.pointer = 0
	w.obj_type = reflect.Func
	C.pushFunction(L.s, C.longlong(w.id))
}

func (L *State) ExportGoFunction(namedFunc GoExportedFunction) {
	L.pushFunction(namedFunc)
	L.SetGlobal(namedFunc.Name())
}

func (L *State) ExportGoModule(namedMod GoExportedModule) {
	w := L.newWrapper()
	w.v = namedMod
	w.isFunction = 1
	w.pointer = 0
	w.obj_type = reflect.Func
	C.pushObject(L.s, C.longlong(w.id), 1)
	L.SetGlobal(namedMod.Name())
}

func (L *State) ToBoolean(index int) bool {
	return C.lua_toboolean(L.s, C.int(index)) != 0
}

func (L *State) ToString(index int) string {
	str := C.toString(L.s, C.int(index))
	return C.GoString(str)
}

func (L *State) ToInteger(index int) int {
	var i C.int
	return int(C.lua_tointegerx(L.s, C.int(index), &i))
}

func (L *State) ToNumber(index int) float64 {
	var i C.int
	return float64(C.lua_tonumberx(L.s, C.int(index), &i))
}

func (L *State) ToInterface(index int) interface{} {
	debug("Calling to Interface\n")
	id := C.toUserData(L.s, C.int(index))
	if id > -1 {
		w := L.obj_table[int64(id)]
		debug(w.v)
		return w.v
	}
	return nil
}

func (L *State) Type(index int) int {
	return int(C.lua_type(L.s, C.int(index)))
}

func (L *State) Typename(tp int) string {
	return C.GoString(C.lua_typename(L.s, C.int(tp)))
}

func (L *State) SetField(index int, k string) {
	cstr := C.CString(k)
	C.lua_setfield(L.s, C.int(index), cstr)
	C.free(unsafe.Pointer(cstr))
}

func (L *State) GetField(index int, k string) {
	cstr := C.CString(k)
	C.lua_getfield(L.s, C.int(index), cstr)
	C.free(unsafe.Pointer(cstr))
}

func (L *State) SetGlobal(name string) {
	cstr := C.CString(name)
	C.lua_setglobal(L.s, cstr)
	C.free(unsafe.Pointer(cstr))
}

func (L *State) GetGlobal(name string) int {
	cstr := C.CString(name)
	i := int(C.lua_getglobal(L.s, cstr))
	C.free(unsafe.Pointer(cstr))
	return i
}

func (L *State) SetMetaTable(index int) {
	C.lua_setmetatable(L.s, C.int(index))
}

func (L *State) GetMetaTable(index int) bool {
	return C.lua_getmetatable(L.s, C.int(index)) != 0
}

//Table functions
func (L *State) NewTable() {
	C.lua_createtable(L.s, 0, 0)
}

func (L *State) Next(index int) int {
	return int(C.lua_next(L.s, C.int(index)))
}

func (L *State) SetTable(index int) {
	C.lua_settable(L.s, C.int(index))
}

func (L *State) GetTable(index int) {
	C.lua_gettable(L.s, C.int(index))
}

// Stack functions
func (L *State) SetTop(index int) {
	C.lua_settop(L.s, C.int(index))
}

func (L *State) Pop(n int) {
	C.lua_settop(L.s, C.int(-n-1))
}

func (L *State) GetTop() int {
	return int(C.lua_gettop(L.s))
}

func (L *State ) ReadFormTable( reader LuaTableReader, idx int) error {
	if (L.IsTable(idx)) {
		return reader.FromLUATable(L)
	}else {
		err := new(luaError)
		err.errStr = "Given index is not a LuaTable"
		return err
	}
} 

// Loading the chunk
func (L *State) LoadCodeString(code string, name string) error {
	cstr := C.CString(code)
    cname := C.CString(name)
	err := int(C.loadCodeSegment(L.s, cstr,cname))
    
    C.free(unsafe.Pointer(cstr))
    C.free(unsafe.Pointer(cname))
     
	if err != 0 {
		e := new(luaError)
		e.errStr = L.ToString(-1)
		L.Pop(1) /* pop error message from the stack */
		return e
	}else {
		return L.PCall(0, 0)
	}
	return nil
}

func (L *State) PCall(nargs int, nresults int) (err error) {
	//	defer func() {
	//		if r := recover(); r != nil {
	//            var ok bool
	//            err, ok = r.(error)
	//            if !ok {
	//          	   err = fmt.Errorf("Error on lua script %v", r)
	//            }
	//         }
	//	}()

	errval := int(C.callCode(L.s, C.int(nargs), C.int(nresults)))
	if errval != 0 {
		errStr := L.ToString(-1)
		err = fmt.Errorf("Error on lua script --> %s", errStr)
		L.Pop(1) /* pop error message from the stack */
	}

	return
}

func (L *State) Error(err string) {
	debug(err)
	C.luaL_where(L.s, 1)
	pos := L.ToString(-1)
	L.Pop(1)
	panic("Error on line " +pos + err)
}


func get_method(val interface{}, lookFor string) (reflect.Value, bool) {
	//			debug ("Looking for 3"+lookFor)
	var ptr reflect.Value
	var value reflect.Value
	var method reflect.Value

	if reflect.ValueOf(val).Kind() == reflect.Ptr {
		ptr = reflect.ValueOf(val)
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(val))
		temp := ptr.Elem()
		value = reflect.ValueOf(val)
		temp.Set(value)
	}

	//Check the method on the pointer
	method = ptr.MethodByName(lookFor)
	if method.IsValid() {
		return method, true
	} else {
		method = value.MethodByName(lookFor)
		if method.IsValid() {
			return method, true
		} else {
			return method, false
		}
	}
}

func goToLua(L *State, val reflect.Value) {
	if !val.IsValid() {
		L.PushNil()
		return
	}
	t := val.Type()
	if t.Kind() == reflect.Interface && !val.IsNil() { // unbox interfaces!
		val = reflect.ValueOf(val.Interface())
		t = val.Type()
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	kind := t.Kind()

	switch kind {
	case reflect.Float64, reflect.Float32:
		{
			L.PushNumber(val.Float())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			L.PushNumber(float64(val.Int()))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			L.PushNumber(float64(val.Uint()))
		}
	case reflect.String:
		{
			L.PushString(val.String())
		}
	case reflect.Bool:
		{
			L.PushBoolean(val.Bool())
		}
	case reflect.Slice, reflect.Map, reflect.Struct:
		{
			if (!val.IsNil()) {
				debug(" Pushing " + kind.String())
				L.PushInterface(val.Interface())
			} else {
				L.PushNil()
			}
		}

	default:
		{
			if val.IsNil() {
				L.PushNil()
			} else {
				L.PushInterface(val)
			}
		}
	}
}

func luaToGo(L *State, val reflect.Value, idx int) {
	kind := val.Kind()

	switch kind {
	case reflect.String:
		{
			if !L.IsNil(idx) {
				val.SetString(L.ToString(idx))
			} else {
				val.SetString("")
			}

		}
	case reflect.Float64, reflect.Float32:
		{
			val.SetFloat(L.ToNumber(idx))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			val.SetInt(int64(L.ToNumber(idx)))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			val.SetUint(uint64(L.ToNumber(idx)))
		}
	case reflect.Struct, reflect.Interface, reflect.Ptr:
		{
			val.Set(reflect.ValueOf(L.ToInterface(idx)))
		}
	}
	
}

/** Exported functions to C**/

//export go_callback_getter
func go_callback_getter(id int64, go_sate unsafe.Pointer) C.int {
	var ret int
	//	debug ((*wrapper)(obj).isFunction)
	//To do any reflection we need to figure out the type
	temState := (*State)(go_sate)
	p := temState.obj_table[id]
	//	debug (p)
	if p.isFunction == 0 {
		val := p.v

		var itype reflect.Value
		if p.pointer == 1 {
			itype = reflect.ValueOf(val).Elem()
		} else {
			itype = reflect.ValueOf(val)
		}

		switch p.obj_type {
		case reflect.Struct:
			{
				if temState.IsString(2) {
					lookFor := temState.ToString(2)
//					debug("Looking for 1 " + lookFor)
					
					fname := get_field_name(p.name, lookFor)
//					fmt.Println("Field is"+fname)
					field := itype.FieldByName(fname)
					
					//		debug ("Looking for 2"+lookFor)
					if !field.IsValid() {
						method_name := get_method_name(p.name,lookFor)
						_, ok := get_method(val, method_name)
						temState.SetTop(0)
						if ok {
							ic := new(methodInvoker)
							ic.method = method_name
							ic.value = val
							w := temState.newWrapper()
							w.v = ic
							w.isFunction = 1
							w.pointer = 0
							w.obj_type = reflect.Func
							temState.SetTop(0)
							C.pushFunction(temState.s, C.longlong(w.id))
							ret = 1
						} else {
							temState.Error("No method found \"" + lookFor + "\"")
						}
					} else {
						temState.SetTop(0)
						ret = 1
						goToLua(temState, field)
					}
				} else {
					// Need to find a way to sort this out luaL_error does a longjmp which GO does not like and
					// Get a split stack panic
					// Given that GO threads are not based on ptheads its bit dangarous as well for now I am going push nil here
					temState.Error("No valid filed/method specified")
				}
			}
		case reflect.Slice:
			{
				// If this is a slice second argument should be a int
				if temState.IsNumber(2) {
					idx := (temState.ToInteger(2) - 1)
					if idx >= 0 && idx < itype.Len() {
						retVal := itype.Index(idx)
						temState.SetTop(0)
						ret = 1
						goToLua(temState, retVal)
					} else {
						temState.SetTop(0)
						ret = 1
						temState.PushNil()
					}
				} else {
					temState.SetTop(0)
					ret = 1
					temState.PushNil()
				}
			}
		case reflect.Map:
			{
				debug("In maps")
				m := itype.Type()
				kt := m.Key()
				keyV := reflect.New(kt)
				luaToGo(temState, keyV.Elem(), 2)
				temState.SetTop(0)
				ret = 1
				retVal := itype.MapIndex(keyV.Elem())
				debug("Key is " + keyV.String())
				goToLua(temState, retVal)
			}
		}
	} else {
		// This is module
		if temState.IsString(2) {
			fn := temState.ToString(2)
			namedMod := p.v.(GoExportedModule)
			fList := namedMod.ExportedFunctions()
			var selectedF GoExportedFunction
			selectedF = nil
			for i := 0; i < len(fList); i++ {
				f := fList[i]
				if f.Name() == fn {
					selectedF = f
					break
				}
			}
			if selectedF != nil {
				temState.pushFunction(selectedF)
				ret = 1
			}
		}
	}
	return C.int(ret)
}

//export go_callback_setter
func go_callback_setter(id int64, go_sate unsafe.Pointer) C.int {
	var ret int
	temState := (*State)(go_sate)
	p := temState.obj_table[id]
	if p.isFunction != 1 {
		val := p.v

		//To do any reflection we need to figure out the type
		var itype reflect.Value
		if p.pointer == 1 {
			itype = reflect.ValueOf(val).Elem()
		} else {
			itype = reflect.ValueOf(val)
		}

		switch p.obj_type {
		case reflect.Struct:
			{
				if temState.IsString(2) {
					lookFor := temState.ToString(2)
					fname := get_field_name(p.name, lookFor)
					field := itype.FieldByName(fname)
					if !field.IsValid() || !field.CanSet() {
						//			temState.Error("No Filed named \"" + lookFor + "\" found")
						temState.PushNil()
					} else {
						ret = 0
						luaToGo(temState, field, 3)
						temState.SetTop(0)
					}
				} else {
					temState.Error("No valid filed/method specified")
				}
			}
		case reflect.Slice:
			{
				idx := temState.ToInteger(2) - 1
				newVal := reflect.New(itype.Type().Elem())
				luaToGo(temState, newVal.Elem(), 3)
				if idx < 0 || idx > itype.Len() {
					itype.Index(idx).Set(newVal)
				}
			}
		case reflect.Map:
			{
				m := itype.Type()
				kt := m.Key()
				keyV := reflect.New(kt)
				vt := m.Elem()
				vV := reflect.New(vt)
				luaToGo(temState, keyV.Elem(), 2)
				luaToGo(temState, vV.Elem(), 3)
				itype.SetMapIndex(keyV.Elem(), vV.Elem())
			}
		}
	}
	return C.int(ret)
}

// Fix the types
func set_filed_value(v reflect.Value, L *State) {
	kind := v.Kind()
	switch kind {
	case reflect.String:
		v.SetString(L.ToString(3))
		break

	}
}

//export go_callback_method
func go_callback_method(id int64, go_sate unsafe.Pointer) C.int {
	var ret int
	temState := (*State)(go_sate)
	p := temState.obj_table[id]
	if p.isFunction == 1 {
		f := p.v.(GOLuaFunction)
		debug("f : ")
		debug(f)
		debug("\n")
		ret = f.Invoke(temState)
		a,ok := f.(*methodInvoker) // Methods invoked does not need to stay in memory
		if ok {
//			fmt.Printf("%d : %v \n", p.id, a.method )
			a.value = nil	
			delete(temState.obj_table, p.id)
		}
	}
	return C.int(ret)
}

//export go_callback_len
func go_callback_len (id int64, go_sate unsafe.Pointer) C.int {
	var ret int
	ret = 1
	temState := (*State)(go_sate)
	temState.SetTop(0)
	p := temState.obj_table[id]
	if p.isFunction == 0 {
		var val reflect.Value
		if p.pointer == 1 {
			val = reflect.ValueOf(p.v).Elem()
		} else {
			val = reflect.ValueOf(p.v)
		}
		switch p.obj_type {
			case reflect.Slice, reflect.Map: {
				temState.PushInteger(val.Len())
			}
			case reflect.Struct: {
				temState.PushInteger(val.NumField())
			}
			default :{
				temState.PushInteger(0)
			}
		}
	}
	return C.int(ret)
}


type loopStruct struct {
	v          	interface{}
	obj_type   	reflect.Kind
	pointer    	int
	ip_pairs    int
	current_idx int
}

func (p *loopStruct) Invoke(L *State) int {
		ret := 1
		var val reflect.Value
		if p.pointer == 1 {
			val = reflect.ValueOf(p.v).Elem()
		} else {
			val = reflect.ValueOf(p.v)
		}
		pairs := p.ip_pairs == 0
		switch p.obj_type {
			case reflect.Slice, reflect.Map: {
				max := val.Len()
				var retVal reflect.Value
				current_idx := 0
				current_idx = p.current_idx
				
				if current_idx >= max {
					L.PushNil()
				}else {
					if p.obj_type == reflect.Slice {
						retVal = val.Index(current_idx)
						L.PushInteger((current_idx + 1))
					}else {
						k := val.MapKeys()[current_idx]
						if (pairs) {
							goToLua(L, k)
						}else {
							L.PushInteger(current_idx)
						}
						retVal = val.MapIndex(k)
					}
				}
				current_idx ++
				p.current_idx = current_idx
				goToLua(L, retVal)
				ret = 2
			}
			case reflect.Struct: {
				max := val.NumField() 
				current_idx := 0
				if (!pairs) {
					current_idx = L.ToInteger(2)
				}else {
					current_idx = p.current_idx
				}
				
				if current_idx >= max {
					L.PushNil()
				} else {
					f := val.Field(current_idx)
					current_idx ++
					p.current_idx = current_idx
					if (pairs) {
						L.PushString(val.Type().Field(current_idx-1).Name)
					}else {
						L.PushInteger(current_idx)
					}
					goToLua(L, f)
					ret = 2
				}
			}
			default :{
				L.PushNil()
			}
		}
	return ret
}

func clone_wrapper (L *State , ori *wrapper) *wrapper {
	w := L.newWrapper()
	w.isFunction = ori.isFunction
	w.name = ori.name
	w.obj_type = ori.obj_type
	w.pointer = ori.pointer
	return w
}

//export go_callback_pairs
func go_callback_pairs (id int64, go_sate unsafe.Pointer) C.int {
	var ret int
	ret = 2
	temState := (*State)(go_sate)
	temState.SetTop(0)
	p := temState.obj_table[id]
	if p.isFunction == 0 {
		loop := new (loopStruct)
		loop.v = p.v
		loop.pointer = p.pointer
		loop.current_idx = 0
		loop.obj_type = p.obj_type
		temState.pushFunction(loop)
		loop.ip_pairs = 0
//		p = clone_wrapper(temState,p)
		C.pushObject(temState.s, C.longlong(p.id), 1)
		temState.PushNil()
		ret = 3
	}
	return C.int(ret)
}

//export go_callback_ipairs
func go_callback_ipairs (id int64, go_sate unsafe.Pointer) C.int {
	var ret int
	ret = 1
	temState := (*State)(go_sate)
	temState.SetTop(0)
	p := temState.obj_table[id]
	if p.isFunction == 0 {
		loop := new (loopStruct)
		loop.v = p.v
		loop.pointer = p.pointer
		loop.current_idx = 0
		loop.obj_type = p.obj_type
		temState.pushFunction(loop)
		loop.ip_pairs = 1
//		p = clone_wrapper(temState,p)
		C.pushObject(temState.s, C.longlong(p.id), 1)
		temState.PushInteger(0)
		ret = 3
	}
	return C.int(ret)
}

//export go_cleanup
func go_cleanup (id int64, go_sate unsafe.Pointer) {
//	fmt.Printf("Removing id %d \n",(id))
	temState := (*State)(go_sate)
	temState.SetTop(0)
	p := temState.obj_table[id]
	if p != nil {
		
//		fmt.Printf("Removing 2 id  %d : %v \n",(id), p.v)
		delete(temState.obj_table, id)
		p.v = nil
	}
}

