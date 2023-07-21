# inject
Go语言的依赖注入库

这是[codegangsta/inject](https://github.com/codegangsta/inject)的修改版本。
并参考了[flamego/flamego/inject](https://github.com/flamego/flamego/tree/main/inject)。

[English](./README.md)

## 使用方法

#### func InterfaceOf

```go
func InterfaceOf(value interface{}) reflect.Type
```
InterfaceOf函数将指向接口类型的指针解引用。如果value不是指向接口的指针，则会引发panic。

#### type Applicator

```go
type Applicator interface {
	// 将Type映射中的依赖项映射到标记为“inject”的结构体中的每个字段。
	// 如果注入失败，则返回错误。
	Apply(interface{}) error
}
```

Applicator表示将依赖项映射到结构体的接口。

#### type Injector

```go
// Injector表示将依赖项映射和注入到结构体和函数参数中的接口。
type Injector interface {
    Applicator
    Invoker
    TypeMapper
    // Reset将重置Injector，包括重置映射值和父级。
    Reset()
    // SetParent设置Injector的父级。如果Injector在其Type映射中找不到依赖项，
    // 则会在返回错误之前检查其父级。
    SetParent(Injector) Injector
}
```

Injector表示将依赖项映射和注入到结构体和函数参数中的接口。

#### func New

```go
func New() Injector
```
New函数返回一个新的Injector。

#### type Invoker

```go
type Invoker interface {
	// Invoke尝试调用提供的interface{}作为函数，
	// 根据Type为函数参数提供依赖项。返回表示函数返回值的reflect.Value切片。
	// 如果注入失败，则返回错误。
	Invoke(interface{}) ([]reflect.Value, error)
}
```

Invoker表示通过反射调用函数的接口。

#### type TypeMapper

```go
type TypeMapper interface {
	// 根据reflect.TypeOf的立即类型映射interface{}值。
	Map(interface{}) TypeMapper
	// 根据提供的Interface的指针映射interface{}值。
	// 这只对将值映射为接口有用，因为接口目前不能直接引用而不使用指针。
	MapTo(val interface{}, pointerToInterface interface{}) TypeMapper
	// 提供直接插入基于类型和值的映射的可能性。
	// 这使得可以直接映射使用reflect无法实例化的类型参数，例如单向通道。
	Set(reflect.Type, reflect.Value) TypeMapper
	// 返回映射到当前类型的Value。如果该类型尚未映射，则返回零值。
	Get(reflect.Type) reflect.Value
}
```

TypeMapper表示根据类型映射interface{}值的接口。