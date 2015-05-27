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
#include "luanative.h"
*/
import "C"

import (
	"reflect"
	"unsafe"
	//	"strconv"
)

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

type State struct {
	s *C.lua_State
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
}

type luaError struct {
	errStr string
}

type methodInvoker struct {
	method reflect.Value
}

func (m *methodInvoker) Invoke(L *State) int {
	in := make([]reflect.Value, m.method.Type().NumIn())
	var ret int
	if L.GetTop() >= len(in) {
		funcT := m.method.Type()
		for j := 0; j < len(in); j++ {
			t := funcT.In(j)
			d := reflect.New(t)
			luaToGo(L, d.Elem(), (j + 1))
			in[j] = d.Elem()
		}
		out := m.method.Call(in)
		ret = len(out)
		L.SetTop(0)
		for i := 0; i < ret; i++ {
			goToLua(L, out[i], (i + 1))
		}
	} else {
		print("method Invoked Failed not enough params")
		return 0
	}

	return ret
}

func (err *luaError) Error() string {
	return err.errStr
}

func NewState(loadDefaultLibs bool) (*State, error) {
	L := new(State)
	L.s = C.luaL_newstate()
	if L.s != nil {
		C.initNewState(L.s, unsafe.Pointer(L))
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
}

func (L *State) OpenLib(l Lib) {
	C.openDefaultLib(L.s, C.int(int(l)))
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
	defer C.free(unsafe.Pointer(cstr))
	C.lua_pushstring(L.s, cstr)
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
	var w wrapper
	w.v = val
	w.pointer = 0
	if reflect.ValueOf(val).Kind() == reflect.Ptr {
		w.pointer = 1
	}
	//	print (val)
	//	w.name = "Test"
	w.isFunction = 0
	var k reflect.Kind
	if w.pointer == 1 {
		//		print ("Pointer ....")
		k = reflect.ValueOf(val).Elem().Kind()
	} else {
		//		print ("Not Pointer")
		k = reflect.ValueOf(val).Kind()
	}
	w.obj_type = k
	if (k != reflect.Slice) && (k != reflect.Struct) && (k != reflect.Map) {
		panic("The pushed interface can only be a Slice, Map or Struct")
	}
	//	print ("---------------")
	//	print (reflect.ValueOf(&w).Pointer())
	//	C.pushObject(L.s, unsafe.Pointer(&w))
	C.pushObject(L.s, unsafe.Pointer(reflect.ValueOf(&w).Pointer()), 1)
	//	print (w.pointer)
}

func (L *State) pushFunction(f GOLuaFunction) {
	var w wrapper
	w.v = f
	w.isFunction = 1
	w.pointer = 0
	w.obj_type = reflect.Func
	C.pushFunction(L.s, unsafe.Pointer(reflect.ValueOf(&w).Pointer()))
}

func (L *State) ExportGoFunction(namedFunc GoExportedFunction) {
	L.pushFunction(namedFunc)
	L.SetGlobal(namedFunc.Name())
}

func (L *State) ExportGoModule(namedMod GoExportedModule) {
	var w wrapper
	w.v = namedMod
	w.isFunction = 1
	w.pointer = 0
	w.obj_type = reflect.Func
	C.pushObject(L.s, unsafe.Pointer(reflect.ValueOf(&w).Pointer()), 1)
	L.SetGlobal(namedMod.Name())
}

func (L *State) ToBoolean(index int) bool {
	return C.lua_toboolean(L.s, C.int(index)) != 0
}

func (L *State) ToString(index int) string {
	return C.GoString(C.toString(L.s, C.int(index)))
}

func (L *State) ToInteger(index int) int {
	var i C.int
	return int(C.lua_tointegerx(L.s, C.int(index), &i))
}

func (L *State) ToNumber(index int) float64 {
	var i C.int
	return float64(C.lua_tonumberx(L.s, C.int(index), &i))
}

func (L *State) Type(index int) int {
	return int(C.lua_type(L.s, C.int(index)))
}

func (L *State) Typename(tp int) string {
	return C.GoString(C.lua_typename(L.s, C.int(tp)))
}

func (L *State) SetField(index int, k string) {
	cstr := C.CString(k)
	defer C.free(unsafe.Pointer(cstr))
	C.lua_setfield(L.s, C.int(index), cstr)
}

func (L *State) GetField(index int, k string) {
	cstr := C.CString(k)
	defer C.free(unsafe.Pointer(cstr))
	C.lua_getfield(L.s, C.int(index), cstr)
}

func (L *State) SetGlobal(name string) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	C.lua_setglobal(L.s, cstr)
}

func (L *State) GetGlobal(name string) int {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	return int(C.lua_getglobal(L.s, cstr))
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

// Loading the chunk

func (L *State) LoadCodeString(code string) error {
	cstr := C.CString(code)
	defer C.free(unsafe.Pointer(cstr))

	err := int(C.loadCodeSegment(L.s, cstr))

	if err != 0 {
		e := new(luaError)
		e.errStr = L.ToString(-1)
		L.Pop(1) /* pop error message from the stack */
		return e
	}
	return nil
}

func (L *State) PCall(nargs int, nresults int) (e error) {

	defer func() {
		if pa := recover(); pa != nil {
			print(pa.(error).Error())
			e := new(luaError)
			e.errStr = pa.(error).Error()
		}
	}()

	err := int(C.callCode(L.s, C.int(nargs), C.int(nresults)))
	if err != 0 {
		er := new(luaError)
		er.errStr = L.ToString(-1)
		L.Pop(1) /* pop error message from the stack */
		return er
	}
	return
}

func (L *State) Error(err string) {
	print(err)
	str := C.CString(err)
	defer C.free(unsafe.Pointer(str))
	C.doLuaError(L.s, str)
}

//export go_callback_getter
func go_callback_getter(obj unsafe.Pointer, go_sate unsafe.Pointer) C.int {
	var ret int
	//	print ((*wrapper)(obj).isFunction)
	p := (*wrapper)(obj)
	//To do any reflection we need to figure out the type
	temState := (*State)(go_sate)
	//	print (p)
	if p.isFunction == 0 {
		val := p.v
		//		print (val)

		var itype reflect.Value
		if p.pointer == 1 {
			//			print ("Is pointer")
			itype = reflect.ValueOf(val).Elem()
		} else {
			//			print ("not pointer")
			itype = reflect.ValueOf(val)

		}

		if temState.IsString(2) {
			lookFor := temState.ToString(2)
			//			print("Looking for 1 " + lookFor)
			field := itype.FieldByName(lookFor)
			//		print ("Looking for 2"+lookFor)
			if !field.IsValid() {
				//			print ("Looking for 3"+lookFor)
				var ptr reflect.Value
				var value reflect.Value
				var method reflect.Value

				if p.pointer == 1 {
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
					return C.int(handle_method(method, temState))
				} else {
					method = value.MethodByName(lookFor)
					if method.IsValid() {
						return C.int(handle_method(method, temState))
					} else {
						// Need to find a way to sort this out luaL_error does a longjmp which GO does not like and
						// Get a split stack panic
						// Given that GO threads are not based on ptheads its bit dangarous as well for now I am going push nil here
						temState.Error("No method found")
					}
				}
			} else {
				temState.SetTop(0)
				ret = 1
				goToLua(temState, field, 1)
			}
		} else {
			// Need to find a way to sort this out luaL_error does a longjmp which GO does not like and
			// Get a split stack panic
			// Given that GO threads are not based on ptheads its bit dangarous as well for now I am going push nil here
			temState.Error("No valid filed/method specified")
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
func go_callback_setter(obj unsafe.Pointer, go_sate unsafe.Pointer) C.int {
	var ret int
	p := (*wrapper)(obj)
	if p.isFunction != 1 {
		val := p.v

		//To do any reflection we need to figure out the type
		temState := (*State)(go_sate)

		var itype reflect.Value
		if p.pointer == 1 {
			itype = reflect.ValueOf(val).Elem()
		} else {
			itype = reflect.ValueOf(val)
		}

		if temState.IsString(2) {
			lookFor := temState.ToString(2)
			field := itype.FieldByName(lookFor)
			if !field.IsValid() || !field.CanSet() {
				//			temState.Error("No Filed named \"" + lookFor + "\" found")
				temState.PushNil()
			} else {
				ret = 0
				set_filed_value(field, temState)
				temState.SetTop(0)
			}
		} else {
			temState.Error("No valid filed/method specified")
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
func go_callback_method(obj unsafe.Pointer, go_sate unsafe.Pointer) C.int {
	var ret int
	temState := (*State)(go_sate)
	p := (*wrapper)(obj)
	if p.isFunction == 1 {
		f := p.v.(GOLuaFunction)
		ret = f.Invoke(temState)
	}
	return C.int(ret)
}

func handle_method(method reflect.Value, L *State) int {
	ic := new(methodInvoker)
	ic.method = method
	var w wrapper
	w.v = ic
	w.isFunction = 1
	w.pointer = 0
	w.obj_type = reflect.Func
	L.SetTop(0)
	C.pushFunction(L.s, unsafe.Pointer(reflect.ValueOf(&w).Pointer()))
	return 1
}

func goToLua(L *State, val reflect.Value, idx int) {
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

	//	// underlying type is 'primitive' ? wrap it as a proxy!
	//	if isPrimitiveDerived(t, kind) != nil {
	//		makeValueProxy(L, val, cINTERFACE_META)
	//		return
	//	}

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
			L.PushInterface(val)
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
	//	print (L.Typename(idx))
	//	print("\n")
	//	print (L.GetTop())
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
			//				print(idx)
			//				print("-----------")
			//				print(val.Int())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			val.SetUint(uint64(L.ToNumber(idx)))
		}
	}
}
