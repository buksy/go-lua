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

import(
	"unsafe"
	"reflect"
)


type Lib int;

const (
	OS 		Lib = 5
	MATH 	Lib = 8
	PACKAGE Lib = 1
	IO		Lib	= 4
	TABLE	Lib = 3
	STRING	Lib = 6
	DEBUG	Lib = 9
	BASE 	Lib = 0
) 

type State struct {
	s *C.lua_State
}

type wrapper struct {
	v interface{}
	n string
}

type luaError struct {
	errStr string;
}

func (err *luaError) Error() string {
	return err.errStr
}

func NewState(loadDefaultLibs bool) (*State ,error){
	L := new (State)
	L.s = C.luaL_newstate()
	if L.s != nil {
		C.initNewState(L.s)
		if (loadDefaultLibs) {
			L.OpenLib (BASE)
			L.OpenLib (OS)
			L.OpenLib (IO)
			L.OpenLib (PACKAGE)
			L.OpenLib (MATH)
			L.OpenLib (STRING)
			L.OpenLib (TABLE)
			L.OpenLib (DEBUG)
		}
		return L, nil	
	}
	err := new (luaError)
	err.errStr = "Could not initialize the lua environment"
	return nil, err
}

func (L *State) Close() {
	C.lua_close (L.s)
}

func (L *State ) OpenLib(l Lib) {
	C.openDefaultLib(L.s ,C.int(int(l)))
}

// Check methods
func (L *State) IsNil(index int) bool		{
	 return int(C.lua_type(L.s, C.int(index))) == C.LUA_TNIL 
}

func (L *State) IsNone(index int) bool		{
	 return int(C.lua_type(L.s, C.int(index))) == C.LUA_TNONE 
}

func (L *State) IsNoneOrNil(index int) bool	{ 
	return int(C.lua_type(L.s, C.int(index))) <= 0
}

func (L *State) IsNumber(index int) bool	{
	 return C.lua_isnumber(L.s, C.int(index)) == 1 
 }

func (L *State) IsString(index int) bool	{
	return C.lua_isstring(L.s, C.int(index)) == 1 
}

func (L *State) IsTable(index int) bool		{
	return int(C.lua_type(L.s, C.int(index))) == C.LUA_TTABLE 
}

func (L *State) IsUserdata(index int) bool	{
	return C.lua_isuserdata(L.s, C.int(index)) == 1 
}

func (L *State) IsBoolean(index int) bool 	{
	return int(C.lua_type(L.s, C.int(index))) == C.LUA_TBOOLEAN 
}


// Push methods 
func (L *State) PushBoolean(b bool) {
	var bint int;
	if b {
		bint = 1;
	} else {
		bint = 0;
	}
	C.lua_pushboolean(L.s, C.int(bint));
}

func (L *State) PushString(str string) {
	C.lua_pushstring(L.s,C.CString(str));
}

func (L *State) PushInteger(n int) {
	C.lua_pushinteger(L.s,C.lua_Integer(n));
}

func (L *State) PushNil() {
	C.lua_pushnil(L.s);
}

func (L *State) PushNumber(n float64) {
	C.lua_pushnumber(L.s, C.lua_Number(n));
}

func (L *State) PushValue(index int) {
	C.lua_pushvalue(L.s, C.int(index));
}

func (L *State ) PushInterface(val interface{} ) {
	b := reflect.ValueOf(val).Kind() == reflect.Ptr
	print (b)
	var w wrapper
	w.v = val
	w.n = "gihan"
	C.pushObject (L.s, unsafe.Pointer(&w))
}

func (L *State) ToBoolean(index int) bool {
	return C.lua_toboolean(L.s, C.int(index)) != 0;
}

func (L *State) ToString(index int) string {
	return C.GoString(C.toString(L.s,C.int(index)));
}

func (L *State) ToInteger(index int) int {
	var i C.int
	return int(C.lua_tointegerx(L.s, C.int(index), &i));
}

func (L *State) ToNumber(index int) float64 {
	var i C.int
	return float64(C.lua_tonumberx(L.s, C.int(index), &i));
}

func (L *State) Type(index int) int {
	return int(C.lua_type(L.s, C.int(index)));
}

func (L *State) Typename(tp int) string {
	return C.GoString(C.lua_typename(L.s, C.int(tp)));
}

func (L *State) SetField(index int, k string) {
	C.lua_setfield(L.s, C.int(index), C.CString(k));
}

func (L *State) GetField(index int, k string) {
	C.lua_getfield(L.s, C.int(index), C.CString(k))
}

func (L *State) SetGlobal(name string) {
	C.lua_setglobal(L.s, C.CString(name))
}

func (L *State) GetGlobal(name string) int{
	return int(C.lua_getglobal(L.s, C.CString(name)))
}

func (L *State) SetMetaTable(index int) {
	C.lua_setmetatable(L.s, C.int(index));
}

func (L *State) GetMetaTable(index int) bool {
	return C.lua_getmetatable(L.s, C.int(index)) != 0
}

func (L *State) SetTable(index int) {
	C.lua_settable(L.s, C.int(index));
}

func (L *State) GetTable(index int)	{
	 C.lua_gettable(L.s, C.int(index)) 
 }

func (L *State) SetTop(index int) {
	C.lua_settop(L.s, C.int(index));
}

func (L *State) Pop(n int) {
	C.lua_settop(L.s, C.int(-n-1));
}

func (L *State) GetTop() int {
	 return int(C.lua_gettop(L.s)) 
}

// Loading the chunk 

func (L *State) LoadCodeString ( code string) error{
	
	err := int(C.loadCodeSegment(L.s, C.CString(code)))
	
	if (err != 0 ) {
		e := new(luaError)
		e.errStr = L.ToString(-1);
        L.Pop(1);  /* pop error message from the stack */
        return e
	}
	return nil
} 

func (L *State) PCall(nargs int, nresults int) error {
	err := int(C.callCode(L.s, C.int(nargs), C.int(nresults)));
	if (err != 0 ) {
		e := new(luaError)
		e.errStr = L.ToString(-1);
        L.Pop(1);  /* pop error message from the stack */
        return e
	}
	return nil
}

func (L *State) Error( err string ) {
	L.PushString(err)
	C.lua_error(L.s)
}

//export go_callback_getter
func go_callback_getter(obj unsafe.Pointer, s *C.lua_State) C.int {
	var ret int
	p := *(*wrapper)(obj)
	val := p.v
	//To do any reflection we need to figure out the type 
	temState := &State{}
	temState.s = s
	
	
	itype := reflect.ValueOf(val).Elem()
	if temState.IsString(2) {
		lookFor := temState.ToString(2)
		field := itype.FieldByName(lookFor)
		if !field.IsValid() {
			method := itype.MethodByName(lookFor)
			if !method.IsValid(){
				temState.Error("No filed/method named " + lookFor +" found")
			}else {
				//TODO: work on method reflection
			}
		}else {
			temState.SetTop(0)
			ret = 1
			get_filed_value(field, temState)
		}
	}else {
		temState.Error("No valid filed/method specified")
	}
	temState.s = nil
	return C.int(ret)
} 

// Fix the types 
func get_filed_value (v reflect.Value, L *State) {
	kind := v.Kind()
	switch (kind) {
		case reflect.String:
			L.PushString(v.String())
		break
	}
}

//export go_callback_setter
func go_callback_setter(obj unsafe.Pointer, s *C.lua_State) C.int {
	var ret int
	p := *(*wrapper)(obj)
	val := p.v
	//To do any reflection we need to figure out the type 
	temState := &State{}
	temState.s = s
	
	itype := reflect.ValueOf(val).Elem()
	if temState.IsString(2) {
		lookFor := temState.ToString(2)
		field := itype.FieldByName(lookFor)
		if !field.IsValid() || !field.CanSet() {
			temState.Error("No Filed named \"" + lookFor +"\" found")
		}else {
			ret = 0
			set_filed_value(field, temState)
			temState.SetTop(0)
		}
	}else {
		temState.Error("No valid filed/method specified")
	}
	temState.s = nil
	return C.int(ret)
} 

// Fix the types 
func set_filed_value (v reflect.Value, L *State) {
	kind := v.Kind()
	switch (kind) {
		case reflect.String:
			v.SetString(L.ToString(3))
		break
	}
}

