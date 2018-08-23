package golua

import (
	"sync"
)

var (
	patchMap  = make(map[string]int)
	patchLock sync.RWMutex
)

func RegisterPatch(luapath string) {
	patchLock.Lock()
	if _, ok := patchMap[luapath]; !ok {
		patchMap[luapath] = 0
	}
	patchLock.Unlock()
}

func ReLoadLuaScript(luapath string) {
	patchLock.Lock()
	if luapath == "*" {
		for key, val := range patchMap {
			val += 1
			patchMap[key] = val
			goout.Infof("[%s] Reload:%d", key, val)
		}
	} else {
		if val, ok := patchMap[luapath]; ok {
			val += 1
			patchMap[luapath] = val
			goout.Infof("[%s] Reload:%d", luapath, val)
		}
	}
	patchLock.Unlock()
}

func GetLuaPatch(luapath string) int {
	var patchVer int
	patchLock.RLock()
	if val, ok := patchMap[luapath]; ok {
		patchVer = val
	}
	patchLock.RUnlock()
	return patchVer
}
