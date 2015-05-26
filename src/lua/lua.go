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

//#cgo LDFLAGS: -llua
/*
#include <stdio.h>
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

type wrapper struct {
	v          interface{}
	obj_type   reflect.Kind
	pointer    bool
	isFunction bool
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
		out := m.method.Call(in)
		ret = len(out)
		print("method Invoked ")
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
	C.lua_pushstring(L.s, C.CString(str))
}

func (L *State) PushInteger(n int) {
	C.lua_pushinteger(L.s, C.lua_Integer(n))
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
	w.pointer = reflect.ValueOf(val).Kind() == reflect.Ptr
	var k reflect.Kind
	if w.pointer {
		k = reflect.ValueOf(val).Elem().Kind()
	} else {
		k = reflect.ValueOf(val).Kind()
	}
	w.obj_type = k
	if (k != reflect.Slice) && (k != reflect.Struct) && (k != reflect.Map) {
		panic("The pushed interface can only be a Slice, Map or Struct")
	}
	C.pushObject(L.s, unsafe.Pointer(&w))
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
	C.lua_setfield(L.s, C.int(index), C.CString(k))
}

func (L *State) GetField(index int, k string) {
	C.lua_getfield(L.s, C.int(index), C.CString(k))
}

func (L *State) SetGlobal(name string) {
	C.lua_setglobal(L.s, C.CString(name))
}

func (L *State) GetGlobal(name string) int {
	return int(C.lua_getglobal(L.s, C.CString(name)))
}

func (L *State) SetMetaTable(index int) {
	C.lua_setmetatable(L.s, C.int(index))
}

func (L *State) GetMetaTable(index int) bool {
	return C.lua_getmetatable(L.s, C.int(index)) != 0
}

func (L *State) SetTable(index int) {
	C.lua_settable(L.s, C.int(index))
}

func (L *State) GetTable(index int) {
	C.lua_gettable(L.s, C.int(index))
}

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

	err := int(C.loadCodeSegment(L.s, C.CString(code)))

	if err != 0 {
		e := new(luaError)
		e.errStr = L.ToString(-1)
		L.Pop(1) /* pop error message from the stack */
		return e
	}
	return nil
}

func (L *State) PCall(nargs int, nresults int) error {
	err := int(C.callCode(L.s, C.int(nargs), C.int(nresults)))
	if err != 0 {
		e := new(luaError)
		e.errStr = L.ToString(-1)
		L.Pop(1) /* pop error message from the stack */
		return e
	}
	return nil
}

func (L *State) Error(err string) {
	C.doLuaError(L.s, C.CString(err))
}

//export go_callback_getter
func go_callback_getter(obj unsafe.Pointer, go_sate unsafe.Pointer) C.int {
	var ret int
	p := *(*wrapper)(obj)
	if !p.isFunction {
		val := p.v

		//To do any reflection we need to figure out the type
		temState := (*State)(go_sate)

		var itype reflect.Value
		if p.pointer {
			itype = reflect.ValueOf(val).Elem()
		} else {
			itype = reflect.ValueOf(val)
		}

		if temState.IsString(2) {
			lookFor := temState.ToString(2)
			print("Looking for 1 " + lookFor)
			field := itype.FieldByName(lookFor)
			//		print ("Looking for 2"+lookFor)
			if !field.IsValid() {
				//			print ("Looking for 3"+lookFor)
				var ptr reflect.Value
				var value reflect.Value
				var method reflect.Value

				if p.pointer {
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
					}
				}
			} else {
				temState.SetTop(0)
				ret = 1
				get_filed_value(field, temState)
			}
		} else {
			// Need to find a way to sort this out luaL_error does a longjmp which GO does not like and
			// Get a split stack panic
			// Given that GO threads are not based on ptheads its bit dangarous as well for now I am going push nil here
		}
	}
	return C.int(ret)
}

// Fix the types
func get_filed_value(v reflect.Value, L *State) {
	kind := v.Kind()
	switch kind {
	case reflect.String:
		L.PushString(v.String())
		break
	}
}

//export go_callback_setter
func go_callback_setter(obj unsafe.Pointer, go_sate unsafe.Pointer) C.int {
	var ret int
	p := *(*wrapper)(obj)
	if !p.isFunction {
		val := p.v

		ispointer := reflect.ValueOf(val).Kind() == reflect.Ptr

		//To do any reflection we need to figure out the type
		temState := (*State)(go_sate)

		var itype reflect.Value
		if ispointer {
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
			//		temState.Error("No valid filed/method specified")
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
	//To do any reflection we need to figure out the type
	temState := (*State)(go_sate)
	p := *(*wrapper)(obj)
	if p.isFunction {
		f := p.v.(GOLuaFunction)
		ret = f.Invoke(temState)
		//		print ("Call getting invoked "+temState.ToString(2))
	}
	return C.int(ret)
}

func handle_method(method reflect.Value, L *State) int {
	ic := new(methodInvoker)
	ic.method = method
	var w wrapper
	w.v = ic
	w.isFunction = true
	L.SetTop(0)
	C.pushObject(L.s, unsafe.Pointer(&w))
	return 1
}
