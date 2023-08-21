package inject

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"unsafe"
)

type specialString interface{}

type testStruct struct {
	Dep1 string        `inject:"" json:"-"`
	Dep2 specialString `inject:""`
	Dep3 string
}

type greeter struct {
	Name string
}

func (g *greeter) String() string {
	return "Hello, My name is" + g.Name
}

/* Test Helpers */
func expect(t testing.TB, actual interface{}, expect interface{}) {
	if actual != expect {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", expect, reflect.TypeOf(expect), actual, reflect.TypeOf(actual))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

type myFastInvoker func(string)

func (myFastInvoker) Invoke([]interface{}) ([]reflect.Value, error) {
	return nil, nil
}

func TestInjectorSize(t *testing.T) {
	// prevent unnecessary memory usage increases
	expect(t, unsafe.Alignof(injector{}), uintptr(8))
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	var j Injector
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j = New()
	}
	_ = j
}

func TestInjector_Invoke(t *testing.T) {
	t.Parallel()

	t.Run("invoke functions", func(t *testing.T) {
		inj := New()

		dep := "some dependency"
		inj.Map(dep)
		dep2 := "another dep"
		inj.MapTo(dep2, (*specialString)(nil))
		dep3 := make(chan *specialString)
		dep4 := make(chan *specialString)
		typRecv := reflect.ChanOf(reflect.RecvDir, reflect.TypeOf(dep3).Elem())
		typSend := reflect.ChanOf(reflect.SendDir, reflect.TypeOf(dep4).Elem())
		inj.Set(typRecv, reflect.ValueOf(dep3))
		inj.Set(typSend, reflect.ValueOf(dep4))

		_, err := inj.Invoke(func(d1 string, d2 specialString, d3 <-chan *specialString, d4 chan<- *specialString) {
			expect(t, dep, d1)
			expect(t, dep2, d2)
			expect(t, reflect.TypeOf(dep3).Elem(), reflect.TypeOf(d3).Elem())
			expect(t, reflect.TypeOf(dep4).Elem(), reflect.TypeOf(d4).Elem())
			expect(t, reflect.RecvDir, reflect.TypeOf(d3).ChanDir())
			expect(t, reflect.SendDir, reflect.TypeOf(d4).ChanDir())
		})
		expect(t, err, nil)

		_, err = inj.Invoke(myFastInvoker(func(string) {}))
		expect(t, err, nil)
	})

	t.Run("invoke functions with return values", func(t *testing.T) {
		inj := New()

		dep := "some dependency"
		inj.Map(dep)
		dep2 := "another dep"
		inj.MapTo(dep2, (*specialString)(nil))

		result, err := inj.Invoke(func(d1 string, d2 specialString) string {
			expect(t, dep, d1)
			expect(t, dep2, d2)
			return "Hello world"
		})
		expect(t, err, nil)

		expect(t, "Hello world", result[0].String())
	})
}

func TestInjector_Apply(t *testing.T) {
	inj := New()
	inj.Map("a dep").MapTo("another dep", (*specialString)(nil))

	s := testStruct{}
	expect(t, inj.Apply(&s), nil)

	expect(t, "a dep", s.Dep1)
	expect(t, "another dep", s.Dep2)
}

func TestInjector_Load(t *testing.T) {
	inj := New()

	dep1 := "a dep"
	inj.Map(&dep1)
	dep1l := ""
	expect(t, inj.Load(&dep1l), nil)

	g := &greeter{"Jeremy"}
	inj.Map(g)
	g2 := &greeter{}
	expect(t, inj.Load(g2), nil)
	expect(t, g.Name, g2.Name)
}

func TestInjector_InterfaceOf(t *testing.T) {
	iType := InterfaceOf((*specialString)(nil))
	expect(t, reflect.Interface, iType.Kind())

	iType = InterfaceOf((**specialString)(nil))
	expect(t, reflect.Interface, iType.Kind())

	defer func() {
		refute(t, recover(), nil)
	}()
	InterfaceOf((*testing.T)(nil))
}

func TestInjector_Map(t *testing.T) {
	inj := New()

	g := &greeter{"Jeremy"}
	inj.Map(g)

	expect(t, inj.Value(InterfaceOf((*fmt.Stringer)(nil))).IsValid(), true)
}

func BenchmarkInjector_Map(b *testing.B) {
	b.ReportAllocs()
	inj := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inj.Map("Jeremy")
	}
}

func TestInjector_Set(t *testing.T) {
	inj := New()

	typ := reflect.TypeOf("string")
	typSend := reflect.ChanOf(reflect.SendDir, typ)
	typRecv := reflect.ChanOf(reflect.RecvDir, typ)

	// Instantiating unidirectional channels is not possible using reflect, see body
	// of reflect.MakeChan for detail.
	chanRecv := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, typ), 0)
	chanSend := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, typ), 0)

	inj.Set(typSend, chanSend)
	inj.Set(typRecv, chanRecv)

	expect(t, inj.Value(typSend).IsValid(), true)
	expect(t, inj.Value(typRecv).IsValid(), true)
	expect(t, inj.Value(chanSend.Type()).IsValid(), false)
}

func TestInjector_GetVal(t *testing.T) {
	inj := New()
	inj.Map("some dependency")

	expect(t, inj.Value(reflect.TypeOf("string")).IsValid(), true)
	expect(t, inj.Value(reflect.TypeOf(11)).IsValid(), false)
}

func TestInjector_Reset(t *testing.T) {
	inj := New()
	inj.Map("some dependency")
	expect(t, inj.Value(reflect.TypeOf("string")).IsValid(), true)

	inj.Reset()
	expect(t, inj.Value(reflect.TypeOf("string")).IsValid(), false)

	injFather := New()
	injFather.Map("some dependency")
	inj.SetParent(injFather)
	expect(t, inj.Value(reflect.TypeOf("string")).IsValid(), true)

	inj.Reset()
	expect(t, inj.Value(reflect.TypeOf("string")).IsValid(), false)
}

func BenchmarkInjector_Reset(b *testing.B) {
	b.ReportAllocs()
	inj := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inj.Reset()
	}
}

func TestInjector_SetParent(t *testing.T) {
	inj := New()
	inj.MapTo("another dep", (*specialString)(nil))

	inj2 := New()
	inj2.SetParent(inj)

	expect(t, inj2.Value(InterfaceOf((*specialString)(nil))).IsValid(), true)
}

func TestIsFastInvoker(t *testing.T) {
	expect(t, IsFastInvoker(myFastInvoker(nil)), true)
}

func BenchmarkInjector_Invoke(b *testing.B) {
	inj := New()
	inj.Map("some dependency").MapTo("another dep", (*specialString)(nil))

	fn := func(d1 string, d2 specialString) string { return "something" }

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = inj.Invoke(fn)
	}
}

type testFastInvoker func(d1 string, d2 specialString) string

func (f testFastInvoker) Invoke(args []interface{}) ([]reflect.Value, error) {
	f(args[0].(string), args[1].(specialString))
	return nil, nil
}

func BenchmarkInjector_FastInvoke(b *testing.B) {
	inj := New()
	inj.Map("some dependency").MapTo("another dep", (*specialString)(nil))

	fn := testFastInvoker(func(d1 string, d2 specialString) string { return "something" })

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = inj.Invoke(fn)
	}
}

func TestConcurrent(t *testing.T) {
	t.Parallel()
	inj := New()
	t.Run("Map", func(t *testing.T) {
		var trigger, wg sync.WaitGroup
		trigger.Add(1)
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				trigger.Wait()
				inj.Map("")
				wg.Done()
			}()
		}
		trigger.Done()
		wg.Wait()
	})
	t.Run("MapTo", func(t *testing.T) {
		var trigger, wg sync.WaitGroup
		trigger.Add(1)
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				trigger.Wait()
				inj.MapTo("", (*Injector)(nil))
				wg.Done()
			}()
		}
		trigger.Done()
		wg.Wait()
	})
	t.Run("Set", func(t *testing.T) {
		var trigger, wg sync.WaitGroup
		trigger.Add(1)
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				trigger.Wait()
				inj.Set(reflect.TypeOf(""), reflect.ValueOf(""))
				wg.Done()
			}()
		}
		trigger.Done()
		wg.Wait()
	})
}
