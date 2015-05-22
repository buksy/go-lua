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

void initNewState(lua_State *L) ;
#endif
