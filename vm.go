/*
lua 虚拟机模块
*/
package golua

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"

	"utils"

	"golua/lua"
	"golua/luar"
)

var INIT_VM_GC_LIMIT_COUNT int = 2000

type LuaVM struct {
	// lua virtual machine
	L *lua.State

	VmSize int // 初始化后的内存大小(Kb)
	VmSeq  int // 虚拟机编号

	PatchVer int // 热更补丁版本
	UseCount int // VM使用一定次数后GC

	luaLoad string // Lua脚本路径

	luaCallObj map[string]*luar.LuaObject
}

func (vm *LuaVM) setLuaPath(rootPath string) {
	// path
	var luapath = path.Join(rootPath, "?.lua")
	luapath += ";" + path.Join(rootPath, "?", "init.lua")
	//	luapath += ";" + path.Join(rootPath, "lualib", "?.lua")

	// cpath
	var luacpath = ""
	if runtime.GOOS == "windows" {
		luacpath = path.Join(rootPath, "luaclib", "?.dll")
	} else if runtime.GOOS == "linux" {
		luacpath = path.Join(rootPath, "luaclib", "?.so")
	}

	vm.Register("package", luar.Map{
		"cpath": luacpath,
		"path":  "./?.lua;" + luapath,
	})
}

func (vm *LuaVM) setGLua() {
	vm.Register("glua", luar.Map{
		"HotFix":     ReLoadLuaScript,
		"GetUnixF":   utils.GetUnixF,
		"Sleep":      utils.Sleep,
		"MillSleep":  utils.MillSleep,
		"EnJson":     utils.EnJson,
		"DeJson":     utils.DeJson,
		"Print":      luaout.Log,
		"PrintDebug": luaout.Debugln,
		"PrintInfo":  luaout.Infoln,
		"PrintError": luaout.Errorln,
	})
}

func (vm *LuaVM) Status() string {
	if vm.L == nil {
		return "LuaVM `L` is nil"
	}

	var status = "LuaVM Status:\n"
	status += fmt.Sprintf("\tVmSize:%d\n", vm.VmSize)
	status += fmt.Sprintf("\tVmSeq:%d\n", vm.VmSeq)
	status += fmt.Sprintf("\tPatchVer:%d\n", vm.PatchVer)
	status += fmt.Sprintf("\tUseCount:%d\n", vm.UseCount)

	return status
}

func (vm *LuaVM) Init() {
	if vm.L != nil {
		vm.L.Close()
	}

	vm.L = luar.Init()

	vm.setGLua()

	vm.VmSize = vm.L.GC(lua.LUA_GCCOUNT, 0)
	vm.UseCount = 0
	vm.luaCallObj = make(map[string]*luar.LuaObject)
}

func (vm *LuaVM) Register(tName string, tValue luar.Map) {
	luar.Register(vm.L, tName, tValue)
	vm.VmSize = vm.L.GC(lua.LUA_GCCOUNT, 0)
}

func (vm *LuaVM) LoadScript(sLuaPath string) error {
	// 检查是不是lua文件
	if strings.HasSuffix(sLuaPath, ".lua") {
		// set lua path
		var base = path.Dir(sLuaPath)
		vm.setLuaPath(base)

		// load luafile
		vm.luaLoad = sLuaPath
		err := vm.L.DoFile(sLuaPath) // vm.L.LoadFile(sLuaPath) //
		if err != nil {
			return errors.New(fmt.Sprintf("LoadScript[%s] Err:\n %s", sLuaPath, err))
		}

		vm.VmSize = vm.L.GC(lua.LUA_GCCOUNT, 0)

		// 保存 luaPath 热更使用
		RegisterPatch(sLuaPath)
	}

	return nil
}

func (vm *LuaVM) CallMethod(objName string, funName string, args ...interface{}) (error, interface{}) {
	if vm.L == nil {
		return errors.New("vm.L is nil"), nil
	}

	// 是否需要GC
	vm.UseCount++
	//	if vm.UseCount > INIT_VM_GC_LIMIT_COUNT {
	//		// GC
	//		vm.L.GC(lua.LUA_GCCOLLECT, 0)
	//		vm.UseCount = 0
	//	}

	var callName = funName
	if len(objName) > 0 {
		callName = fmt.Sprintf("%s.%s", objName, funName)
	}

	var luaObj *luar.LuaObject
	if obj, ok := vm.luaCallObj[callName]; ok {
		luaObj = obj
	} else {
		luaObj = luar.NewLuaObjectFromName(vm.L, callName)
		if luaObj != nil {
			vm.luaCallObj[callName] = luaObj
		}
	}
	if luaObj == nil {
		var errInfo = fmt.Sprintf("LuaMethod `%s` not exist", callName)
		return errors.New(errInfo), nil
	}

	var callResult []interface{}
	var err = luaObj.Call(&callResult, args...)

	//	luaObj.Close()

	vm.VmSize = vm.L.GC(lua.LUA_GCCOUNT, 0)
	return err, callResult
}
