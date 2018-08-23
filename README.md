Go Bindings for the lua C API
=========================

快速开始
---------------------

创建一个Lua虚拟机

```go
L := lua.NewState()
L.OpenLibs()
defer L.Close()
```

Lua虚拟机是基于栈实现的，我们可以这样调用函数:

```go
// push "print" function on the stack
L.GetField(lua.LUA_GLOBALSINDEX, "print")
// push the string "Hello World!" on the stack
L.PushString("Hello World!")
// call print with one argument, expecting no results
L.Call(1, 0)
```

lua语句块、lua脚本文件

```go
// executes a string of lua code
err := L.DoString("...")
// executes a file
err = L.DoFile(filename)
```

导出Go函数到Lua

```go
func adder(L *lua.State) int {
	a := L.ToInteger(1)
	b := L.ToInteger(2)
	L.PushInteger(a + b)
	return 1 // number of return values
}

func main() {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	L.Register("adder", adder)
	L.DoString("print(adder(2, 2))")
}
```

SEE ALSO
---------------------

- [Luar](https://github.com/stevedonovan/luar/) is a reflection layer on top of golua API providing a simplified way to publish go functions to a Lua VM.
- [Golua](https://github.com/aarzilli/golua) is go bindings for Lua C API
