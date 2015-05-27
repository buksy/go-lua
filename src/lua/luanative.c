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

#include <stdio.h>
#include <stdlib.h>

#include "luanative.h"
#include "_cgo_export.h"

#define GO_LUA_OBJECT		"buksy.go.lua"
#define GO_SATE 	  		"buksy.go.state"

typedef struct GoObject {
	void *go;
}GoObject;

void openDefaultLib (lua_State *L,  int openlib) {

	const char *libname;
	lua_CFunction openfunc;

	switch (openlib) {
	case 0:
		libname = "_G";
		openfunc = luaopen_base;
		break;
	case 1:
		libname = LUA_LOADLIBNAME;
		openfunc = luaopen_package;
		break;
	case 2:
		libname = LUA_COLIBNAME;
		openfunc = luaopen_coroutine;
		break;
	case 3:
		libname = LUA_TABLIBNAME;
		openfunc = luaopen_table;
		break;
	case 4:
		libname = LUA_IOLIBNAME;
		openfunc = luaopen_io;
		break;
	case 5:
		libname = LUA_OSLIBNAME;
		openfunc = luaopen_os;
		break;
	case 6:
		libname = LUA_STRLIBNAME;
		openfunc = luaopen_string;
		break;
	case 7:
		libname = LUA_BITLIBNAME;
		openfunc = luaopen_bit32;
		break;
	case 8:
		libname = LUA_MATHLIBNAME;
		openfunc = luaopen_math;
		break;
	case 9:
		libname = LUA_DBLIBNAME;
		openfunc = luaopen_debug;
		break;

	}
	luaL_requiref(L, libname, openfunc, 1);
}

int callCode (lua_State *L , int nargs, int retargs) {
	int ret = lua_pcall(L, nargs, retargs, 0);
	return ret;
}

const char *toString (lua_State *L , int idx) {
	return lua_tostring(L, idx);
}

int loadCodeSegment(lua_State *L, const char *code) {
	return luaL_dostring (L, code);
}

void pushObject(lua_State *L, void *obj) {
	GoObject *o = lua_newuserdata (L, sizeof(GoObject));
	o->go = obj;
	luaL_getmetatable (L, GO_LUA_OBJECT);
	lua_setmetatable (L, -2);
}

void pushFunction(lua_State *L, void *obj) {
	GoObject *o = lua_newuserdata (L, sizeof(GoObject));
	o->go = obj;
	luaL_getmetatable (L, GO_LUA_OBJECT);
	lua_setmetatable (L, -2);
//	fprintf(stderr, "go_index Looking for %d\n",lua_gettop(L));
}

static int gc_goobj (lua_State * L) {
	GoObject *obj = (GoObject *) luaL_checkudata (L, 1, GO_LUA_OBJECT);
	if (obj) {
		obj->go = NULL;
	}
	return 0;
}

static GoObject * get_go_state(lua_State *L) {
	lua_getfield(L, LUA_REGISTRYINDEX, GO_SATE);
	if (!lua_isuserdata(L, -1)) {
			/* Java state has been cleared as the Java VM was destroyed. Cannot call. */
		lua_pushliteral(L, "no go state found");
		lua_error(L);
		return NULL;
	}
	GoObject  *go_sate = (GoObject *) lua_touserdata(L, -1);
	lua_pop(L, 1);
	return go_sate;
}

static int go_index (lua_State * L) {

	GoObject *go_sate = get_go_state(L);

	GoObject *obj = (GoObject *) luaL_checkudata (L, 1, GO_LUA_OBJECT);
	int ret = 0;
	if (obj) {
//		fprintf(stderr, "go_index Looking for %s\n",toString(L, 2));
		ret = go_callback_getter(obj->go, go_sate->go);
	}
	return ret;
}

static int go_new_index (lua_State * L) {

	GoObject *go_sate = get_go_state(L);

	GoObject *obj = (GoObject *) luaL_checkudata (L, 1, GO_LUA_OBJECT);
	int ret = 0;
	if (obj) {
//		fprintf(stderr, " go_new_index Looking for %s\n",toString(L, 2));
		ret = go_callback_setter(obj->go, go_sate->go);
	}
	return ret;
}

static int go_call (lua_State * L) {

	GoObject *go_sate = get_go_state(L);

	GoObject *obj = (GoObject *) luaL_checkudata (L, 1, GO_LUA_OBJECT);
	int ret = 0;
	if (obj) {
//		fprintf(stderr, " go_new_index Looking for %d\n",lua_gettop(L));
		lua_remove(L,1);
		ret = go_callback_method(obj->go, go_sate->go);
	}
	return ret;
}

static int go_lua_atpanic(lua_State *L) {
//	fprintf(stderr, " hahaha Panic %s",toString(L, -1));
	return 0;
}

void initNewState(lua_State *L, void *go_stae) {
	lua_atpanic(L, &go_lua_atpanic);
	/* Set the go state state in the Lua state. */
	GoObject *ref = lua_newuserdata(L, sizeof(GoObject));
	ref->go = go_stae;
	lua_createtable(L, 0, 1);
	lua_pushcfunction(L, gc_goobj);
	lua_setfield(L, -2, "__gc");
	lua_setmetatable(L, -2);
	lua_setfield(L, LUA_REGISTRYINDEX, GO_SATE);

	// Meta table for struct
	luaL_newmetatable(L, GO_LUA_OBJECT);
	lua_pushboolean(L, 0);
	lua_setfield(L, -2, "__metatable");

	lua_pushcfunction(L, gc_goobj);
	lua_setfield(L, -2, "__gc");

	lua_pushcfunction(L, go_index);
	lua_setfield(L, -2, "__index");

	lua_pushcfunction(L, go_new_index);
	lua_setfield(L, -2, "__newindex");

	lua_pushcfunction(L, go_call);
	lua_setfield(L, -2, "__call");
}

void deinitState(lua_State *L ) {
	lua_pushnil(L);
	lua_setfield(L, LUA_REGISTRYINDEX, GO_SATE);
}

void doLuaError (lua_State *L, const char * errorMsg){
	luaL_error(L, errorMsg);
}
