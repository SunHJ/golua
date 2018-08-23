package golua

import (
	"errors"

	"golua/luar"
)

var INIT_VM_RESTART_LIMIT_SIZE int = 1024 // Kb

type LuaProxy struct {
	chVm chan *LuaVM
	Size int

	Token int64

	luaRoot string

	go2lua map[string]luar.Map
}

// 创建一个虚拟机 并加载脚本
func (lp *LuaProxy) createVM(nVmNo int) (*LuaVM, error) {
	var vm = new(LuaVM)
	vm.VmSeq = nVmNo

	vm.Init()

	// 导出的Go接口
	for kName, vValue := range lp.go2lua {
		vm.Register(kName, vValue)
	}

	// 加载脚本
	var err error
	if len(lp.luaRoot) > 0 {
		err = vm.LoadScript(lp.luaRoot)
		vm.PatchVer = GetLuaPatch(lp.luaRoot)
	}

	return vm, err
}

func (lp *LuaProxy) RegLuaTable(tName string, tValue luar.Map) {
	if lp.go2lua == nil {
		lp.go2lua = make(map[string]luar.Map)
	}

	if vMap, ok := lp.go2lua[tName]; ok && tName == "" {
		for k, v := range tValue {
			vMap[k] = v
		}
	} else {
		lp.go2lua[tName] = tValue
	}
}

func (lp *LuaProxy) ProxyInit(nSize int, sLuaRoot string) {
	lp.Size = nSize
	lp.luaRoot = sLuaRoot

	if lp.chVm != nil {
		close(lp.chVm)
	}
	lp.chVm = make(chan *LuaVM, nSize+1)
	for i := 0; i < nSize; i++ {
		// 创建
		vm, err := lp.createVM(i)
		if err != nil {
			goout.Errorf("LuaProxy Init[%s] Err:%s", sLuaRoot, err)
			break
		}

		lp.chVm <- vm
	}

	//	vm, err := lp.createVM(1)
	//	if err != nil {
	//		goout.Errorf("LuaProxy Init[%s] Err:%s", sLuaRoot, err)
	//	}
	//	lp.vm = vm
}

func (lp *LuaProxy) CallMethod(token int64, objName string, funName string, args ...interface{}) (error, interface{}) {
	//	if lp.vm == nil {
	//		return errors.New("vm is nil"), nil
	//	}
	//	var vm = lp.vm

	var vm = <-lp.chVm
	if vm == nil {
		return errors.New("vm is nil"), nil
	}

	// Reload
	var curPatch = GetLuaPatch(lp.luaRoot)
	if curPatch != vm.PatchVer {
		var e error
		vm, e = lp.createVM(vm.VmSeq)
		if e != nil {
			goout.Errorf("LuaProxy Reload Err:%s", e)
		}
	}

	// check token
	if lp.Token == token {
		var err, reval = vm.CallMethod(objName, funName, args...)
		if err == nil {
			if vm.VmSize >= INIT_VM_RESTART_LIMIT_SIZE {
				var e error
				vm, e = lp.createVM(vm.VmSeq)
				if e != nil {
					goout.Errorf("LuaProxy Reset Err:%s", e)
				}
			}
		}
		return err, reval
	}

	lp.chVm <- vm

	return nil, nil
}

func (lp *LuaProxy) CallGlobalMethod(token int64, name string, args ...interface{}) (error, interface{}) {
	return lp.CallMethod(token, "", name, args...)
}
