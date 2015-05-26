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

#ifndef _H_LUA_NATIVE

#define _H_LUA_NATIVE
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>

void openDefaultLib (lua_State *L,  int openlib);

int callCode (lua_State *L , int nargs, int retargs);

const char *toString (lua_State *L , int idx);

int loadCodeSegment(lua_State *L, const char *code);

void pushObject(lua_State *L, void *obj) ;

void initNewState(lua_State *L, void *go_stae) ;

void deinitState (lua_State *L);

void doLuaError (lua_State *L, const char * errorMsg);

#endif
