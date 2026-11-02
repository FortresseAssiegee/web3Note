# 第四讲：Go 函数、指针与 defer 完整学习笔记

> 整理日期：2026-10-21。验证工具链：Go 1.27.3，linux/amd64。完整示例统一采用 `go 1.27.0` 模块语言设置；`new(值表达式)` 的版本变化在对应小节单独解释。
>
> 本讲从函数的定义、参数和返回值开始，逐步进入方法与指针，再用 `panic`、`recover` 和 `error` 说明异常路径，最后集中解释 `defer` 的登记、执行与资源清理。各部分围绕调用过程展开，避免把语法、内存和错误处理割裂成互不相关的知识点。

<a id="reading-notes"></a>

> **统一运行基线：** 建议使用 Go 1.27.3，示例模块声明采用 `go 1.27.0`。文中出现其他 Go 版本时，仅用于说明特性引入时间或历史实现差异；第三方依赖仍以对应模块的版本要求为准。


*图示：函数调用与资源释放。*

## 正文导航

- [一、主要内容](#chapter-1)
  - [1.1 本讲要解决的问题](#section-1-1)
  - [1.2 示例与来源的阅读方式](#section-1-2)
  - [1.3 阅读时保留的三个问题](#section-1-3)
- [二、Go 语言的函数机制（函数基础部分）](#chapter-2)
  - [2.1 函数的定义](#section-2-1)
  - [2.2 函数的多返回值](#section-2-2)
  - [2.3 具名函数](#section-2-3)
  - [2.4 匿名函数](#section-2-4)
  - [2.5 闭包函数](#section-2-5)
  - [2.6 函数的可变参数](#section-2-6)
  - [2.7 函数作为参数和返回值](#section-2-7)
  - [2.8 函数里面的 defer](#section-2-8)
  - [2.9 递归函数](#section-2-9)
  - [2.10 函数的错误处理机制](#section-2-10)
  - [2.11 函数方法](#section-2-11)
- [三、make 和 new（预声明内置函数）](#chapter-3)
  - [3.1 用途](#section-3-1)
  - [3.2 返回值](#section-3-2)
  - [3.3 语法示例](#section-3-3)
  - [3.4 总结](#section-3-4)
- [四、指针](#chapter-4)
  - [4.1 指针的基本概念](#section-4-1)
  - [4.2 指针的使用场景](#section-4-2)
  - [4.3 指针的零值问题](#section-4-3)
  - [4.4 指针常见问题](#section-4-4)
  - [4.5 总结](#section-4-5)
- [五、函数的高阶内容](#chapter-5)
  - [5.1 指针作为函数参数](#section-5-1)
  - [5.2 使用指针作为方法的接收者](#section-5-2)
  - [5.3 函数异常处理](#section-5-3)
    - [5.3.1 使用 `panic` 和 `recover` 进行异常处理](#section-5-3-1)
    - [5.3.2 `panic` 的场景](#section-5-3-2)
    - [5.3.3 函数错误处理的 `error` 接口](#section-5-3-3)
    - [5.3.4 `error` 和 `panic` 的区别](#section-5-3-4)
    - [5.3.5 结合 `defer`、`panic` 和 `recover` 处理资源泄漏](#section-5-3-5)
    - [5.3.6 总结](#section-5-3-6)
- [六、defer 关键字](#chapter-6)
  - [6.1 defer 执行顺序：先进后出](#section-6-1)
  - [6.2 在函数里面，defer 和 return 的执行](#section-6-2)
  - [6.3 具名返回值遇见 defer](#section-6-3)
  - [6.4 defer 遇见 panic 会发生什么](#section-6-4)
  - [6.5 defer 中包含 panic](#section-6-5)
  - [6.6 defer 下的函数参数包含子函数](#section-6-6)
- [附录 A：综合示例——读取文件并统计非负整数](#appendix-a)
  - [A.1 将函数、指针方法、error 和 defer 放到同一流程](#section-a-1)
  - [A.2 沿着一次失败调用阅读程序](#section-a-2)
- [附录 B：练习与参考答案](#appendix-b)
  - [B.1 练习题](#section-b-1)
  - [B.2 参考答案与推导](#section-b-2)
- [附录 C：官方文档与后续阅读](#appendix-c)
  - [C.1 官方资料按用途索引](#section-c-1)
  - [C.2 中文辅助资料](#section-c-2)
  - [C.3 版本和示例验证说明](#section-c-3)
  - [C.4 来源、改编与许可](#section-c-4)

<a id="chapter-1"></a>

## 一、主要内容

<a id="section-1-1"></a>

### 1.1 本讲要解决的问题

本讲围绕“函数怎样接收数据、形成结果，以及离开函数前做什么”展开。先掌握普通调用，再理解共享数据和失败路径，最后用 `defer` 的求值时机把它们串起来。

| 学习问题 | 对应章节 | 学会后的判断能力 |
| --- | --- | --- |
| 一段行为如何定义、传递、返回并保存状态？ | 第二章：函数机制 | 看懂函数签名、多返回值、闭包和高阶函数 |
| 应该得到一个可用容器，还是一个指向变量的指针？ | 第三章：make 与 new | 区分返回类型、零值与容器初始化 |
| 函数为什么有时能改动调用方的数据？ | 第四章：指针；第五章 5.1 | 找出复制的值和共享的对象，区分修改与重新绑定 |
| 一个类型的方法怎样与接口要求对应？ | 第五章 5.2 | 区分方法调用的自动取地址和接口的方法集 |
| 普通失败怎样返回，panic 怎样传播？ | 第五章 5.3 | 使用 error、错误链、恢复边界及资源清理 |
| return、defer、panic 的输出次序如何推导？ | 第六章：defer | 根据求值、赋值、执行和栈展开逐步推演 |

基础关系的官方入口是：[函数类型](https://go.dev/ref/spec#Function_types)、[调用](https://go.dev/ref/spec#Calls)、[指针类型](https://go.dev/ref/spec#Pointer_types)、[defer 语句](https://go.dev/ref/spec#Defer_statements)。

<a id="section-1-2"></a>

### 1.2 示例与来源的阅读方式

每段标记 `// 运行示例：编号` 且包含 `package main` 的代码都是一个独立程序。复制一个完整程序到 `main.go` 后运行；不同示例有各自的 `main`，不要把它们直接拼进同一个包。只演示签名、语法或编译错误的短代码会明确标注为片段。

示例的两类来源在所在小节说明：一类是根据官方规则原创的讲解程序，另一类是对官方示例的改编，补齐 `main`、改换输入或加入执行轨迹。它们都附有对应文档，不能将“符合官方规则”理解成“全部代码逐字摘录自官方”。

建议用单独的练习目录：

```sh
mkdir go-ch04-demo
cd go-ch04-demo
go mod init example.com/go-ch04-demo
go mod edit -go=1.22
# 将一个完整示例保存为 main.go
go run .
```

`new(值表达式)` 需要 Go 1.26 或更高工具链及相应语言版本；运行标注 Go 1.26+ 的例子前，把模块设置改成 `go mod edit -go=1.27.0`。版本说明见 [Go 1.26 的语言变化](https://go.dev/doc/go1.26#language)。关于 `go.mod` 中 `go` 指令的含义，见 [Go Modules Reference](https://go.dev/ref/mod#go-mod-file-go)。

<a id="section-1-3"></a>

### 1.3 阅读时保留的三个问题

遇到一个例子，先问“现在复制了哪个值”，再问“哪个对象仍然被共享”，最后问“这条语句现在执行，还是只登记了一个以后要调用的函数”。这三个问题分别帮助理解参数、指针与闭包，以及 `defer`。正文的执行表和输出推导会反复使用这组问题，但每次都落到具体变量和语句上。

先阅读第二至四章即可完成基本函数练习；需要解释接口方法集、错误包装和复杂退出流程时，再进入第五、六章。附录包含综合例子、练习答案和官方阅读索引，便于之后查阅。

<a id="chapter-2"></a>

## 二、Go 语言的函数机制（函数基础部分）

函数把一段有明确输入、输出和职责的处理过程封装起来。学习函数时，需要同时看清三件事：调用时传入了什么值，执行过程中修改了哪一个变量，以及结束时把什么结果交给调用者。本章先建立这些基础；指针接收者、错误包装、`panic/recover` 和复杂的 `defer` 顺序会在后文继续展开。

本章的完整程序均按官方规范编写为独立示例；标注“参考官方示例”的地方说明了具体来源。代码中的 `F-01` 等编号用于查找和核对，每一个程序都应放在独立目录运行，不能把全部 `main` 函数直接拼进一个文件。

<a id="section-2-1"></a>

### 2.1 函数的定义

<a id="heading-009"></a>

#### 语法与组成

下面是语法示意，不是独立程序：

```go
func 函数名(参数名 参数类型, 参数名 参数类型) 返回值类型 {
    // 函数体
    return 返回值
}
```

`func` 声明函数；参数列表说明调用者必须提供哪些值；返回值列表说明调用者可以得到哪些结果。相邻参数类型相同时可以合并，例如 `func add(a, b int) int` 等价于分别写出两个 `int`。无返回值时省略返回类型；单个未命名返回值可以不加括号；多个返回值需要括号。

参数与实参要分开理解：`add` 声明中的 `a`、`b` 是参数变量，`add(3, 5)` 中的 `3`、`5` 是本次调用提供的实参。函数调用会把实参的值赋给参数，然后执行函数体。调用者的变量和被调用函数的参数是不同变量，即使它们碰巧用了相同名称。

```go
// 运行示例：F-01
package main

import "fmt"

func add(a, b int) int {
	return a + b
}

func increase(n int) int {
	n++
	return n
}

func main() {
	x := 10
	y := increase(x)
	fmt.Println(add(3, 5))
	fmt.Println(x, y)
}
```

输出：

```text
8
10 11
```

`increase(x)` 把整数值 `10` 交给参数 `n`。`n++` 修改的是这一调用中的 `n`；`return n` 把 `11` 返回给 `y`。没有任何语句给 `x` 重新赋值，所以 `x` 仍为 `10`。

**所有参数都按值传递。** 后面传入指针、切片、map 或函数值时，这条规则仍然成立。只是某些值包含通向共享数据的引用，复制这些值以后仍可能操作同一份数据；“能影响外部数据”不能直接推导成“按引用传递”。

定义函数时还应注意：

- 普通具名函数声明位于包级，不能直接嵌套写在另一个函数体中；局部处理逻辑可使用函数字面量。
- Go 不根据参数类型或数量自动选择同名普通函数。在同一包级作用域中，不能分别声明 `add(int, int)` 和 `add(float64, float64)` 来实现重载。
- 参数必须满足类型赋值规则。已有的 `int64` 变量不会自动转换成 `int` 参数，必要时应显式转换并考虑范围。
- 有返回结果的函数，必须保证控制流以规范认可的终止语句结束。最常见的写法是让每一条可能执行到末尾的路径明确 `return`。
- 函数名首字母大写涉及跨包导出；它与函数能否返回值、能否递归没有关系。

Go 的普通函数参数没有在声明处填写默认值的语法，也不会因为调用时少写一个参数就自动补零。确实需要可选配置时，可以提供含默认行为的包装函数、传入配置结构体，或者设计明确的可变参数接口。不要为了模拟重载而把所有参数都写成 `any`，否则原本由编译器检查的类型约束会转移到运行时。

阅读一个函数时，可以先忽略函数体，只看签名并回答：“我需要准备哪些数据？调用后需要检查哪几个结果？”再进入实现，寻找边界检查、正常路径和返回位置。这个顺序有助于把接口契约与内部算法分开，也便于为复杂函数选择更准确的名字。

官方依据：[函数声明](https://go.dev/ref/spec#Function_declarations)、[函数调用](https://go.dev/ref/spec#Calls)、[FAQ：参数传递](https://go.dev/doc/faq#pass_by_value)、[FAQ：为什么不支持重载](https://go.dev/doc/faq#overloading)。

<a id="section-2-2"></a>

### 2.2 函数的多返回值

一个函数可以返回多个彼此相关的结果。常见设计包括“结果与错误”“值与是否找到”“处理的数据量与处理错误”。这些信息通常来自同一次计算，一起返回可以避免重复执行，也让调用者获得完整状态。

下面的函数返回最小值、最大值和是否存在数据。遇到空切片时，仅返回 `0, 0` 无法区分“没有数据”和“数据都为零”，因此额外提供 `bool`。

```go
// 运行示例：F-02
package main

import "fmt"

func bounds(values []int) (int, int, bool) {
	if len(values) == 0 {
		return 0, 0, false
	}
	low, high := values[0], values[0]
	for _, v := range values[1:] {
		if v < low {
			low = v
		}
		if v > high {
			high = v
		}
	}
	return low, high, true
}

func main() {
	low, high, ok := bounds([]int{8, 3, 12, 5})
	fmt.Println(low, high, ok)

	low, high, ok = bounds(nil)
	fmt.Println(low, high, ok)

	_, maximum, found := bounds([]int{6, 9})
	fmt.Println(maximum, found)
}
```

输出：

```text
3 12 true
0 0 false
9 true
```

第一次调用从第一个元素建立初值，再逐个更新边界；第二次调用在访问 `values[0]` 之前检查了长度，因此 `nil` 切片不会触发越界。第三次调用使用空白标识符 `_` 丢弃最小值，但仍保留 `found` 以判断其他结果是否有效。

接收返回值时，左侧变量数量必须与右侧结果匹配。多返回值不是一个可随意当成单个变量存储的“元组对象”，不能写 `result := bounds(values)` 再通过 `result[0]` 访问结果。如果确实希望把一组数据作为单一对象传递，可设计一个结构体。

某些位置允许直接转交多返回值，例如返回类型匹配时的 `return anotherFunction()`，或符合规范约束时将一个多返回值调用作为另一个函数的唯一实参表达式。初学时先拆成命名变量，通常更容易看清每一项的含义。对 `(value, error)` 这样的返回组合，也不要机械地认为“有错误时其他值一定没有用”：是否存在部分结果，应由具体 API 契约决定，例如读写接口可能同时报告已处理数量与错误。

官方依据：[返回语句](https://go.dev/ref/spec#Return_statements)、[Effective Go：多返回值](https://go.dev/doc/effective_go#multiple-returns)、[io.Reader 的返回约定](https://pkg.go.dev/io#Reader)。

<a id="section-2-3"></a>

### 2.3 具名函数

本章的“具名函数”指通过函数声明给函数绑定名字，例如 `func subtract(...)`。这与“具名返回值”是两件独立的事：前者给整个函数命名，后者给函数的结果参数命名。

| 写法 | 函数是否具名 | 返回值是否具名 |
|---|---|---|
| `func subtract(a, b int) int` | 是，名字是 `subtract` | 否 |
| `func split(n int) (q, r int)` | 是，名字是 `split` | 是，名字是 `q`、`r` |
| `func(n int) (q, r int) { ... }` | 函数字面量没有声明名 | 是 |

```go
// 运行示例：F-03
package main

import "fmt"

func subtract(a, b int) int {
	return a - b
}

func split(n int) (q, r int) {
	q = n / 2
	r = n % 2
	return q, r
}

func main() {
	f := subtract
	fmt.Println(f(7, 2))
	q, r := split(9)
	fmt.Println(q, r)
}
```

输出：

```text
5
4 1
```

`f := subtract` 保存函数值，表达式中没有 `()`，此时不会执行减法。`f(7, 2)` 才是真正的调用。`split` 中的结果参数 `q`、`r` 进入函数时为零值，可以像普通局部变量一样赋值；即使结果已经具名，仍然可以使用显式的 `return q, r`。空的 `return` 也能返回当前结果参数，但在较长函数中，显式写出结果通常更方便阅读。

函数类型由参数和返回结果的类型、数量，以及是否可变参数等因素决定，参数变量的名字不需要一致。例如 `func(a int) int` 和 `func(x int) int` 是相同的函数类型。名称用来帮助读者理解意义，不会形成新的类型差异。

具名结果在使用短变量声明时需要特别留意遮蔽。例如，函数已经声明结果参数 `err`，内层 `if` 初始化语句又用 `:=` 声明一个 `err`，那么内层和外层可能是不同变量。在这种作用域内写空的 `return` 还可能被编译器拒绝。初学时优先采用 `if err != nil { return ..., err }` 这样的明确返回方式，并核对每个 `err` 到底属于哪个作用域。

<a id="heading-012"></a>

#### 函数值的默认状态与比较

函数类型变量的零值是 `nil`。可以比较 `f == nil`，但不能比较两个函数值是否相等，也不能把函数值作为 map 的键。调用 `nil` 函数值会产生运行时 panic，因此可选回调必须先检查或者提供明确默认实现。

```go
// 运行示例：F-04
package main

import "fmt"

type Formatter func(string) string

func main() {
	var format Formatter
	fmt.Println(format == nil)

	if format == nil {
		format = func(s string) string {
			return "[" + s + "]"
		}
	}
	fmt.Println(format("Go"))
}
```

输出：

```text
true
[Go]
```

这里先声明函数类型 `Formatter`，再给变量提供一个符合签名的函数。`format` 是变量名，不是 `func format(...)` 形式的函数声明名。把匿名函数赋给一个有名字的变量，不会改变它最初属于函数字面量这一事实。

官方依据：[函数类型](https://go.dev/ref/spec#Function_types)、[类型相同的判定](https://go.dev/ref/spec#Type_identity)、[比较运算符](https://go.dev/ref/spec#Comparison_operators)、[返回语句](https://go.dev/ref/spec#Return_statements)。

<a id="section-2-4"></a>

### 2.4 匿名函数

匿名函数通过函数字面量创建，基本形式是 `func(参数) 返回类型 { 函数体 }`。它可以保存到变量中、作为实参传入其他函数，也可以在定义后立即调用。适合处理局部、短小且与当前位置紧密相关的逻辑。

```go
// 运行示例：F-05
package main

import "fmt"

func main() {
	double := func(n int) int {
		return n * 2
	}
	fmt.Println(double(6))

	square := func(n int) int {
		return n * n
	}(4)
	fmt.Println(square)
}
```

输出：

```text
12
16
```

`double` 的类型是 `func(int) int`，保存的是将来可调用的函数。`square` 后面的 `(4)` 则立即执行前面的函数字面量，保存的是执行结果 `16`，因此 `square` 的类型是 `int`。看见末尾是否有 `()`，是区分“函数值”和“函数调用结果”的直接方法。

匿名函数不是另一套参数传递机制，它与具名函数一样遵守类型、作用域和值传递规则。其特殊之处在于可以直接写在表达式位置，并访问外层函数的局部变量。函数体内部的 `return` 结束的是这个匿名函数本身，不会跨过函数边界直接结束外层调用者。

匿名函数并非越多越好：逻辑变长、需要在多处复用或需要独立命名解释职责时，提取为具名函数通常更清晰。也不要为了制造局部作用域而引入层层立即调用函数，普通代码块在不涉及返回与延迟清理时往往已经足够。

如果局部匿名函数需要递归地引用自己的变量，通常要先用 `var` 声明函数变量，再给它赋函数字面量。直接写 `f := func() { f() }` 时，新声明的 `f` 在其初始化表达式中还没有进入作用域。这与包级具名函数可以在函数体中直接使用自己的名字不同；区别来自变量声明作用域，不是匿名函数无法递归。

官方依据：[函数字面量](https://go.dev/ref/spec#Function_literals)、[函数调用](https://go.dev/ref/spec#Calls)、[声明与作用域](https://go.dev/ref/spec#Declarations_and_scope)。

<a id="section-2-5"></a>

### 2.5 闭包函数

闭包可以理解为“函数及其所使用的外部变量环境”。Go 的函数字面量可以引用外层函数声明的变量。只要这些变量仍可由闭包访问，它们就不会因为外层函数已经返回而失效。

<a id="heading-015"></a>

#### 保留状态：一个计数器实例有一份状态

```go
// 运行示例：F-06
package main

import "fmt"

func newCounter(start int) func() int {
	count := start
	return func() int {
		count++
		return count
	}
}

func main() {
	a := newCounter(0)
	b := newCounter(10)
	fmt.Println(a())
	fmt.Println(a())
	fmt.Println(b())
	fmt.Println(a())

	limit := 2
	show := func() { fmt.Println(limit) }
	limit = 5
	show()
}
```

输出：

```text
1
2
11
3
5
```

第一次 `newCounter` 调用创建供 `a` 使用的 `count`，第二次调用创建供 `b` 使用的另一份 `count`。因此，反复调用 `a` 会持续更新同一份状态，而调用 `b` 不会改变 `a` 的计数器。即使 `newCounter` 已经返回，被闭包引用的状态仍然有效。

末尾的 `show` 演示另一条关键规则：闭包引用的是变量，不会自动把创建时的 `2` 冻结下来。调用前 `limit` 改为 `5`，所以输出是 `5`。如果希望保留某个时刻的值，可以先复制到一个独立变量，或者通过一次函数调用把该值放进新的参数变量中。

不要把闭包状态理解成“每次调用自动重建”。上例的 `count := start` 属于工厂函数 `newCounter` 的函数体，不属于返回的匿名函数体。每调用一次工厂，创建一个计数器；每调用一次计数器，只执行自增和返回。若多个 goroutine 共同调用同一个计数器，需要另行同步，它不会因为封装在闭包内而自动具备并发安全性。

<a id="heading-016"></a>

#### 循环捕获：区分 `:=` 声明与 `=` 赋值

从 Go 1.22 的语言语义开始，由循环声明的迭代变量按迭代分别创建。用 `=` 给循环外已有变量赋值时，仍然复用原变量。以下示例避免引入并发，直接保存闭包后依次调用，便于观察差别。

```go
// 运行示例：F-07
package main

import "fmt"

func evaluate(functions []func() int) []int {
	results := make([]int, 0, len(functions))
	for _, f := range functions {
		results = append(results, f())
	}
	return results
}

func main() {
	values := []int{10, 20, 30}

	var separate []func() int
	for _, value := range values {
		separate = append(separate, func() int { return value })
	}
	fmt.Println(evaluate(separate))

	var shared []func() int
	var value int
	for _, value = range values {
		shared = append(shared, func() int { return value })
	}
	fmt.Println(evaluate(shared))
}
```

输出（Go 1.22 及以上语言语义）：

```text
[10 20 30]
[30 30 30]
```

第一组中的每个闭包引用各自迭代的 `value`，第二组中的三个闭包共同引用循环外的那个 `value`。循环结束后，这个共享变量的值是最后一次赋入的 `30`。

判断旧代码时不能只看本机安装的编译器版本，也应查看模块 `go.mod` 的 `go` 指令等语言版本控制。本笔记统一使用 `go 1.27.0` 模块语言设置核验基础示例。旧教程常用“所有循环闭包都会看到同一个最终值”概括旧语义，直接套用到新语义下的 `:=` 循环会得出错误结论。

官方依据：[函数字面量与共享变量](https://go.dev/ref/spec#Function_literals)、[Go 1.22 语言变化](https://go.dev/doc/go1.22#language)、[官方博客：修复循环变量问题](https://go.dev/blog/loopvar-preview)。

<a id="section-2-6"></a>

### 2.6 函数的可变参数

可变参数让函数接收零个或多个同类型实参，声明形式为 `...T`，只能出现在参数列表最后一项。进入函数后，这个参数按 `[]T` 使用；没有提供该部分实参时，其值为 `nil` 切片。

```go
// 运行示例：F-08
package main

import "fmt"

func sum(values ...int) int {
	total := 0
	for _, v := range values {
		total += v
	}
	return total
}

func main() {
	fmt.Println(sum())
	fmt.Println(sum(2, 4, 6))
	numbers := []int{3, 5, 7}
	fmt.Println(sum(numbers...))
}
```

输出：

```text
0
12
15
```

`sum()` 中循环不执行，返回初值 `0`。`sum(2, 4, 6)` 把三个实参提供给可变参数。已有切片则使用 `numbers...` 展开，使该切片作为可变参数的切片值传入；直接写 `sum(numbers)` 会把一个 `[]int` 当成一个 `int` 实参，类型不匹配。

<a id="heading-018"></a>

#### 展开切片不会自动复制元素

```go
// 运行示例：F-09
package main

import "fmt"

func setFirst(values ...int) {
	if len(values) > 0 {
		values[0] = 99
	}
}

func main() {
	data := []int{1, 2, 3}
	setFirst(data...)
	fmt.Println(data)

	names := []string{"Go", "Gopher"}
	items := make([]any, len(names))
	for i, name := range names {
		items[i] = name
	}
	fmt.Println(items...)
}
```

输出：

```text
[99 2 3]
Go Gopher
```

`data...` 传入的切片与调用者的切片共享底层数组。参数变量本身仍是按值传递，但 `values[0] = 99` 写入的是共同的数组元素。若函数不应影响输入，可以在函数中显式复制所需元素再处理；若需要把 `append` 后可能变化的切片长度和底层数组选择传给调用者，应返回新的切片值。

`fmt.Println` 的参数是 `...any`。单个 `string` 可以赋给 `any`，但这不意味着 `[]string` 可以整体赋给 `[]any`，所以 `fmt.Println(names...)` 不合法。示例逐个把字符串装入 `[]any` 后再展开；若只是查看整个切片，直接 `fmt.Println(names)` 即可，它会把切片作为一个实参输出。

还应避免把 `...T` 与 `[]T` 当成相同的函数类型。它们在函数体内都能用切片方式处理参数，但调用语法不同，`func(...int)` 和 `func([]int)` 不是相同类型。对同一可变参数部分，也不能直接把若干独立实参与一个展开切片混合成 `sum(1, numbers...)`；先构造所需切片再展开更明确。存在固定前置参数时，例如 `logf(prefix string, args ...any)`，固定参数之后使用 `args...` 则是正常写法。

官方依据：[可变参数的传递规则](https://go.dev/ref/spec#Passing_arguments_to_..._parameters)、[函数类型](https://go.dev/ref/spec#Function_types)、[fmt.Println](https://pkg.go.dev/fmt#Println)。

<a id="section-2-7"></a>

### 2.7 函数作为参数和返回值

函数值可以像其他值一样保存、传递和返回。这样可以把“如何组织处理流程”和“对每一个元素执行什么操作”分别表达出来。接受函数参数或返回函数的函数，通常被称为高阶函数。

```go
// 运行示例：F-10
package main

import "fmt"

func transform(values []int, fn func(int) int) []int {
	result := make([]int, len(values))
	for i, v := range values {
		result[i] = fn(v)
	}
	return result
}

func makeScaler(factor int) func(int) int {
	return func(value int) int {
		return factor * value
	}
}

func main() {
	data := []int{1, 2, 3}
	timesThree := makeScaler(3)
	fmt.Println(transform(data, timesThree))
	fmt.Println(transform(data, func(v int) int { return v + 10 }))
	fmt.Println(data)
}
```

输出：

```text
[3 6 9]
[11 12 13]
[1 2 3]
```

`transform` 决定遍历顺序和结果存储位置，`fn` 决定元素如何转换。`makeScaler(3)` 生成捕获 `factor` 的函数，所以回调每收到一个值，就将其乘以 `3`。第二次调用直接传入匿名函数，复用相同遍历框架完成加法转换。原切片不变，是因为 `transform` 新建结果切片并且未向输入元素写入；不能单凭“使用了回调”推断没有副作用。

设计此类接口时，应明确回调是否允许为 `nil`、调用次数、是否按顺序调用以及是否会并发执行。上例约定回调必须非 `nil`，并按切片顺序逐次调用；本程序传入的两个回调都满足这一前提。对于公共库，可以在入口检查 `nil` 并按接口约定返回错误或采用默认行为。一个回调是否修改外部状态，也取决于它的具体实现。

另一个实用判断是“是否真的存在多种处理策略”。如果一段遍历永远只执行同一个固定操作，把它拆成回调接口不一定能提高可读性；当同一流程需要替换不同计算、筛选或比较规则时，函数参数才把变化点表达得更直接。返回闭包则适合先固定一部分配置，之后多次处理不同输入，例如先确定倍率，再逐个转换数字。

标准库大量使用这类设计。例如 `sort.Slice` 通过 `less(i, j int) bool` 回调获取排序关系；`strings.FieldsFunc` 通过 `func(rune) bool` 判断分隔字符。回调签名限定输入输出，具体语义仍要阅读对应 API 文档，不能只凭类型猜测调用顺序。

官方依据：[函数类型](https://go.dev/ref/spec#Function_types)、[sort.Slice](https://pkg.go.dev/sort#Slice)、[strings.FieldsFunc](https://pkg.go.dev/strings#FieldsFunc)。

<a id="section-2-8"></a>

### 2.8 函数里面的 defer

`defer` 把一个调用推迟到当前函数即将返回时执行，常用于释放资源或补充结束处理。本节先掌握三点：注册时确定调用所需的函数值和实参；退出时执行延迟调用；多个延迟调用按注册顺序的逆序执行。

```go
// 运行示例：F-11
package main

import "fmt"

func task() {
	n := 1
	defer fmt.Println("defer argument:", n)
	defer func() {
		fmt.Println("defer closure:", n)
	}()
	n = 2
	fmt.Println("body:", n)
}

func main() {
	task()
	fmt.Println("caller")
}
```

输出：

```text
body: 2
defer closure: 2
defer argument: 1
caller
```

第一条 `defer` 的实参 `n` 在注册时求值，因此保存了 `1`。第二条注册的是匿名函数调用，注册时不会执行函数体中的 `fmt.Println`；该函数体后来通过闭包读取 `n`，看到的是已经修改后的 `2`。最后注册的闭包先执行，然后才执行第一条延迟调用。两者执行完毕后，`task` 才真正把控制权交回 `main`。

`defer` 以函数调用为边界，而不是以 `if`、`for` 或普通大括号代码块为边界。把资源关闭放在长循环里的 `defer`，资源会一直累积到整个函数结束；需要逐项及时释放时，可将单次处理提取成小函数。普通错误返回同样会执行已注册的 `defer`，因此它特别适合集中清理工作。

这里的三个规则参考官方博客中的基础演示重新编写。涉及具名返回值、嵌套调用、`panic` 和清理错误时，还需结合后面的完整 `defer` 章节理解，不应把它简单记成“函数最后打印一下”。

官方依据：[defer 语句](https://go.dev/ref/spec#Defer_statements)、[官方博客：Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)。

<a id="section-2-9"></a>

### 2.9 递归函数

递归是函数在处理问题时再次调用自身。适用于问题本身具有递归结构的情况，例如树遍历、目录遍历和分治算法。写递归时必须说明终止条件，并保证每一次递归调用都朝终止条件推进。

下面用阶乘演示调用与返回。为了避免无限递归和整数溢出，公开入口把输入限定为 `0` 到 `20`，返回值使用 `uint64`；`20!` 可被其表示，而 `21!` 已超出范围。

```go
// 运行示例：F-12
package main

import (
	"errors"
	"fmt"
)

func factorial(n int) (uint64, error) {
	if n < 0 || n > 20 {
		return 0, errors.New("n must be in [0, 20]")
	}
	return factorialUnchecked(n), nil
}

func factorialUnchecked(n int) uint64 {
	if n <= 1 {
		return 1
	}
	return uint64(n) * factorialUnchecked(n-1)
}

func main() {
	value, err := factorial(5)
	fmt.Println(value, err)
	value, err = factorial(0)
	fmt.Println(value, err)
	value, err = factorial(21)
	fmt.Println(value, err)
}
```

输出：

```text
120 <nil>
1 <nil>
0 n must be in [0, 20]
```

计算 `5!` 时，当前调用需要先取得 `4!` 的结果；后者继续等待 `3!`。到 `1!` 的终止分支返回 `1` 后，等待中的调用逐层完成乘法，最终得到 `120`。每次调用都有自己本次调用的参数 `n`，并不是一个共享的 `n` 被反复覆盖。

这个算法需要线性数量的调用；用普通循环也能计算阶乘，并可避免递归层数增长。Go 语言规范没有承诺通用的尾调用消除，因此不能把“最后一行调用自己”当作允许无限递归的理由。对深度不受控制的输入，应考虑显式栈或循环；不要依赖发生栈溢出后再通过 `recover` 兜底。

还要区分递归本身和重复计算：朴素递归 Fibonacci 经常对同一子问题计算多次，性能问题来自其计算结构。增加缓存或改用迭代，才是在这种场景下减少工作量的办法。对于阶乘这种数值算法，还应同时分析结果范围；终止条件正确不意味着数值结果一定不会溢出。

官方依据：[函数调用](https://go.dev/ref/spec#Calls)、[整数类型](https://go.dev/ref/spec#Numeric_types)、[整数溢出规则](https://go.dev/ref/spec#Integer_overflow)。递归过程与范围限制为本笔记基于这些规则编写的讲解分析。

<a id="section-2-10"></a>

### 2.10 函数的错误处理机制

Go 通常把可预期的失败作为普通结果返回。常见签名为 `func ... (...) (T, error)`：调用者先检查 `err`，再按函数约定使用结果。`error` 是接口类型；返回 `nil` 表示没有错误，返回非 `nil` 的错误值用于描述失败。

下面的解析函数把字符串转换成正整数。它参考标准库 `strconv.Atoi` 的官方调用方式，并增加业务范围检查和错误上下文。

```go
// 运行示例：F-13
package main

import (
	"errors"
	"fmt"
	"strconv"
)

func parseCount(text string) (int, error) {
	n, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("parse count %q: %w", text, err)
	}
	if n <= 0 {
		return 0, errors.New("count must be positive")
	}
	return n, nil
}

func main() {
	for _, text := range []string{"8", "0", "eight"} {
		count, err := parseCount(text)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println("count:", count)
	}
}
```

输出：

```text
count: 8
error: count must be positive
error: parse count "eight": strconv.Atoi: parsing "eight": invalid syntax
```

处理 `"8"` 时，转换成功且范围检查通过，函数返回 `8, nil`。处理 `"0"` 时，转换本身成功，但业务规则要求正数，所以返回自建错误。处理 `"eight"` 时，`Atoi` 已经失败，函数补充“正在解析数量”的上下文后立即返回，不继续执行后面的范围判断。

`errors.New` 创建错误消息；`fmt.Errorf` 可把具体输入等上下文放入消息。这里的 `%w` 包装原始错误，允许调用方后续通过 `errors.Is`、`errors.As` 识别其原因或类型；如果只是把 `err.Error()` 拼到字符串里，原始错误的结构关系便不会因此保留。

建议形成固定检查习惯：调用可能失败的函数后，立即考虑失败分支，决定当前层是能够处理、需要重试，还是应补充上下文后向上返回。不要为了让代码简短而随意用 `_` 丢弃错误，也不要在底层函数里既记录同一个错误又层层返回、导致上层重复记录。究竟由哪一层向调用方展示或写日志，应按应用职责确定。

返回错误不会自动停止调用者，也不会触发异常跳转；示例中失败后仍能处理下一项，是因为 `main` 明确使用了 `continue`。`panic` 处理的是另一种控制流程，不能把“返回了非 `nil` error”和“发生 panic”混为一谈。自定义错误类型、带类型的 `nil`、错误链及恢复边界见后面的高阶章节。

官方依据：[官方教程：返回和处理错误](https://go.dev/doc/tutorial/handle-errors)、[strconv.Atoi 及示例](https://pkg.go.dev/strconv#Atoi)、[errors.New](https://pkg.go.dev/errors#New)、[fmt.Errorf](https://pkg.go.dev/fmt#Errorf)、[官方博客：Error handling and Go](https://go.dev/blog/error-handling-and-go)。

<a id="section-2-11"></a>

### 2.11 函数方法

方法是带有接收者的函数。接收者写在 `func` 和方法名之间，说明该操作与哪种类型关联。调用时使用 `对象.方法(...)`，适合把行为组织在相关数据类型旁边。

以下语法片段展示与普通函数的区别：

```go
// 普通函数：接收者是普通参数。
func area(r Rectangle) int { return r.Width * r.Height }

// 方法：接收者写在方法名前。
func (r Rectangle) Area() int { return r.Width * r.Height }
```

这两段都要求另外定义 `Rectangle`，不能作为独立程序运行。下面提供完整的值接收者和指针接收者示例。

```go
// 运行示例：F-14
package main

import "fmt"

type Rectangle struct {
	Width  int
	Height int
}

func (r Rectangle) Area() int {
	return r.Width * r.Height
}

func (r Rectangle) GrowCopy(delta int) {
	r.Width += delta
	r.Height += delta
}

func (r *Rectangle) Grow(delta int) {
	r.Width += delta
	r.Height += delta
}

func main() {
	r := Rectangle{Width: 3, Height: 4}
	fmt.Println(r.Area())
	r.GrowCopy(1)
	fmt.Println(r.Width, r.Height)
	r.Grow(1)
	fmt.Println(r.Width, r.Height, r.Area())
}
```

输出：

```text
12
3 4
4 5 20
```

`Area` 和 `GrowCopy` 接收到 `Rectangle` 值的副本。前者只读取字段，后者修改副本，因此 `GrowCopy` 返回后外面的 `r` 仍为 `3×4`。`Grow` 接收到指向原变量的指针副本，访问字段时修改的是原变量，所以宽、高变为 `4`、`5`。

本例的 `r` 是可寻址变量，调用 `r.Grow(1)` 可以使用语言规定的便利形式，对应 `(&r).Grow(1)`。这种调用上的自动取址，并不意味着 `Rectangle` 与 `*Rectangle` 的方法集完全相同；当它们需要实现接口时，应分别判断，后文会给出完整例子。

在本章的普通非泛型形式中，接收者类型可以是本包定义的类型 `T` 或 `*T`；接收者基础类型不能是指针类型或接口类型。方法不只属于结构体，例如本包的 `type Celsius float64` 也可以有方法。不能直接给内置 `int` 或另一个包的类型增添方法；简单的类型别名也不会把外部类型变成自己的新定义类型。若需要扩展行为，可以定义新类型或者使用组合。

<a id="heading-024"></a>

#### 方法也能形成函数值

```go
// 运行示例：F-15
package main

import "fmt"

type Counter struct {
	N int
}

func (c Counter) Value() int {
	return c.N
}

func (c *Counter) Add(delta int) {
	c.N += delta
}

func main() {
	c := Counter{N: 2}
	read := c.Value
	add := c.Add
	c.N = 5
	add(3)
	fmt.Println(read(), c.Value())

	readExplicit := Counter.Value
	fmt.Println(readExplicit(c))
}
```

输出：

```text
2 8
8
```

`c.Value` 是方法值：创建时保存值接收者的副本，所以 `read()` 后来仍读取当时的 `2`。`c.Add` 也是方法值，但保存的是指向 `c` 的指针，因此 `add(3)` 把当前 `c.N` 从 `5` 改成 `8`。

`Counter.Value` 是方法表达式：没有提前绑定具体接收者，得到的函数需要显式把接收者作为第一个参数传入。因此，函数值 `readExplicit` 的类型是 `func(Counter) int`，而调用表达式 `readExplicit(c)` 的结果是一个 `int`。这两种形式适合不同的回调需求，区分“接收者何时确定”也有助于理解后面的延迟方法调用。

值接收者不等于深复制。如果结构体中包含切片、map 或指针字段，复制结构体只复制这些字段的值，仍可能通过字段触及共享数据。因此，是否需要指针接收者既要看是否修改接收者本身，也要看类型的一致用法和接口要求；不能仅凭“方法名称看起来像读取”或“值接收者总是安全”做决定。

官方依据：[方法声明](https://go.dev/ref/spec#Method_declarations)、[方法值](https://go.dev/ref/spec#Method_values)、[方法表达式](https://go.dev/ref/spec#Method_expressions)、[FAQ：值接收者与指针接收者](https://go.dev/doc/faq#methods_on_values_or_pointers)。

<a id="heading-025"></a>

#### 本章自查

完成本章后，应能解释下面这些现象，而不只是记住输出：为什么修改 `int` 参数不会改变外部变量，为什么切片展开传参却可以改到元素；为什么两个计数器的状态独立，而同一个闭包的状态会保留；为什么新语言语义下 `:=` 与 `=` 的循环捕获结果不同；为什么 `f` 和 `f()` 分别表示函数值与执行结果；以及为什么 `c.Value` 可以在之后调用时仍看到创建方法值时保存的接收者。

这些问题都回到同一组操作：声明创建变量，赋值复制值，调用建立本次调用的参数，闭包保留可访问的变量，而指针与其他包含引用的值允许多条访问路径到达同一份数据。

<a id="chapter-3"></a>

## 三、make 和 new（预声明内置函数）

本章的标题是“Make 和 New 关键字”。严格按 Go 的术语，`make` 和 `new` 都是**预声明的内置函数**，不是关键字；写代码时使用小写。它们不需要导入包，但语法由语言专门规定，不能把它们当作普通函数值传给另一个函数。学习时应先问“我要什么类型的值、对象需要什么初始状态”，再决定是否使用它们。[官方规范：内置函数](https://go.dev/ref/spec#Built-in_functions)

本章把传统的 `new(T)` 与 Go 1.26 新增的 `new(value)` 分开讲。除了明确标注的 M-06，其他示例不依赖 Go 1.26 新语法。所谓“零值”是一个类型在未显式指定初始值时得到的值，不能把它理解为“所有类型的值都是整数 0”。

<a id="section-3-1"></a>

### 3.1 用途

<a id="section-3-1-1"></a>

#### 3.1.1 new：创建变量，并取得指向它的指针

`new(T)` 创建一个类型为 `T` 的变量，将它初始化为 `T` 的零值，返回类型为 `*T` 的指针。整数的零值是 `0`，布尔值是 `false`，字符串是空字符串，指针、切片、map、channel、函数和接口的零值是 `nil`；结构体和数组则递归地把各字段或元素置为相应零值。[官方规范：分配](https://go.dev/ref/spec#Allocation)、[初始值](https://go.dev/ref/spec#The_zero_value)

例如，`new(int)` 的结果不是整数 `0`，而是“指向一个值为 `0` 的整数变量的指针”。需要整数值时写 `*p`，需要把可修改的对象传给另一个函数时，可以传 `p`。`new` 不是类的构造方法，也不会自动执行调用方定义的 `Init` 或 `NewXXX` 函数。

**示例 M-01：零值对象与指定初始值的对象。** 本例按官方分配规则自行编写，也对照了 Effective Go 中 `new(T)` 与结构体复合字面量的说明。

```go
// 运行示例：M-01
package main

import "fmt"

type Options struct {
	Retries int
	Debug   bool
}

func main() {
	var value Options
	p := new(Options)
	q := &Options{Retries: 3}

	fmt.Printf("value=%+v\n", value)
	fmt.Println(p == nil, p.Retries, p.Debug)
	fmt.Println(q == nil, q.Retries, q.Debug)
}
```

输出：

```text
value={Retries:0 Debug:false}
false 0 false
false 3 false
```

三种写法解决的是不同表达需求：`var value Options` 得到一个结构体值，`new(Options)` 得到零值结构体的指针，`&Options{Retries: 3}` 得到带初始配置的结构体指针。`p.Retries` 是通过结构体指针访问字段的简写，在这里与 `(*p).Retries` 等价。`p` 非 nil，并不表示它的全部字段都已经符合业务需要；例如业务要求重试次数大于零，就仍然需要设置字段或在构造函数中检查。

来源：[Effective Go：new](https://go.dev/doc/effective_go#allocation_new)、[构造函数与复合字面量](https://go.dev/doc/effective_go#composite_literals)。

<a id="section-3-1-2"></a>

#### 3.1.2 make：创建可使用的切片、map 或 channel

`make` 用于切片、map 和 channel。它返回所请求的类型本身，例如 `make([]int, 3)` 返回 `[]int`，而不是 `*[]int`。这里的“可使用”需要结合类型理解：切片具有指定长度和容量，map 可以添加键值对，channel 可按其缓冲容量进行通信；无缓冲 channel 的发送仍然需要接收方配合。[官方规范：创建切片、map 和 channel](https://go.dev/ref/spec#Making_slices_maps_and_channels)

`make` 不是“为任何类型分配内存”的通用函数。`make(int)`、`make(*int)` 和 `make(MyStruct)` 都不合法。反过来，也不是只要遇到切片或 map 就必须写 `make`：切片与 map 的复合字面量同样可以直接创建对象，nil 切片也可以直接用于 `append`。应根据“已知初始内容”“已知长度”“预计容量”等实际需要选择写法。

<a id="section-3-2"></a>

### 3.2 返回值

理解这一节的关键是把**指针变量本身**与**指针指向的值**分开。

| 表达式 | 表达式的类型 | 得到的初始状态 | 典型用途 |
|---|---|---|---|
| `new(int)` | `*int` | 指针非 nil，`*p == 0` | 需要一个整数变量的指针 |
| `new(User)` | `*User` | 指向零值 `User` | 取得零值对象指针 |
| `new([]int)` | `*[]int` | 指针非 nil，但 `*p` 是 nil 切片 | 少见；需要替换切片变量本身时才考虑 |
| `new(map[string]int)` | `*map[string]int` | 指针非 nil，但 `*p` 是 nil map | 少见；尚不能向该 map 写入 |
| `make([]int, 2, 5)` | `[]int` | 非 nil，长度 2，容量 5，元素为零值 | 指定切片长度及预留容量 |
| `make(map[string]int)` | `map[string]int` | 非 nil，长度 0，可写入 | 创建空 map |
| `make(chan int, 2)` | `chan int` | 非 nil，缓冲容量 2 | 创建有缓冲 channel |
| `new(7)`，Go 1.26+ | `*int` | 指针非 nil，`*p == 7` | 取得带初值变量的指针 |

**示例 M-02：有指针，不代表内部 map 已经创建。** 这是对官方 `new([]int)` 与 `make([]int, n)` 对比的扩展练习。[来源：Effective Go 的 make 小节](https://go.dev/doc/effective_go#allocation_make)

```go
// 运行示例：M-02
package main

import "fmt"

func main() {
	ps := new([]int)
	pm := new(map[string]int)
	s := make([]int, 0)
	m := make(map[string]int)

	fmt.Printf("%T %t %t\n", ps, ps == nil, *ps == nil)
	fmt.Printf("%T %t %t\n", pm, pm == nil, *pm == nil)
	fmt.Println(s == nil, m == nil)

	*ps = append(*ps, 5)
	*pm = make(map[string]int)
	(*pm)["score"] = 8
	fmt.Println(*ps, (*pm)["score"])
}
```

输出：

```text
*[]int false true
*map[string]int false true
false false
[5] 8
```

`ps` 和 `pm` 都保存了有效指针，因此都不等于 nil。但 `new` 只是创建了“一个切片变量”和“一个 map 变量”，这两个变量的初值仍然是 nil。nil 切片可以作为 `append` 的输入，且必须接收 `append` 返回的新切片；nil map 不能直接写入，所以代码先给 `*pm` 赋上新建的 map。

这一例的写法刻意用于比较类型，普通业务通常直接写 `s := make([]int, 0)`、`m := make(map[string]int)`，无需额外的指针层。只有函数确实要替换调用方保存的整个变量时，`*[]T` 或 `*map[K]V` 才可能有意义；很多时候直接返回更新后的值更清楚。

<a id="section-3-3"></a>

### 3.3 语法示例

<a id="section-3-3-1"></a>

#### 3.3.1 切片：区分长度、容量和已存在的元素

切片的基本形式是 `make([]T, length)` 或 `make([]T, length, capacity)`。长度表示当前可以索引的元素个数；容量表示从切片起点开始，该切片通过重新切片能够扩展到的最大长度。三下标切片可以进一步限制容量，例如底层数组有 10 个元素的 `a[:2:3]`，容量也只有 3。

对 `make([]T, length, capacity)`，创建时底层数组的长度就是指定容量；省略容量时，它等于长度。容量不是“当前已有多少个有效索引”，因此 `make([]int, 0, 10)[0]` 仍然越界。[官方说明：Go Slices—usage and internals](https://go.dev/blog/slices-intro)、[切片表达式](https://go.dev/ref/spec#Slice_expressions)。

**示例 M-03：预留容量后追加，并观察共享底层数组。**

```go
// 运行示例：M-03
package main

import "fmt"

func main() {
	s := make([]int, 2, 5)
	s[0], s[1] = 10, 20
	grown := append(s, 30)
	full := grown[:cap(grown)]

	fmt.Println(len(s), cap(s), s)
	fmt.Println(len(grown), cap(grown), grown)
	fmt.Println(full)

	full[0] = 99
	fmt.Println(s[0], grown[0])
}
```

输出：

```text
2 5 [10 20]
3 5 [10 20 30]
[10 20 30 0 0]
99 99
```

创建时底层数组有五个整数位置，最初都为 0；`s` 的长度只有 2。`append` 后的长度 3 没有超过容量 5，因此 `grown` 继续使用同一个底层数组，而 `s` 自己的长度仍是 2。`full` 把可见范围扩展到容量，能够看到还未写入的两个零值。最后通过 `full` 修改第一个元素，另外两个切片也能看到变化，因为它们指向相同的数组位置。

如果后续追加超出容量，`append` 可以返回引用新数组的切片，旧切片与新切片的共享关系可能因此改变。不要把“切片赋值”或“把切片传给函数”等同于“复制所有元素”。确实需要独立元素存储时，可以 `make` 后 `copy`；如果元素本身含指针、切片或 map，这仍然只是逐元素复制，是否还要深拷贝取决于业务。

长度和容量必须满足合法范围。常量形式的负数、`length > capacity` 等可在编译时发现；涉及变量的非法值可能在运行时触发 panic。学习预分配时不需要死记实现中的扩容倍数，代码正确性也不应依赖未在语言规范中保证的具体容量增长策略。

<a id="section-3-3-2"></a>

#### 3.3.2 map：容量提示不是长度，也不是上限

`make(map[K]V)` 创建空 map；`make(map[K]V, n)` 的 `n` 是初始空间的容量提示，不会产生 `n` 个键值对，也不限制日后只能存储 `n` 项。map 的长度由实际键值对数量决定，而且不能用 `cap(m)` 查询容量。[官方规范：make](https://go.dev/ref/spec#Making_slices_maps_and_channels)、[官方博客：Go maps in action](https://go.dev/blog/maps)

**示例 M-04：nil map 可查询，创建后才写入。**

```go
// 运行示例：M-04
package main

import "fmt"

func main() {
	var absent map[string]int
	value, ok := absent["go"]
	delete(absent, "go")
	fmt.Println(absent == nil, len(absent), value, ok)

	counts := make(map[string]int, 8)
	fmt.Println(counts == nil, len(counts))
	counts["go"] = 1
	alias := counts
	alias["go"]++
	fmt.Println(len(counts), counts["go"])
}
```

输出：

```text
true 0 0 false
false 0
1 2
```

第一行说明 nil map 的查询会得到元素类型的零值，同时 `ok == false`；查询不存在的键不等于“键存在且值为 0”，需要判断存在性时应保留 `ok`。对 nil map 调用 `len`、遍历、`delete` 都可以，向它写入键值对则会 panic。第二行表明容量提示 8 没有制造八个元素。最后 `alias` 与 `counts` 引用同一 map 数据，因此通过 `alias` 的写入也能由 `counts` 读到。

**常见误区：**`m2 := m1` 只复制 map 值，不会创建独立副本。若需要复制键值对，应创建新 map 后逐项赋值，或在合适版本中使用标准库的 map 复制工具；值内部的对象共享关系仍需另外考虑。不要为了得到固定输出依赖 map 遍历顺序，本例只按明确的键查询。

<a id="section-3-3-3"></a>

#### 3.3.3 channel：创建通信对象，明确是否有缓冲

`make(chan T)` 与 `make(chan T, 0)` 创建无缓冲 channel；`make(chan T, n)` 在 `n > 0` 时创建缓冲容量为 `n` 的 channel。有缓冲 channel 也不是“永远不会阻塞”：缓冲满时发送需要等待，缓冲为空且未关闭时接收也需要等待。[官方规范：channel 类型](https://go.dev/ref/spec#Channel_types)、[接收运算符](https://go.dev/ref/spec#Receive_operator)

**示例 M-05：同一 goroutine 内使用有缓冲 channel。**

```go
// 运行示例：M-05
package main

import "fmt"

func main() {
	ch := make(chan int, 2)
	fmt.Println(ch == nil, len(ch), cap(ch))
	ch <- 7
	ch <- 8
	fmt.Println(len(ch), cap(ch))

	first := <-ch
	second := <-ch
	fmt.Println(first, second, len(ch))

	close(ch)
	value, ok := <-ch
	fmt.Println(value, ok)
}
```

输出：

```text
false 0 2
2 2
7 8 0
0 false
```

缓冲容量为 2，所以两个发送可以先把值放进缓冲区，再由同一 goroutine 取出。关闭并取尽 channel 后，接收立即得到元素类型的零值和 `ok == false`。`len(ch)` 在这里表示当前缓冲中的元素数，不是总共发送过的次数；在真实并发程序里，观察到的长度也不能保证下一次操作一定不会阻塞。

`var ch chan int` 与 `make(chan int)` 不等价。前者是 nil channel，直接向它发送或从它接收都会永久阻塞；后者是已创建的无缓冲 channel，需要匹配的通信操作。本笔记不把会永久阻塞的写法放进可运行示例。关闭 nil channel 或已关闭的 channel，以及向已关闭的 channel 发送，都会 panic；仅接收已关闭 channel 则是合法操作。

<a id="section-3-3-4"></a>

#### 3.3.4 Go 1.26+：new 可以直接接受值表达式

旧教程常把 `new` 描述为“只能接收一个类型”。从 Go 1.26 开始，`new(x)` 还可以接收值表达式：先求出 `x` 的值，再创建一个以该值初始化的新变量，返回指向新变量的指针。传入无类型常量时，使用它的默认类型，例如 `new(7)` 的类型是 `*int`。未指定类型的 `nil` 本身不能作为参数。[Go 1.26 发布说明：语言变化](https://go.dev/doc/go1.26#language)

**示例 M-06：`new(x)` 是新的变量，`&x` 指向现有变量。本例要求 Go 1.26 或更新的语言版本。**

```go
// 运行示例：M-06
// 运行基线：Go 1.27.3（模块声明 go 1.27.0）
package main

import "fmt"

func main() {
	calls := 0
	next := func() int {
		calls++
		return 40 + calls
	}
	p := new(next())

	x := 7
	q := new(x)
	r := &x
	x = 9
	fmt.Println(*p, calls)
	fmt.Println(x, *q, *r)

	*q = 12
	fmt.Println(x, *q, *r)
	fmt.Printf("%T\n", new(7))
}
```

输出：

```text
41 1
9 7 9
9 12 9
*int
```

`next()` 在创建 `p` 时求值一次，因此调用次数是 1。`q := new(x)` 取得当时的整数 7 并初始化新的整数变量；`r := &x` 则始终指向现有变量 `x`。所以后来修改 `x`，`*r` 随之变化，`*q` 不变；修改 `*q` 也不会修改 `x`。

这里的“新变量”不承诺递归复制引用的数据。如果 `x` 是切片、map 或含指针的结构体，新变量初始化时仍遵守普通赋值语义，底层数据或指向的对象可能共享。`new(value)` 帮助简洁地表达带初值的指针，不等于新增了一套深拷贝机制。

在 Go 1.26 以前，等价的初始化意图可写成一个局部变量再取地址，例如 `value := 7; p := &value`。项目能否使用新语法还受模块语言版本约束；仅安装新编译器而保留过低的 `go.mod` 中的 `go` 指令，不能直接把新语法视为已启用。

<a id="section-3-4"></a>

### 3.4 总结

选择初始化方式时，可以按下面的需求表判断：

| 需求 | 常用写法 | 需要额外留意 |
|---|---|---|
| 普通零值变量 | `var x T` | 零值能否满足业务约束 |
| 零值变量的指针 | `new(T)` | 指向的切片、map、channel 仍可能是 nil |
| 已有变量的地址 | `&x` | 将来通过该指针修改的是同一个变量 |
| 指定字段的结构体指针 | `&T{Field: value}` | 未指定字段保留零值 |
| 指定长度或预留容量的切片 | `make([]T, length, capacity)` | 容量不等于可直接索引的长度 |
| 空的可写 map | `make(map[K]V)` | 容量提示不等于长度 |
| channel | `make(chan T, capacity)` | 缓冲容量决定通信何时等待 |
| 值表达式初始化的新变量指针 | `new(value)`，Go 1.26+ | 新变量仍遵守普通赋值的共享规则 |

不要用“`new` 在堆上，`make` 在栈上”或相反的口诀。对象最终放在栈、堆，或者被优化掉，是编译器根据使用方式决定的实现问题；`new`、`make`、`&T{}` 这些写法本身不能推出固定结论。把“值与指针的语义”先学准确，再用实际性能数据判断是否需要关注分配。[官方 FAQ：栈还是堆](https://go.dev/doc/faq#stack_or_heap)、[官方 GC 指南：值存放在哪里](https://go.dev/doc/gc-guide#Where_Go_Values_Live)

自测：若 `p := new(map[string]int)`，为什么 `p != nil` 仍不足以保证 `(*p)["x"] = 1` 可以执行？因为前者只检查 map 变量的地址，后者操作的是仍为 nil 的 map 值。给 `*p` 赋上 `make(map[string]int)` 后，才创建了可写的 map。

<a id="chapter-4"></a>

## 四、指针

指针把“保存一个值”和“保存访问某个变量的位置”区分开。学习重点不是打印内存地址，而是判断几个名字是否在访问同一个对象，以及通过其中一个名字修改后，其他地方是否能观察到变化。本章的输出不依赖具体地址数值。

<a id="section-4-1"></a>

### 4.1 指针的基本概念

如果变量 `x` 的类型是 `T`，`&x` 的类型就是 `*T`。声明中的 `*T` 表示“指向 T 的指针类型”，表达式中的 `*p` 表示“访问 p 指向的变量”；这两个星号含义相关，但出现在不同位置。一个指针可以是 nil，此时它不指向可供解引用的变量。[官方规范：指针类型](https://go.dev/ref/spec#Pointer_types)、[地址运算符](https://go.dev/ref/spec#Address_operators)

**示例 P-01：复制指针后，两个指针访问同一个整数。**

```go
// 运行示例：P-01
package main

import "fmt"

func main() {
	x := 10
	p := &x
	q := p
	fmt.Printf("%T %d %t\n", p, *p, p == q)

	*q = 25
	fmt.Println(x, *p, *q)

	y := 25
	fmt.Println(p == &y, *p == y)
}
```

输出：

```text
*int 10 true
25 25 25
false true
```

`q := p` 复制的是指针值，因此 `p` 和 `q` 指向同一个 `x`；它没有复制出第二个独立整数。`*q = 25` 改的是被指向的变量，所以读取 `x`、`*p`、`*q` 会得到同一个结果。`y` 是另一个整数变量，即使数值相同，也不与 `x` 具有相同对象身份。**指针比较回答“是否指向同一个变量”，值比较回答“当前内容是否相等”。**

一般业务可以依上述规则理解指针相等性；规范对不同零大小变量的指针是否相等有特殊余地，因此不要用两个 `&struct{}{}` 的比较来演示对象身份。[官方规范：比较运算符](https://go.dev/ref/spec#Comparison_operators)

需要特别区分下面三种写法：

| 写法 | 改变什么 |
|---|---|
| `p = &y` | 改变指针变量 p 保存的地址 |
| `*p = 30` | 改变 p 当前指向的变量 |
| `p = nil` | 清除 p 自己保存的指针值，不会把所有别名一起清空 |

组合类型要按结构阅读：`[]*int` 是“元素为整数指针的切片”，`*[]int` 是“指向整数切片变量的指针”，`*[3]int` 则是“指向三个整数构成的数组的指针”。它们的可用操作和复制效果不同，不能因为都含有星号就混用。声明 `var a, b *int` 时，`a` 和 `b` 都是指针；星号属于类型表达式 `*int`，并不是只修饰最近的一个变量名。

<a id="section-4-2"></a>

### 4.2 指针的使用场景

<a id="section-4-2-1"></a>

#### 4.2.1 函数需要修改调用方的对象

Go 的函数参数始终按值传递。传整数时复制整数，传指针时复制指针值；复制指针没有复制其指向的对象。因此，被调函数可以通过指针修改原对象，却不能仅通过给形参重新赋值，替换调用方的指针变量。[官方 FAQ：参数传递](https://go.dev/doc/faq#pass_by_value)

**示例 P-02：修改对象与让局部指针重新指向。**

```go
// 运行示例：P-02
package main

import "fmt"

func addTen(p *int) {
	*p += 10
}

func pointElsewhere(p *int) {
	other := 90
	p = &other
	*p = 100
	fmt.Println("inside:", *p)
}

func main() {
	x := 5
	p := &x
	addTen(p)
	fmt.Println("after add:", x)
	pointElsewhere(p)
	fmt.Println("outside:", x, *p)
}
```

输出：

```text
after add: 15
inside: 100
outside: 15 15
```

调用 `addTen(p)` 时，实参与形参中保存的指针值相同，所以形参解引用后访问到原来的 `x`。调用 `pointElsewhere(p)` 后，形参 `p` 起初也指向 `x`，但 `p = &other` 只改了这个局部形参，外层 `main` 中的 `p` 没有被赋值。函数返回时，外层仍然通过原指针访问数值为 15 的 `x`。

如果接口设计确实要求更新调用方的指针变量，可以让函数返回新指针，由调用方显式接收；另一个选项是传 `**T`，但多一层间接访问也会增加理解成本。这里先建立一层指针的模型，第五章进一步讨论函数参数与接收者的选择。

<a id="section-4-2-2"></a>

#### 4.2.2 用 nil 表示“没有提供”，保留默认状态的业务含义

有些数据需要三个状态，例如“未知年龄”“年龄为 0”“年龄为 18”。普通 `int` 只能保存整数，无法单独表达“没有提供”；`*int` 则可以用 nil 表示缺失，用非 nil 指针保存包括 0 在内的有效值。是否应使用指针取决于业务，不能把所有整数都改成指针。

**示例 P-03：JSON 中区分缺失字段与显式零值。** 本例根据标准库 `encoding/json.Marshal` 对指针和 `omitempty` 的规定自行编写；使用局部变量取地址，便于旧版本项目理解。Go 1.26+ 可把对应初值写成 `new(0)`。[来源：encoding/json.Marshal](https://pkg.go.dev/encoding/json#Marshal)

```go
// 运行示例：P-03
package main

import (
	"encoding/json"
	"fmt"
)

type Profile struct {
	Name string `json:"name"`
	Age  *int   `json:"age,omitempty"`
}

func main() {
	zero := 0
	profiles := []Profile{
		{Name: "unknown"},
		{Name: "newborn", Age: &zero},
	}
	for _, profile := range profiles {
		data, err := json.Marshal(profile)
		if err != nil {
			fmt.Println("encode:", err)
			return
		}
		fmt.Println(string(data))
	}
}
```

输出：

```text
{"name":"unknown"}
{"name":"newborn","age":0}
```

第一个 `Age` 为 nil，因为字段带有 `omitempty`，编码时省略该字段。第二个 `Age` 是非 nil 的指针，即使它指向 0，仍会输出 `"age":0`。如果去掉 `omitempty`，nil 指针的 JSON 表示将是 `null`。这也是“nil 与零值值对象不同”的实际用途。

<a id="section-4-2-3"></a>

#### 4.2.3 共享身份、较大对象和关联结构

指针也常用于让多个组件访问同一个对象，或表示链表、树等结构中的关联。较大的结构体按值传递在语义上会复制结构体，指针传递可能避免这份整体复制；但这并不意味着指针一定更快，间接访问、逃逸、GC 和缓存行为也会影响成本。先根据是否需要共享修改、是否需要表示缺失来设计，再通过基准测试评估性能。[官方 FAQ：参数与复制](https://go.dev/doc/faq#pass_by_value)、[GC 指南：逃逸分析](https://go.dev/doc/gc-guide#Escape_analysis)

指针本身不提供并发同步。如果两个 goroutine 共享一个对象并发生读写冲突，需要通过互斥锁、channel 或其他同步机制建立正确访问关系；“用了指针”只说明它们能访问同一数据，不能保证访问安全。[官方文档：Go 内存模型](https://go.dev/ref/mem)

<a id="section-4-3"></a>

### 4.3 指针的默认状态

`var p *int` 只声明一个指针变量，其初值是 nil；它不会顺便创建一个整数让 `p` 指向。`new(int)` 则确实创建一个整数变量并返回指针。两者之间的区别是“有没有被指向的变量”，不是整数值到底是不是 0。[官方规范：指针与解引用](https://go.dev/ref/spec#Address_operators)

**示例 P-04：先判断是否存在，再解引用。**

```go
// 运行示例：P-04
package main

import "fmt"

func readOr(p *int, fallback int) int {
	if p == nil {
		return fallback
	}
	return *p
}

func main() {
	var missing *int
	zero := new(int)
	fmt.Println(missing == nil, zero == nil)
	fmt.Println(readOr(missing, -1), readOr(zero, -1))

	outer := &missing
	fmt.Println(outer == nil, *outer == nil)
}
```

输出：

```text
true false
-1 0
false true
```

`readOr` 把“缺失时使用默认值”作为明确的函数约定：先判断 nil，再访问 `*p`，所以 `missing` 不会被解引用；`zero` 指向合法的整数变量，读取到 0。最后 `outer` 的类型是 `**int`，它指向的是指针变量 `missing` 本身，因此 `outer` 非 nil，但 `*outer` 仍然是 nil。这说明多层指针需要逐层判断，不能只检查最外一层。

不要把“所有指针函数都先判空”作为机械规则。若 nil 在业务上合法，可以像 `readOr` 一样定义含义；若 nil 代表调用方错误，应通过清晰的接口约定、构造方式和必要检查暴露问题，而不是静默把所有 nil 转成某个默认值。

以下是**故意错误的局部片段，不属于可运行示例**：

```go
var p *int
*p = 1 // 运行时 panic：p 没有指向可供写入的整数变量
```

修正方法可能是 `p = new(int)`，也可能是把一个已有变量的地址赋给它；应该选择符合对象生命周期和共享关系的方式。`p = nil` 也不是“释放对象”的通用动作：如果别处仍然持有指针，对象仍可被访问；对象何时被回收由可达性和运行时决定。

<a id="section-4-4"></a>

### 4.4 指针常见问题

<a id="section-4-4-1"></a>

#### 4.4.1 返回局部变量地址是否会悬空

在 Go 中，返回局部变量地址是合法用法。只要程序还需要通过有效引用使用该变量，编译器和运行时就必须保证其存储有效；不需要像某些低层语言那样手工安排这块普通 Go 内存的释放。[官方 FAQ：栈与堆](https://go.dev/doc/faq#stack_or_heap)

**示例 P-05：函数返回后，返回的指针仍能读写对象。**

```go
// 运行示例：P-05
package main

import "fmt"

type Counter struct {
	Value int
}

func newCounter(start int) *Counter {
	counter := Counter{Value: start}
	return &counter
}

func main() {
	a := newCounter(3)
	b := newCounter(3)
	a.Value++
	fmt.Println(a.Value, b.Value, a == b)
}
```

输出：

```text
4 3 false
```

两次调用分别创建各自的 `Counter`，因而修改 `a` 不会改变 `b`。这段代码表达的是对象关系，并不能据此断言每次调用一定发生一次堆分配：内联和逃逸分析都可能影响实际实现。若确有优化需要，可用编译器诊断观察当前构建结果，例如在保存该例的目录执行 `go build -gcflags=-m=3 .`；诊断内容可能随版本和编译上下文变化，不应当作语言保证。[官方 GC 指南：逃逸分析](https://go.dev/doc/gc-guide#Escape_analysis)

“内存由 GC 管理”也不等于“文件、连接和锁都会及时自动清理”。对象的内存生命周期与外部资源生命周期是不同问题；后面的 `defer` 一章专门讨论释放资源的安排。

<a id="section-4-4-2"></a>

#### 4.4.2 结构体按值复制，内部引用仍可能共享

结构体赋值会按其值语义复制各字段；整数、布尔值、数组等字段的值会被复制，切片、map、指针字段复制后则可能继续访问同一底层对象。不能根据“最外层是结构体”就断言整个数据图被深拷贝。[官方 FAQ：复制与参数传递](https://go.dev/doc/faq#pass_by_value)

**示例 P-06：同一次结构体复制，不同字段表现不同。**

```go
// 运行示例：P-06
package main

import "fmt"

type Record struct {
	Count  int
	Array  [2]int
	Items  []int
	Labels map[string]string
}

func main() {
	original := Record{
		Count:  1,
		Array:  [2]int{10, 20},
		Items:  []int{30, 40},
		Labels: map[string]string{"name": "old"},
	}
	copyValue := original
	copyValue.Count = 2
	copyValue.Array[0] = 99
	copyValue.Items[0] = 88
	copyValue.Labels["name"] = "new"

	fmt.Println(original.Count, original.Array, original.Items, original.Labels["name"])
	fmt.Println(copyValue.Count, copyValue.Array, copyValue.Items, copyValue.Labels["name"])

	copyValue.Items = []int{7}
	fmt.Println(original.Items, copyValue.Items)
}
```

输出：

```text
1 [10 20] [88 40] new
2 [99 20] [88 40] new
[88 40] [7]
```

`Count` 和固定数组 `Array` 在两个结构体中是各自的值，因此修改副本不会改变原值。`Items` 字段复制的是切片值，两个切片仍指向同一底层数组，所以修改元素会共享；`Labels` 也继续引用同一 map 数据。最后给 `copyValue.Items` 赋上另一个切片，只替换副本的字段，不会让 `original.Items` 自动改指向。

同样的分析方法适用于值接收者的方法：接收者按值复制，里面的切片或 map 仍可能共享。不要把“使用值接收者”直接解释为“方法无法改变任何调用方可见的数据”。

<a id="section-4-4-3"></a>

#### 4.4.3 为什么不能直接取 map 元素的地址

Go 中 `&x` 对普通操作数要求其可寻址，map 索引结果不属于可寻址变量。`m[key]` 可用作 map 赋值的左侧，但这项特殊能力不等于它能用于 `&m[key]`。这属于语言规则，无需依赖某一种 map 内部实现来解释。[官方规范：地址运算符](https://go.dev/ref/spec#Address_operators)、[索引表达式](https://go.dev/ref/spec#Index_expressions)

**示例 P-07：结构体 map 元素读出、修改、写回，或者保存指针。**

```go
// 运行示例：P-07
package main

import "fmt"

type User struct {
	Age int
}

func main() {
	values := map[string]User{"lee": {Age: 18}}
	u := values["lee"]
	u.Age++
	fmt.Println(values["lee"].Age)
	values["lee"] = u
	fmt.Println(values["lee"].Age)

	pointers := map[string]*User{"lee": &User{Age: 18}}
	pointers["lee"].Age++
	fmt.Println(pointers["lee"].Age)
}
```

输出：

```text
18
19
19
```

`u := values["lee"]` 得到的是结构体值，修改 `u` 后必须写回 map 才能更新存储的值。第二种设计在 map 中保存 `*User`，查询得到指针后，可以通过指针访问它指向的对象。指针设计增加了共享修改的含义，还要处理缺失键或 nil 元素；它不是对所有 `map[K]Struct` 的无条件改进。

以下片段均**不能通过编译**：

```go
// 假设 values 的类型是 map[string]User：
// p := &values["lee"]      // map 元素不可寻址
// values["lee"].Age = 20   // 不能直接给取出的结构体字段赋值
```

不要将这个结论套用到切片索引：切片的有效索引元素是可寻址的，通常可以写 `&items[i]`；但如果后来 `append` 导致切片转向新数组，之前保存的指针仍指向旧数组里的那个元素，不会自动改为新数组中的对应元素。共享关系必须结合具体表达式分析。

<a id="section-4-4-4"></a>

#### 4.4.4 普通 Go 指针不支持指针算术

对 `p *int` 写 `p++`、`p + 1` 不能像 C 指针那样访问下一个元素。操作连续元素时，使用数组或切片的索引和 `range`。`unsafe` 提供的底层能力有额外约束，不是学习普通指针时必需的工具，也不能用它绕过类型和生命周期分析。[官方 FAQ：为什么没有指针算术](https://go.dev/doc/faq#no_pointer_arithmetic)

同样，不是所有表达式都能直接取地址。例如整数常量、普通函数返回的整数结果，以及 map 索引结果，均不能直接写 `&`；结构体等复合字面量有规范明确允许的例外，因而 `&User{Age: 18}` 合法。Go 1.26 的 `new(value)` 为“从一个值表达式创建新变量并得到指针”提供了专门写法，但没有把所有表达式都变为可寻址。

<a id="section-4-5"></a>

### 4.5 总结

遇到指针相关代码，可以依次做四步判断：确定变量类型；区分正在操作指针变量还是被指向的变量；检查 nil 的业务含义；最后检查对象是否与其他地方共享。下面的表适合复习时快速定位误解。

| 常见说法或写法 | 准确理解 |
|---|---|
| “传指针就是引用传递” | 指针值本身仍被复制；共享的是被指向的对象 |
| `p = &other` | 让 p 重新指向，不等于修改旧对象 |
| `*p = other` | 修改当前指向的对象，要求 p 非 nil |
| `var p *T` | p 的初始值为 nil，没有自动创建 T 对象 |
| `new(T)` | 创建零值 T，并返回指向它的指针 |
| `new(x)`，Go 1.26+ | 根据 x 的值初始化新变量，不等于 `&x` |
| 返回 `&local` | 合法，Go 保证有效引用所需的变量生命周期 |
| `structCopy := original` | 复制结构体值，内部引用的数据不一定独立 |
| `&m[key]` | 不合法，map 索引结果不可寻址 |
| `p = nil` | 只修改 p，不保证对象立即被回收 |
| “指针一定更快” | 需要结合语义、逃逸、访问模式和测量结果判断 |

可以用 P-02 自测：若把 `pointElsewhere` 中的 `p = &other` 改成 `*p = other`，外层 `x` 就会先被改成 90，紧接着 `*p = 100` 再把它改成 100，最终外层输出变为 `outside: 100 100`。决定效果的正是赋值左侧有没有解引用，而不是函数名或形参名。

本章的规则与示例依据：[Go 规范](https://go.dev/ref/spec)、[Effective Go：Data](https://go.dev/doc/effective_go#data)、[Go FAQ：Pointers and Allocation](https://go.dev/doc/faq#pointers)、[Go Slices](https://go.dev/blog/slices-intro)、[Go maps in action](https://go.dev/blog/maps)、[Go 1.26 发布说明](https://go.dev/doc/go1.26#language)。

<a id="chapter-5"></a>

## 五、函数的高阶内容

这一章把前面的函数、指针和方法放在一起，解决三个实际问题：函数如何修改调用方的数据，方法为什么会影响接口实现，以及失败发生后怎样返回信息并释放资源。阅读时始终区分“一个变量本身”和“这个变量引用的对象”；处理失败时则区分普通返回路径和 `panic` 展开调用栈的路径。

本章完整示例均为依据官方规则编写的独立程序，统一按 Go 1.27.3 工具链和 `go 1.27.0` 模块语言设置组织。下面出现的 `any` 是 `interface{}` 的别名，用于接收任意类型的值。

<a id="section-5-1"></a>

### 5.1 指针作为函数参数

<a id="section-5-1-1"></a>

#### 5.1.1 传入地址，仍然是值传递

函数声明 `func update(p *int)` 表示形参 `p` 的类型是 `*int`。调用 `update(&n)` 时，实参中的指针值被复制到 `p`。此后调用方的指针与函数内的 `p` 可以指向同一个整数，因此 `*p = 20` 会修改那个整数；但 `p = other` 只会修改形参变量 `p`。

不要把“通过指针修改到了外面的数据”解释成 Go 改用了引用传递。应分别判断赋值左边是什么：左边是 `p`，修改本地指针；左边是 `*p`，修改它指向的变量。参数包含切片、映射或接口时，也同样先复制参数值，具体会共享哪些数据取决于值中包含的引用。

**示例依据：**[语言规范：函数调用](https://go.dev/ref/spec#Calls)、[地址与解引用运算](https://go.dev/ref/spec#Address_operators)。下面用同一个整数连续观察修改目标、修改局部指针、修改调用方指针三种行为。

```go
// 运行示例：E-01
package main

import "fmt"

func addFive(p *int) {
	*p += 5
}

func rebindLocal(p *int) {
	other := 99
	p = &other
	fmt.Println("inside rebind:", *p)
}

func replacePointer(pp **int) {
	other := 77
	*pp = &other
}

func main() {
	n := 10
	p := &n

	addFive(p)
	fmt.Println("after add:", n)

	rebindLocal(p)
	fmt.Println("after rebind:", *p)

	replacePointer(&p)
	fmt.Println("after replace:", *p, "original:", n)
}
```

输出：

```text
after add: 15
inside rebind: 99
after rebind: 15
after replace: 77 original: 15
```

执行过程如下：

1. `p := &n` 让 `p` 指向 `n`；`addFive` 中的参数副本也指向 `n`，所以 `n` 变成 `15`。
2. `rebindLocal` 把自己的参数副本改为指向 `other`，函数内打印 `99`，调用方的 `p` 仍然指向 `n`。
3. `replacePointer(&p)` 传入的是“调用方指针变量 `p` 的地址”。形参 `pp` 是一个指针副本，但 `*pp` 正是调用方的 `p`，所以对 `*pp` 赋值改变了调用方的指向。
4. `p` 改指向新的整数后，原来的 `n` 没有消失，值仍为 `15`。局部变量 `other` 的地址离开函数后仍可安全使用，Go 会保证可达对象所需的生命周期；它具体放在栈还是堆上由实现决定。

最后一点可参见 [Go FAQ：栈还是堆](https://go.dev/doc/faq#stack_or_heap)。这里使用 `**int` 是为了观察机制。业务接口如果需要替换一个指针，常常直接返回新的 `*T`，由调用方写 `p = createValue()`，会更容易阅读。

<a id="section-5-1-2"></a>

#### 5.1.2 什么时候需要指针参数

| 函数的意图 | 通常选择 | 调用方需要理解的约定 |
|---|---|---|
| 读取一个小的普通值 | `T` | 被调用方得到值的副本 |
| 修改调用方的结构体字段 | `*T` | 多处可能观察到同一对象的变化 |
| 表示“可能没有这个对象” | `*T`，并定义 `nil` 的含义 | `nil` 是缺省值、无对象还是非法输入 |
| 产生新对象供调用方使用 | 返回 `*T` 或 `T` | 新值的所有权与共享方式要明确 |
| 修改已有指针变量的指向 | 返回新指针，或少数场景使用 `**T` | 修改的是指针变量，而非原对象 |

“结构体很大”可以成为考虑指针参数的原因，但不能据此断言指针一定更快。共享、逃逸和间接访问都会影响代价；先选择符合语义的接口，再对确有性能要求的代码测量。指针也不会自动提供并发安全：两个 goroutine 通过同一指针写数据，仍需要相应的同步措施。

**常见误区：**在函数里执行 `p = new(T)`，不会自动给调用方的空指针完成初始化；直接执行 `*p = value` 前必须保证 `p != nil`；“传入一个指针”也不等于“获得这个对象的独占使用权”。本节示例的三个修改函数都约定参数非 `nil`。

<a id="section-5-2"></a>

### 5.2 使用指针作为方法的接收者

<a id="section-5-2-1"></a>

#### 5.2.1 接收者也会被复制

方法声明中的接收者与普通参数遵循相同的传值规则。`func (c Counter) Add()` 得到一个 `Counter` 副本；`func (c *Counter) Add()` 得到一个指针副本。后者能够通过指针访问原对象，但如果方法仅执行 `c = &Counter{}`，仍然只是在改自己的接收者变量。

对需要修改对象状态的方法，指针接收者通常能直接表达意图；对小型、适合按值使用的数据，值接收者可能更合适。接收者类型还会改变方法集，从而影响类型是否满足接口。选择时不能只检查“这一行方法调用能否编译”。参见 [Go FAQ：方法应使用值还是指针](https://go.dev/doc/faq#methods_on_values_or_pointers)。

<a id="section-5-2-2"></a>

#### 5.2.2 方法集与自动取地址是两条规则

以下表格只讨论没有嵌入字段的普通已定义类型 `T`，且 `T` 本身不是指针或接口类型；嵌入字段的方法提升还有额外规则。

| 声明的方法 | `T` 的方法集包含吗 | `*T` 的方法集包含吗 |
|---|---|---|
| `func (x T) Read()` | 包含 | 包含 |
| `func (x *T) Change()` | 不包含 | 包含 |

如果 `x` 是可以取地址的 `T` 变量，`x.Change()` 可以作为 `(&x).Change()` 的简写。但这项调用便利不会把 `Change` 加入 `T` 的方法集，也不会让一个 `T` 值自动满足要求 `Change` 的接口。

**示例依据：**[语言规范：方法集](https://go.dev/ref/spec#Method_sets)、[Go Wiki：MethodSets](https://go.dev/wiki/MethodSets)。

```go
// 运行示例：E-02
package main

import "fmt"

type Counter struct {
	n int
}

func (c Counter) Value() int {
	return c.n
}

func (c *Counter) Increment() {
	c.n++
}

type Incrementer interface {
	Increment()
}

func main() {
	c := Counter{}
	c.Increment() // c 可取地址，等价于 (&c).Increment()
	fmt.Println(c.Value())

	_, valueOK := any(c).(Incrementer)
	_, pointerOK := any(&c).(Incrementer)
	fmt.Println(valueOK, pointerOK)

	var inc Incrementer = &c
	inc.Increment()
	fmt.Println(c.Value())
}
```

输出：

```text
1
false true
2
```

这里的 `any(c).(Incrementer)` 使用带 `ok` 的类型断言，检查接口中保存的具体值是否实现 `Incrementer`；不满足时得到 `false`，不会触发 `panic`。第一条方法调用虽然成功，`Counter` 仍未实现该接口；保存 `&c` 时，动态类型是 `*Counter`，其方法集包含 `Increment`，所以断言成功。

因此，若把 `var inc Incrementer = &c` 改成 `var inc Incrementer = c`，会在编译期报错。对于 `map[string]Counter` 的元素 `m["a"]`，也不能直接调用指针接收者方法，因为映射元素不可取地址。可以取出值到临时变量、修改后放回，或者根据业务需求把映射元素设计为 `*Counter`。切片元素则可以取地址，两者不要混淆。

示例为了展示方法集而混用了值接收者与指针接收者。实际设计同一类型的公开方法时，应让接收者选择保持清楚、一致，尤其不要在不了解接口影响的情况下随意更改已有接收者类型。

<a id="section-5-2-3"></a>

#### 5.2.3 值接收者不会深复制切片指向的数据

“值接收者不会修改原结构体字段”不能进一步推导成“方法不可能影响外部数据”。如果结构体的字段是切片、映射或指针，复制结构体会复制这些字段的值，复制后的字段仍可能引用同一份数据。

**示例依据：**[语言规范：值的表示](https://go.dev/ref/spec#Representation_of_values)、[结构体类型](https://go.dev/ref/spec#Struct_types)。下面专门把普通字段、切片元素和切片长度放在一起比较。

```go
// 运行示例：E-03
package main

import "fmt"

type Basket struct {
	Name string
	Tags []string
}

func (b Basket) EditCopy() {
	b.Name = "copy"
	b.Tags[0] = "shared"
	b.Tags = append(b.Tags, "extra")
	fmt.Println("inside:", b.Name, b.Tags)
}

func main() {
	b := Basket{Name: "original", Tags: []string{"old"}}
	b.EditCopy()
	fmt.Println("outside:", b.Name, b.Tags)
}
```

输出：

```text
inside: copy [shared extra]
outside: original [shared]
```

`b.Name` 在副本中被重新赋值，原结构体的 `Name` 不变；`b.Tags[0]` 修改的是此时共享的底层数组元素，调用方能够观察到；`append` 的结果重新赋给方法内部的切片字段，只改变副本中的切片值，调用方切片的长度仍然是 `1`。如果希望完全隔离可变数据，应按需求复制相应数组、映射或其他对象，不能只把方法接收者从 `*T` 改成 `T`。

<a id="section-5-2-4"></a>

#### 5.2.4 `nil` 指针可以成为接收者，能否工作取决于方法体

调用指针接收者方法时，接收者可以是一个类型明确的 `nil` 指针。只要方法体在解引用前妥善处理，就可以把 `nil` 设计成有意义的状态。下面把 `nil` 解释为“使用默认配置”，同时保留“明确配置为零”的不同含义。

**示例依据：**[语言规范：方法调用](https://go.dev/ref/spec#Calls)与[地址运算](https://go.dev/ref/spec#Address_operators)。

```go
// 运行示例：E-04
package main

import "fmt"

type Config struct {
	Retries int
}

func (c *Config) MaxRetries() int {
	if c == nil {
		return 3
	}
	return c.Retries
}

func main() {
	var missing *Config
	explicitZero := &Config{Retries: 0}
	fmt.Println(missing.MaxRetries())
	fmt.Println(explicitZero.MaxRetries())
}
```

输出：

```text
3
0
```

不要把这个例子理解成“对空指针调用所有方法都安全”。如果方法第一行就是 `return c.Retries`，就会在访问字段时解引用空指针并触发 `panic`；如果通过空指针调用一个值接收者方法，还需要先取得它指向的值，同样不能依靠这个例子的保护。是否接受 `nil` 应当是该方法明确的契约，而不是调用者的猜测。

<a id="section-5-3"></a>

### 5.3 函数异常处理

目录使用了“异常处理”这个常见术语。在 Go 中，先学会分别阅读两条路径：函数正常返回一个非 `nil` 的 `error`；或者发生 `panic`，当前正常执行路径中断，并沿同一 goroutine 的调用栈执行已经注册的延迟调用。`recover` 只参与后一条路径。

<a id="section-5-3-1"></a>

#### 5.3.1 使用 `panic` 和 `recover` 进行异常处理

**一、`panic` 后执行到哪里**

发生 `panic` 后，不会继续运行触发位置后面的普通语句。当前函数开始执行自己的延迟调用；如果没有被恢复，展开过程继续到调用者。一旦适当位置的延迟函数成功 `recover` 并正常结束，该恢复边界中的剩余延迟调用仍会执行，然后边界函数返回给它的调用者。

关键区别是：**恢复后不会回到原来触发 `panic` 的那一行继续往下执行。**下面的 `boundary` 是恢复边界，`danger` 是发生故障的被调用函数，`main` 是边界之外的调用者。

**示例依据：**[语言规范：处理 panic](https://go.dev/ref/spec#Handling_panics)、[官方博客：Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)。本例按官方调用栈展开规则重新设计了输出标记。

```go
// 运行示例：E-05
package main

import "fmt"

func danger() {
	defer fmt.Println("danger: cleanup")
	fmt.Println("danger: start")
	panic("broken")
	fmt.Println("danger: after panic") // 不会执行
}

func boundary() {
	defer func() {
		if value := recover(); value != nil {
			fmt.Println("boundary: recovered", value)
		}
	}()

	danger()
	fmt.Println("boundary: after danger") // 也不会执行
}

func main() {
	boundary()
	fmt.Println("main: after boundary")
}
```

输出：

```text
danger: start
danger: cleanup
boundary: recovered broken
main: after boundary
```

`danger` 的清理先执行，之后才到 `boundary` 的恢复函数。被放弃的调用状态不会重新进入，因此两个写着 `after` 的内部打印都不出现；`boundary` 返回后，`main` 的下一条语句才继续执行。

如果这个边界要向调用方报告失败，通常应返回 `error`，并在恢复函数中设置具名错误返回值。只打印错误然后结束，可能让调用者误以为操作成功。5.3.5 的综合示例会把这一步也补齐。

**二、`recover` 要由延迟函数直接调用**

`recover` 在正常执行过程中返回 `nil`。即使当前正在展开 `panic`，把它藏进延迟函数调用的另一个辅助函数中，也不能完成恢复。直接性指实际调用关系，不是要求源代码必须写成匿名函数：`defer handler()` 是可行的，前提是 `handler` 本身直接调用 `recover`；若 `handler` 再调用 `helper`，由 `helper` 调用 `recover`，则不能恢复外层的 `panic`。

**示例依据：**[语言规范：`recover` 的调用条件](https://go.dev/ref/spec#Handling_panics)。

```go
// 运行示例：E-06
package main

import "fmt"

func probe() any {
	return recover()
}

func run() {
	defer func() {
		fmt.Println("direct:", recover())
	}()
	defer func() {
		fmt.Println("indirect:", probe())
	}()
	panic("boom")
}

func main() {
	fmt.Println("outside:", recover())
	run()
	fmt.Println("finished")
}
```

输出：

```text
outside: <nil>
indirect: <nil>
direct: boom
finished
```

在 `main` 正常执行时，`recover()` 没有正在处理的 `panic`。`run` 中后注册的延迟函数先执行，但它通过 `probe` 间接调用 `recover`，因此得到 `nil`，展开过程仍继续；前面注册的延迟函数直接恢复了 `"boom"`。同理，单独写 `defer recover()` 不能替代一个在其函数体内调用 `recover` 的延迟函数。

**三、恢复边界必须位于发生 `panic` 的同一个 goroutine**

在创建 goroutine 的父函数里放置 `recover`，并不能捕获新 goroutine 中的 `panic`。两者有不同的调用栈。如果一个任务需要独立的恢复边界，应把边界放到该 goroutine 的入口调用链中，再把结果明确传回协调者。

下面的通道只用于把任务结果送回 `main`；主 goroutine 收到结果后才打印，因此输出顺序确定。它并不是依靠 `main` 恢复子 goroutine 的 `panic`。

**示例依据：**[语言规范：处理 panic](https://go.dev/ref/spec#Handling_panics)与[Go 语句](https://go.dev/ref/spec#Go_statements)。

```go
// 运行示例：E-07
package main

import "fmt"

func worker(done chan<- string) {
	result := "worker ok"
	defer func() {
		if value := recover(); value != nil {
			result = fmt.Sprintf("worker recovered: %v", value)
		}
		done <- result
	}()

	panic("job failed")
}

func main() {
	done := make(chan string, 1)
	go worker(done)
	fmt.Println(<-done)
	fmt.Println("main continues")
}
```

输出：

```text
worker recovered: job failed
main continues
```

本例为了观察隔离边界而恢复指定任务的故障。真实任务应约定恢复后如何报告失败、是否允许继续使用共享状态、由谁负责剩余清理。恢复意味着停止了这次展开，不代表中断之前的修改已经自动回滚。

<a id="section-5-3-2"></a>

#### 5.3.2 `panic` 的场景

先从“调用方是否应该把这种失败作为正常结果处理”判断接口。调用方输入不合法、文件不存在、网络请求失败，通常可以预期，应通过 `error` 返回。程序违反内部不变量或发生运行时非法操作，才可能进入 `panic` 路径。确有深层递归提前退出的需求时，包内部也可能使用 `panic`，但应在清楚的边界转换成对外的错误结果。

| 情况 | 典型表现 | 处理思路 |
|---|---|---|
| 显式调用 `panic(value)` | 使用传入值说明故障 | 调用者应能理解该 API 为什么允许发生 panic |
| 运行时越界、空指针解引用 | 运行时触发 panic | 修复边界条件或对象状态，避免靠恢复代替正常检查 |
| 向 `nil` 映射写入、向已关闭通道发送 | 违反相应操作规则 | 修正初始化或资源生命周期 |
| 输入解析失败、找不到记录 | 通常是可预期失败 | 返回 `error`，让调用方决定补救或退出 |
| `os.Exit` 终止进程 | 立即结束进程 | 它不是 panic，不能用 recover 捕获 |

其中运行时非法操作的规则见[语言规范：运行时 panic](https://go.dev/ref/spec#Run_time_panics)、[索引表达式](https://go.dev/ref/spec#Index_expressions)、[发送语句](https://go.dev/ref/spec#Send_statements)；进程退出行为见 [`os.Exit`](https://pkg.go.dev/os#Exit)。不要将“运行时报告了严重错误”全部当成可恢复的语言级 `panic`。

`panic` 可以接收 `error`，也可以接收字符串等其他值。`panic(errors.New("bad state"))` 中虽然使用了 `error` 类型的值，控制流仍然是 `panic`；只有 `return err` 才是向调用方正常返回一个错误结果。

**版本细节：`panic(nil)`。**从 Go 1.21 的默认行为起，`panic(nil)` 会形成类型为 `*runtime.PanicNilError` 的运行时 panic；在正确的延迟函数内恢复时得到非 `nil` 值。旧教程中“`panic(nil)` 被恢复后仍返回 `nil`”的描述不应直接套到当前默认行为上。历史兼容设置 `GODEBUG=panicnil=1` 以及旧模块版本会影响此行为，本文示例不依赖该旧行为。参见 [Go 1.21 发布说明](https://go.dev/doc/go1.21#language)与 [`runtime.PanicNilError`](https://pkg.go.dev/runtime#PanicNilError)。

<a id="section-5-3-3"></a>

#### 5.3.3 函数错误处理的 `error` 接口

**一、错误是实现接口的值**

预声明的 `error` 接口可以用下面的形式理解。这是接口定义说明，不是完整程序，也无需在自己的程序里重新声明它。

```go
type error interface {
    Error() string
}
```

一个类型只要提供所需的 `Error() string` 方法，就可以作为错误值返回。常用的函数签名是 `func operation(...) (Result, error)`：成功时返回有效结果和 `nil`；失败时返回相应错误。失败时的结果是否包含可用的部分数据，应由具体 API 说明，不能一律假设任何非 `nil` 错误都意味着“所有结果无效”。例如 `io.Reader` 允许同时返回已读字节数与错误。

官方基础说明见 [Error handling and Go](https://go.dev/blog/error-handling-and-go)、[语言规范：Errors](https://go.dev/ref/spec#Errors)与 [`io.Reader`](https://pkg.go.dev/io#Reader)。一个错误被返回以后，不会自动像 `panic` 那样展开上层调用栈；调用者必须选择处理、包装后返回，或者明确忽略。

**二、使用 `%w` 保留错误关系，用 `errors.Is` 判断类别**

`errors.New` 适合创建简单错误值；`fmt.Errorf` 可以补充操作与参数等上下文。如果需要让上层识别原始错误，应使用 `%w` 包装，而不是只把错误文字拼接起来。判断某个哨兵错误是否位于包装链中，应使用 `errors.Is`，不要把提示字符串当成稳定的程序接口。

**示例依据：**[`errors.New` 与 `errors.Is`](https://pkg.go.dev/errors)、[`fmt.Errorf`](https://pkg.go.dev/fmt#Errorf)、[Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)。本例自行设计了可查询的记录名。

```go
// 运行示例：E-08
package main

import (
	"errors"
	"fmt"
)

var ErrMissing = errors.New("item missing")

func lookup(name string) (int, error) {
	if name == "go" {
		return 1, nil
	}
	return 0, fmt.Errorf("lookup %q: %w", name, ErrMissing)
}

func main() {
	_, err := lookup("rust")
	fmt.Println(err)
	fmt.Println(err == ErrMissing)
	fmt.Println(errors.Is(err, ErrMissing))
	fmt.Println(errors.Is(errors.New("item missing"), ErrMissing))
}
```

输出：

```text
lookup "rust": item missing
false
true
false
```

包装后的错误本身不是 `ErrMissing`，所以直接比较为 `false`；包装关系仍保留了 `ErrMissing`，`errors.Is` 能找到它。最后一次调用重新创建了另一个错误值，文字虽然相同，也不意味着它与原哨兵错误相同。

`%w` 不只是增加一段描述，它使底层错误成为调用者可检查的行为。如果一个包承诺返回包装后的某个错误，上层可能据此分支；将来更换底层实现时需要考虑兼容性。是否暴露具体底层错误应是接口设计决定。官方错误包装文章的 “Whether to Wrap” 专门讨论了这一点。

**三、使用 `errors.As` 获取结构化错误信息**

当调用方需要错误中的字段，比如哪个输入字段出错、哪个路径打开失败，只知道错误的文字通常不够。可以定义错误类型，再通过 `errors.As` 沿包装关系找到该类型的值。下面的 `*FieldError` 既提供人能读懂的信息，也保留了程序可读取的 `Field` 和 `Value` 字段。

**示例依据：**[`errors.As`](https://pkg.go.dev/errors#As)及官方[错误包装文章中关于指向错误变量的指针](https://go.dev/blog/go1.13-errors)的说明。

```go
// 运行示例：E-09
package main

import (
	"errors"
	"fmt"
)

type FieldError struct {
	Field string
	Value int
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("%s cannot be %d", e.Field, e.Value)
}

func validateAge(age int) error {
	if age < 0 {
		return &FieldError{Field: "age", Value: age}
	}
	return nil
}

func register(age int) error {
	if err := validateAge(age); err != nil {
		return fmt.Errorf("register: %w", err)
	}
	return nil
}

func main() {
	err := register(-1)
	fmt.Println(err)

	var fieldErr *FieldError
	if errors.As(err, &fieldErr) {
		fmt.Println(fieldErr.Field, fieldErr.Value)
	}
}
```

输出：

```text
register: age cannot be -1
age -1
```

这里 `fieldErr` 本来是一个 `nil` 的 `*FieldError` 变量。把 `&fieldErr` 传给 `errors.As`，是为了让函数能够修改这个变量，填入找到的错误指针。因此参数的实际类型是 `**FieldError`，与 5.1 中修改调用方指针的原理一致。

不能改成 `errors.As(err, fieldErr)`：此时传入的是空指针，也没有给函数提供可写入结果变量的正确目标。`errors.As` 要求目标是适当类型的非 `nil` 指针，目标不合法会触发 `panic`。只有匹配成功后，才按目标类型使用所获得的信息。

**Go 1.26+ 补充：**标准库增加了 [`errors.AsType[E error]`](https://pkg.go.dev/errors#AsType)，可以直接返回匹配的错误值和布尔结果，例如在本例中写 `fieldErr, ok := errors.AsType[*FieldError](err)`。当前文档建议多数适用场景优先使用它。本节同时保留 `errors.As` 与 `errors.AsType`，以比较目标指针写法和类型参数写法。

**四、类型化的空指针不等于空 `error`**

接口值是否为 `nil`，取决于接口本身有没有保存动态类型和值。把一个类型为 `*Failure`、值为 `nil` 的指针装入 `error`，仍然保留了动态类型 `*Failure`，所以得到的接口不等于 `nil`。

**示例依据：**[Go FAQ：Why is my nil error value not equal to nil?](https://go.dev/doc/faq#nil_error)。下例用一个简短错误类型复现该问题，不依赖打印错误内容来判断状态。

```go
// 运行示例：E-10
package main

import "fmt"

type Failure struct{}

func (f *Failure) Error() string {
	return "operation failed"
}

func wrongResult() error {
	var failure *Failure
	return failure
}

func correctResult() error {
	return nil
}

func main() {
	fmt.Println(wrongResult() == nil)
	fmt.Println(correctResult() == nil)
	fmt.Printf("%T\n", wrongResult())
}
```

输出：

```text
false
true
*main.Failure
```

修复方向是让成功路径明确返回 `nil`，而不是先声明一个具体错误指针、最后无条件把它转换为 `error` 返回。这个问题与“某个指针接收者方法能否接受 `nil`”也不同：即使 `Error` 方法可以安全处理空接收者，装有该类型的接口依然不是 `nil` 接口值。

**五、需要保留多个失败时使用错误树**

读取失败后关闭资源也可能失败。只覆盖原 `err` 会丢掉前一个原因；只保留原 `err` 则会丢掉关闭失败。Go 1.20 起的 [`errors.Join`](https://pkg.go.dev/errors#Join) 可以保留多个错误，忽略其中的 `nil`，并在全部参数为 `nil` 时返回 `nil`。这时包装关系可能形成树，`errors.Is` 与 `errors.As` 可以遍历它。

`errors.Unwrap` 只处理 `Unwrap() error` 的单个包装层，不会替开发者遍历 `errors.Join` 产生的 `Unwrap() []error` 多分支结构。因此识别底层错误时，应直接选择适用的 `Is`、`As` 或 `AsType`，而不是手写只认识单链的循环。下一小节的资源示例会实际组合读取错误与关闭错误。

<a id="section-5-3-4"></a>

#### 5.3.4 `error` 和 `panic` 的区别

| 比较点 | 返回 `error` | 触发 `panic` |
|---|---|---|
| 控制流 | 函数正常返回到直接调用者 | 当前普通执行路径中断并展开调用栈 |
| 调用方如何处理 | 检查返回值并分支 | 在同一 goroutine 的适当边界设置直接调用 `recover` 的延迟函数 |
| 清理动作 | 函数返回时执行已注册的 defer | 展开过程中执行相应函数的已注册 defer |
| 是否自动传播到更上层 | 不会，上层需要选择返回错误 | 在未恢复的情况下沿调用链继续展开 |
| 典型用途 | 可预期的操作失败、校验失败、外部依赖失败 | 内部契约被破坏、运行时非法操作，或有明确边界的内部控制流 |
| 调用者能否继续 | 检查后可按业务选择继续 | 恢复边界返回后，其调用者可以继续；不回到原 panic 点 |

这里的“可继续”只描述控制流，并不保证业务状态一致。例如函数先从账户扣款，再在更新账单时失败，返回错误或恢复 `panic` 都不会自动撤销已经完成的扣款。需要原子性时，应设计事务、补偿动作或先验证后提交的流程。

普通库函数通常应让调用者通过返回值处理可预期的失败。若内部算法使用 `panic` 快速退出，也应明确限定可恢复的故障类型，在包的出口转换为可检查的错误，而不是把每个未知 `panic` 都转换成“成功结束”。官方博客给出了内部使用 `panic`、外部仍提供错误返回值的惯例说明：[Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)。

学习时可以依次问三个问题：这是不是调用者预期会遇到的失败？失败发生前已经改变了哪些状态？当前函数有能力补救，还是只能加上上下文交给上层？回答这三个问题，比把所有失败都写成同一种处理模板更可靠。

<a id="section-5-3-5"></a>

#### 5.3.5 结合 `defer`、`panic` 和 `recover` 处理资源泄漏


**一、清理的基础是资源生命周期**

文件、锁和连接等资源一旦成功取得，就应明确由谁释放，以及在什么函数返回时释放。最常见的局部资源模式是：取得资源，检查错误，然后立即注册 `defer`。失败时资源可能根本没有取得，不能先假设返回的指针一定有效。

对文件而言，`defer f.Close()` 可以避免遗漏关闭，但它会丢弃 `Close` 返回的错误。关闭是否需要影响整个操作结果，要由操作语义决定；例如写入的完整性可能与关闭阶段的结果相关。不能把“已经安排关闭”和“已经妥善报告关闭失败”当成同一件事。参见 [`os.Open`](https://pkg.go.dev/os#Open)、[`File.Close`](https://pkg.go.dev/os#File.Close)。

GC 管理内存，也不能作为及时归还文件描述符、连接或锁的保证。`recover` 解决的是控制流恢复，更不会替开发者寻找并关闭遗失的资源。清理应当在资源所在的作用范围内安排好，即便没有任何 `recover`，`panic` 展开时这些延迟调用仍会参与执行。

**二、同时保留读取错误与关闭错误**

下面用一个实现 `io.ReadCloser` 的小对象，确定性地模拟“读取失败，关闭也失败”。它的作用是让两个错误分支都能观察到，避免依赖磁盘故障。`readAndClose` 明确接管传入对象的关闭责任，调用方不应再重复关闭。

**示例依据：**[`io.ReadCloser`](https://pkg.go.dev/io#ReadCloser)、[`io.ReadAll`](https://pkg.go.dev/io#ReadAll)、[`errors.Join`](https://pkg.go.dev/errors#Join)与[具名返回值和 defer 的规则](https://go.dev/ref/spec#Defer_statements)。

```go
// 运行示例：E-11
package main

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrRead  = errors.New("read failed")
	ErrClose = errors.New("close failed")
)

type failingReadCloser struct{}

func (failingReadCloser) Read(p []byte) (int, error) {
	return 0, ErrRead
}

func (failingReadCloser) Close() error {
	return ErrClose
}

func readAndClose(r io.ReadCloser) (data []byte, err error) {
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close resource: %w", closeErr))
		}
	}()
	return io.ReadAll(r)
}

func main() {
	data, err := readAndClose(failingReadCloser{})
	fmt.Println("bytes:", len(data))
	fmt.Println("read error:", errors.Is(err, ErrRead))
	fmt.Println("close error:", errors.Is(err, ErrClose))
}
```

输出：

```text
bytes: 0
read error: true
close error: true
```

`io.ReadAll` 返回读取错误后，`return` 先把返回表达式的结果赋给具名结果变量 `data` 与 `err`。随后延迟函数运行，关闭失败被包装并与已有 `err` 合并；调用方最终得到的错误保留了两个原因。这里不需要 `panic`，也不需要 `recover`，只是把正常返回路径上的清理做完整。

如果只有关闭失败，原 `err` 为 `nil`，最终仍能返回关闭错误；如果两个步骤都成功，`err` 保持 `nil`。而当读取路径触发 `panic` 时，已注册的关闭函数仍会执行，但外层是否恢复这个 `panic` 是另一项明确的设计决定。

**三、真实文件、已知内部 panic 与对外错误返回**

下面创建临时文件，读取后触发一个特意设计的内部解析中断。示例只恢复私有的 `parseAbort`，未知 `panic` 会再次抛出；它展示一种边界组织方式，不表示普通输入校验都应该改写成 `panic`。对于这里这么简单的解析逻辑，实际业务直接返回 `error` 就足够。

本例综合了官方博客的“包内部控制流、对外错误返回”思路，并依据 [`os.MkdirTemp`](https://pkg.go.dev/os#MkdirTemp)、[`os.WriteFile`](https://pkg.go.dev/os#WriteFile)、[`os.File.Close`](https://pkg.go.dev/os#File.Close) 和 [`io.ReadAll`](https://pkg.go.dev/io#ReadAll) 自行编写了可运行程序。临时目录在退出前删除，程序不需要预先准备输入文件。

```go
// 运行示例：E-12
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type parseAbort struct {
	err error
}

func decode(raw []byte) string {
	if string(raw) == "bad" {
		panic(parseAbort{err: errors.New("bad record")})
	}
	return string(raw)
}

func readKnownFormat(name string) (text string, err error) {
	// 先注册恢复边界，让后注册的资源清理先执行。
	defer func() {
		if value := recover(); value != nil {
			failure, ok := value.(parseAbort)
			if !ok {
				panic(value)
			}
			text = ""
			err = errors.Join(err, failure.err)
			fmt.Println("recovered: parse abort")
		}
	}()

	f, err := os.Open(name)
	if err != nil {
		return "", fmt.Errorf("open input: %w", err)
	}
	defer func() {
		closeErr := f.Close()
		fmt.Println("file closed:", closeErr == nil)
		if closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close input: %w", closeErr))
		}
	}()

	raw, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	return decode(raw), nil
}

func run() (err error) {
	dir, err := os.MkdirTemp("", "go-notes-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(dir))
	}()

	name := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(name, []byte("bad"), 0600); err != nil {
		return err
	}

	text, readErr := readKnownFormat(name)
	fmt.Printf("text=%q\n", text)
	return readErr
}

func main() {
	fmt.Printf("error=%v\n", run())
}
```

输出：

```text
file closed: true
recovered: parse abort
text=""
error=bad record
```

执行过程要沿函数边界阅读：

1. `run` 建立临时目录并登记删除动作，写入演示文件。
2. `readKnownFormat` 先登记恢复函数；只有 `os.Open` 成功以后，才登记文件关闭函数。
3. `decode(raw)` 在返回表达式求值时触发 `panic`，该表达式没有正常完成，调用栈开始展开。
4. 文件关闭函数后登记，先执行；若关闭失败，它会保存在具名结果 `err` 中。
5. 恢复函数直接调用 `recover`，识别本例约定的私有中断类型，清空结果并合并错误。它不会把任意未知故障都吞掉。
6. `readKnownFormat` 返回给 `run`，`run` 把错误继续返回；删除临时目录的动作完成后，`main` 才打印最终错误。

即使恢复成功，文件也需要自己的关闭动作；即使无需恢复，文件关闭动作也应独立存在。`errors.Join` 在这个结构中还避免了一个容易漏掉的问题：恢复函数设置解析错误时，不会覆盖先前关闭函数记录的错误。

**四、不要让循环意外扩大资源的生命周期**

`defer` 属于当前函数，不属于最近的 `if` 或 `for` 代码块。如果在一个处理大量文件的循环中持续 `defer f.Close()`，文件可能一直保持打开，直到整个函数结束。资源并非永远不释放，但过长的持有时间也可能耗尽文件描述符。

一种直接的改法是把每次处理放到辅助函数里，让每次调用返回时就完成关闭。下面只是函数体片段，假定 `names` 已定义并使用上面的 `readKnownFormat`，不能单独运行：

```go
for _, name := range names {
    if _, err := readKnownFormat(name); err != nil {
        return err
    }
}
```

此时每次 `readKnownFormat` 结束，自己的文件关闭动作就已经运行完，不会等到整个循环结束。释放时机仍须根据资源是否要跨多次操作共享来安排，不应机械地把所有清理都放到最外层。

还要记住 [`os.Exit`](https://pkg.go.dev/os#Exit) 会立即结束进程，不执行已经登记的 `defer`；调用 `os.Exit` 的 `log.Fatal` 也具有这一点。若某个函数需要先完成清理，应通过正常返回路径把错误交给外层决定退出。官方说明见 [`log.Fatal`](https://pkg.go.dev/log#Fatal)。

<a id="section-5-3-6"></a>

#### 5.3.6 总结

本章可以用下面的排查表回顾，遇到输出与预期不同时按顺序确认：

| 问题 | 要检查的具体位置 |
|---|---|
| 为什么函数内修改了，函数外没变 | 赋值左边是本地参数 `p`，还是目标 `*p` |
| 为什么变量能调用方法，却不能赋给接口 | 看 `T` 与 `*T` 的方法集，别把自动取地址当成接口转换 |
| 为什么值接收者仍改变了外部元素 | 看副本中的切片、映射或指针是否引用同一对象 |
| 为什么 `recover` 得到 `nil` | 是否正在 panic；是否由延迟函数直接调用；是否同一 goroutine |
| 为什么恢复后没有继续执行故障后的语句 | 被展开的调用状态已经放弃；恢复边界会返回其调用者 |
| 为什么有错误文字却无法识别底层错误 | 是否用 `%w` 保留关系；是否用 `errors.Is` 或适当的类型匹配 API |
| 为什么看起来是空错误却不等于 `nil` | `error` 中是否装入了带动态类型的空指针 |
| 为什么文件仍然占用 | 是否在成功取得资源后登记关闭；defer 所属函数是否一直没返回 |

实际写一个资源操作函数时，先定义成功和失败的返回约定，再安排资源取得与释放的位置，最后判断是否需要额外的恢复边界。大多数普通错误沿 `error` 返回就足够；如果确实恢复了 `panic`，必须明确让调用者知道本次操作的结果，而不是默认视为成功。

<a id="chapter-6"></a>


*图示：defer 的登记、执行与返回顺序。*

## 六、defer 关键字

`defer` 的作用是：当前函数即将返回时，调用事先登记的函数。函数正常走到结尾、执行 `return`，或者因为本 goroutine 的 `panic` 而开始退出，都会涉及这些已登记的调用。理解它需要同时观察两个时刻：**执行到 defer 语句时求值并登记，退出当前函数时实际调用**。它不是定时器，也不会另起一个 goroutine。[官方依据：Defer statements](https://go.dev/ref/spec#Defer_statements)。

官方博客 [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) 用三个重点组织讲解：参数的求值时机、后进先出、具名结果参数可被延迟函数修改。本章沿用这三个观察角度，并扩展为图片要求的六节。

<a id="section-6-1"></a>

### 6.1 defer 执行顺序：先进后出

<a id="section-6-1-1"></a>

#### 6.1.1 登记顺序与调用顺序相反

“先进后出”也可表述为“后登记的先执行”，即 LIFO。它适合处理成对资源：先取得外层资源，再取得内层资源，离开时先释放内层资源，再释放外层资源。

**示例 D-01：循环和普通代码块中的 defer。** 改编自官方博客的循环延迟打印示例，增加一层代码块，观察退出代码块与退出函数的区别。

```go
// 运行示例：D-01
package main

import "fmt"

func work() {
	defer fmt.Println("first registered")
	for i := 0; i < 3; i++ {
		defer fmt.Println("loop", i)
	}
	{
		defer fmt.Println("block")
		fmt.Println("inside block")
	}
	fmt.Println("body end")
}

func main() {
	work()
	fmt.Println("caller continues")
}
```

输出：

```text
inside block
body end
block
loop 2
loop 1
loop 0
first registered
caller continues
```

登记顺序是 `first registered`、`loop 0`、`loop 1`、`loop 2`、`block`。这些调用都归属于 `work`，离开内层花括号或结束循环不会立即执行它们；`work` 走到函数末尾才按相反顺序调用。`main` 必须等这些调用完成，才能打印最后一行。

循环中的 `i` 作为 `Println` 的实参，登记时已被读取。因此本例的 `2、1、0` 不依赖 Go 1.22 对循环闭包捕获语义的修改；这里没有闭包在退出时读取同一个循环变量。

<a id="section-6-1-2"></a>

#### 6.1.2 实参快照与闭包捕获是两件事

**示例 D-02：同一个变量的三种观察方式。** 在官方博客“参数立即求值”的示例上增加闭包与显式传参作对照。

```go
// 运行示例：D-02
package main

import "fmt"

func main() {
	n := 1
	defer fmt.Println("argument", n)
	defer func() {
		fmt.Println("closure", n)
	}()
	defer func(snapshot int) {
		fmt.Println("parameter", snapshot)
	}(n)
	n = 8
	fmt.Println("body", n)
}
```

输出：

```text
body 8
parameter 1
closure 8
argument 1
```

| 写法 | 登记时保存的内容 | 延迟调用时读取什么 |
| --- | --- | --- |
| `defer fmt.Println("argument", n)` | 字符串与当时的整数 `1` | 已保存的实参 |
| `defer func() { ... n ... }()` | 一个捕获变量 `n` 的函数值 | 共享变量当时的值 `8` |
| `defer func(snapshot int) { ... }(n)` | 函数值及整数实参 `1` | 形参 `snapshot` 的值 `1` |

闭包捕获的变量仍可在函数体中变化；把一个整数作为参数传入，会复制该整数。若参数本身是指针或切片，复制的则是指针值或切片值，共享对象依然可能变化，不能把“保存参数”推广成“深拷贝整份数据”。依据：[函数字面量与闭包](https://go.dev/ref/spec#Function_literals)、[调用](https://go.dev/ref/spec#Calls)。

<a id="section-6-1-3"></a>

#### 6.1.3 函数值本身也会被保存

**示例 D-03：修改函数变量，不会改掉先前登记的调用。** 根据规范关于 function value 的求值规则编写。

```go
// 运行示例：D-03
package main

import "fmt"

func main() {
	f := func() { fmt.Println("old function") }
	defer f()
	f = func() { fmt.Println("new function") }
	f()
}
```

输出：

```text
new function
old function
```

执行 `defer f()` 时，当前函数值已经保存下来。后续给变量 `f` 重新赋值，只影响后来的 `f()` 调用。若写成 `defer func() { f() }()`，则延迟调用的闭包会在退出时再读取 `f`；那是另一种行为。

**实际使用建议：** 每次成功取得资源后，马上登记相应清理操作，能让“取得”和“释放”的对应关系在代码中靠得更近。若必须在每轮循环结束就释放资源，可以将单轮工作提取成函数；循环内的 `defer` 不会自动以一轮为生命周期。[官方示例：Effective Go — Defer](https://go.dev/doc/effective_go#defer)。

<a id="section-6-2"></a>

### 6.2 在函数里面，defer 和 return 的执行

<a id="section-6-2-1"></a>

#### 6.2.1 将 return 分成可观察的步骤

对带返回表达式的普通 `return`，按下面的次序推导即可：

| 阶段 | 操作 | 是否已经回到调用方 |
| --- | --- | --- |
| 1 | 计算 `return` 后的表达式，包括其中的函数调用 | 否 |
| 2 | 用计算结果设置函数的结果参数 | 否 |
| 3 | 按登记的逆序执行当前函数的延迟调用 | 否 |
| 4 | 延迟调用正常完成后，将最终结果交给调用方 | 是 |

这里的“结果参数”既包括源码中有名字的结果参数，也包括未命名结果对应的返回值。规范不要求程序员掌握某种固定的底层“返回槽”布局，以上是语言层面可依赖的顺序。[官方依据：Return statements](https://go.dev/ref/spec#Return_statements)、[Defer statements](https://go.dev/ref/spec#Defer_statements)。

**示例 D-04：修改普通局部变量，不会重算已求值的整数返回值。**

```go
// 运行示例：D-04
package main

import "fmt"

func evaluate(v int) int {
	fmt.Println("evaluate return", v)
	return v
}

func score() int {
	n := 5
	defer func() {
		n = 20
		fmt.Println("defer local", n)
	}()
	fmt.Println("before return", n)
	return evaluate(n)
}

func main() {
	fmt.Println("caller receives", score())
}
```

输出：

```text
before return 5
evaluate return 5
defer local 20
caller receives 5
```

`evaluate(n)` 先产生整数 `5`，随后该值被用作结果；延迟函数修改的是局部变量 `n`。已形成的整数结果不会因为 `n` 改变而再求值一次。调用方的 `fmt.Println` 要等 `score()` 全部完成，最后才打印。

<a id="section-6-2-2"></a>

#### 6.2.2 “返回值已确定”不等于共享对象不会变化

**示例 D-05：返回指针或切片时，指向的对象仍可被延迟函数修改。** 根据普通赋值和 defer 规则编写。

```go
// 运行示例：D-05
package main

import "fmt"

func number() *int {
	n := 5
	defer func() { n = 20 }()
	return &n
}

func values() []int {
	s := []int{1, 2}
	defer func() { s[0] = 9 }()
	return s
}

func main() {
	fmt.Println(*number())
	fmt.Println(values())
}
```

输出：

```text
20
[9 2]
```

`number` 返回的指针值已经确定，但它指向的 `n` 在真正返回前被改为 `20`。`values` 返回的切片值已经复制，但它与局部切片仍共享底层数组，因而调用方读到更新后的元素。

这个例子没有反驳 D-04：两者返回的值都已按顺序形成，区别在于整数直接存放数值，指针或切片还关联着另一个可变对象。返回局部变量的地址在 Go 中合法，其生命周期由实现按可达性管理，不会因为函数返回就自动变成悬空指针。[官方依据：FAQ — stack or heap](https://go.dev/doc/faq#stack_or_heap)、[值的表示](https://go.dev/ref/spec#Representation_of_values)。

<a id="section-6-3"></a>

### 6.3 具名返回值遇见 defer

本章的“有名函数返回值”在本节按“具名返回值”解释，即 `func f() (n int)` 中的结果参数 `n`。它是函数内可以读写的变量，进入函数时初始化为相应类型的零值。显式 `return 表达式` 仍然会先为它赋值，随后才调用 defer。[官方依据：Return statements](https://go.dev/ref/spec#Return_statements)。

<a id="section-6-3-1"></a>

#### 6.3.1 显式 return 也允许 defer 修改具名结果

**示例 D-06：先设置结果，再由两个 defer 依次修改。** 改编自规范的具名结果修改示例，加入两个延迟步骤及可观察输出。

```go
// 运行示例：D-06
package main

import "fmt"

func calculate() (result int) {
	defer func() {
		result *= 2
		fmt.Println("multiply", result)
	}()
	defer func() {
		result++
		fmt.Println("increment", result)
	}()
	result = 100
	fmt.Println("body", result)
	return 3
}

func main() {
	fmt.Println("final", calculate())
}
```

输出：

```text
body 100
increment 4
multiply 8
final 8
```

执行过程：`result` 先是零值 `0`；函数体改为 `100`；`return 3` 又把结果参数设为 `3`；后登记的 defer 把它加一得到 `4`；先登记的 defer 再乘二得到 `8`；调用方最终收到 `8`。因此不要把最后一行源码中的 `return 3` 当成不可再改变的最终结果。

<a id="section-6-3-2"></a>

#### 6.3.2 修改具名结果，必须确实访问那个变量

**示例 D-07：形参副本、闭包变量、指针参数的区别。**

```go
// 运行示例：D-07
package main

import "fmt"

func byValue() (n int) {
	defer func(v int) { v++ }(n)
	return 7
}

func byClosure() (n int) {
	defer func() { n++ }()
	return 7
}

func byPointer() (n int) {
	defer func(p *int) { (*p)++ }(&n)
	return 7
}

func main() {
	fmt.Println(byValue(), byClosure(), byPointer())
}
```

输出：

```text
7 8 8
```

`byValue` 登记时把当时为 `0` 的 `n` 复制给延迟调用的实参，之后 `v++` 只改形参副本；`byClosure` 捕获结果参数；`byPointer` 保存结果参数的地址。后两者在 `return 7` 设置结果后，修改同一个结果变量，因而返回 `8`。

还需注意三个边界：

- 延迟函数自己的返回值会被丢弃，不能仅通过“在 defer 中 return 一个值”替代外层结果。
- 内层声明一个同名变量可能遮蔽具名结果。不要看到相同拼写就认定是同一个变量；长函数中优先写出清晰的 `return value, err`。
- 具名 `err` 可用于把 `Close` 失败并入函数结果，但要决定是保留原错误、添加上下文还是合并多个错误；不能无条件用清理错误覆盖业务错误。第五章资源清理示例展示了相关写法。

这些机制用于完善退出流程很有价值；仅为让代码更短而堆叠多个修改结果的 defer，会使返回值难以推导。以上整数运算示例用于看清语言规则，实际业务代码通常应将核心计算写在正常执行路径中。

<a id="section-6-4"></a>

### 6.4 defer 遇见 panic 会发生什么

<a id="section-6-4-1"></a>

#### 6.4.1 从当前函数向调用方逐层执行

`panic` 开始之后，当前函数的普通语句不再往下执行。该函数已登记的 defer 按逆序运行；若没有恢复，沿同一 goroutine 的调用链继续向外展开。遇到符合条件的 `recover` 后，恢复边界所在函数结束并返回其调用方；不会跳回触发 panic 的位置。[官方依据：Handling panics](https://go.dev/ref/spec#Handling_panics)。

**示例 D-08：恢复前后，哪些语句还能执行？** 按官方博客的分层调用与恢复流程重新编写，保留两个调用层级以便观察。

```go
// 运行示例：D-08
package main

import "fmt"

func inner() {
	defer fmt.Println("inner cleanup")
	fmt.Println("inner body")
	panic("bad input")
}

func outer() {
	defer fmt.Println("outer cleanup")
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover", r)
		}
	}()
	inner()
	fmt.Println("unreachable outer body")
}

func main() {
	outer()
	fmt.Println("caller continues")
}
```

输出：

```text
inner body
inner cleanup
recover bad input
outer cleanup
caller continues
```

| 步骤 | 发生的事情 | 后果 |
| --- | --- | --- |
| 1 | `inner` 触发 panic | 普通执行中断 |
| 2 | `inner` 的 defer 运行 | 本层先清理 |
| 3 | 展开到 `outer`，先调用后登记的恢复函数 | `recover` 直接取得 panic 值 |
| 4 | 恢复函数正常结束，继续执行 `outer` 更早登记的 defer | 打印 `outer cleanup` |
| 5 | `outer` 返回到 `main` | `main` 可继续执行 |

`unreachable outer body` 不会出现：`inner()` 的调用没有以普通返回的方式完成，恢复操作也不会重新进入 `outer` 的普通函数体。编写恢复边界时，应明确是否返回一个错误结果、记录上下文或者重新 panic，而非仅观察到程序没退出就认定工作成功。

<a id="section-6-4-2"></a>

#### 6.4.2 不应把 defer 理解为任何情况下都保证执行

已登记的 defer 会参与普通返回与本 goroutine 的 panic 展开，但它无法保证覆盖整个程序的所有终止方式。`os.Exit` 会立即结束进程，文档明确说明不调用 defer；`log.Fatal` 最终调用 `os.Exit(1)`，也有相同问题。此外，`main` 返回后程序不会等待其他 goroutine 完成，因此不能靠主函数的退出让其他 goroutine 的清理逻辑有机会执行。[官方依据：os.Exit](https://pkg.go.dev/os#Exit)、[log.Fatal](https://pkg.go.dev/log#Fatal)、[Program execution](https://go.dev/ref/spec#Program_execution)。

资源管理建议是：有清理责任的函数用返回值向上传递错误，完成自身 defer；最外层根据最终错误决定程序退出策略。recover 的范围、直接调用条件和普通 `error` 的区别见第五章。

<a id="section-6-5"></a>

### 6.5 defer 中包含 panic

<a id="section-6-5-1"></a>

#### 6.5.1 延迟函数也是函数，也可能失败

延迟调用内部可以执行任意普通函数逻辑，其中也可能触发新的 panic。例如清理时解引用了 nil 指针、关闭已关闭的 channel，或者显式调用 `panic`。因此“正在执行清理”并不意味着没有失败路径。

**示例 D-09：清理阶段的新 panic 会影响外层 recover 看到的值。** 这是根据规范的栈展开规则构造的对照例，不建议在实际清理函数中故意使用此结构。

```go
// 运行示例：D-09
package main

import "fmt"

func job() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered", r)
		}
	}()
	defer fmt.Println("earlier cleanup still runs")
	defer func() {
		fmt.Println("cleanup starts")
		panic("cleanup failure")
	}()
	panic("original failure")
}

func main() {
	job()
	fmt.Println("caller continues")
}
```

输出：

```text
cleanup starts
earlier cleanup still runs
recovered cleanup failure
caller continues
```

原始 panic 先触发最后登记的清理函数；该清理函数又触发一个 panic。其他已登记的 defer 仍会参与展开。本例最外层的 `recover` 观察到的是 `cleanup failure`，两个 panic 值不会自动组合成 `errors.Join` 那样的错误树。若没有恢复，运行时的最终诊断可能展示多个 panic，不能据此把程序内 `recover` 的结果理解为一个完整的失败历史。

写清理代码时，优先把可预期的清理失败作为 `error` 返回或记录，并保留主要失败的上下文。是否在一个明确边界恢复特定 panic，要由接口约定决定；把所有清理失败都转成新的 panic 往往会让最初的问题更难定位。语义依据：[Handling panics](https://go.dev/ref/spec#Handling_panics)；多错误合并接口：[errors.Join](https://pkg.go.dev/errors#Join)。

<a id="section-6-5-2"></a>

#### 6.5.2 nil 函数值的 panic 发生在调用时

**示例 D-10：登记 nil 函数值，函数体仍会先继续运行。** 规范明确规定：如果延迟的函数值是 nil，panic 在实际调用它时发生。

```go
// 运行示例：D-10
package main

import "fmt"

func main() {
	defer func() {
		fmt.Println("recovered", recover() != nil)
	}()
	var f func()
	defer f()
	f = func() { fmt.Println("replacement") }
	fmt.Println("body finished")
}
```

输出：

```text
body finished
recovered true
```

登记时保存的是 nil 函数值；随后给变量 `f` 赋值不能修复已登记的调用，所以 `replacement` 不会输出。这个例子把 D-03 的“函数值会保存”与 panic 的“调用阶段发生”结合起来。程序只比较恢复值是否非 nil，避免依赖不同平台的运行时错误文字。[官方依据：Defer statements](https://go.dev/ref/spec#Defer_statements)。

<a id="section-6-6"></a>

### 6.6 defer 下的函数参数包含子函数

<a id="section-6-6-1"></a>

#### 6.6.1 先计算子函数，再登记外层调用

看到 `defer report(makeValue())`，要先区分两次调用：`makeValue()` 负责产生实参，立即执行；`report(...)` 是被延迟的那一次调用，稍后执行。参数里的函数调用并不会因为外层有 `defer` 就一起被推迟。

**示例 D-11：两个参数子函数的执行时机。** 根据规范的 defer 求值规则和通常的函数调用求值顺序编写。

```go
// 运行示例：D-11
package main

import "fmt"

func value(label string, n int) int {
	fmt.Println("evaluate", label, n)
	return n
}

func report(a, b int) {
	fmt.Println("report", a, b)
}

func main() {
	defer report(value("A", 2), value("B", 3))
	fmt.Println("body")
}
```

输出：

```text
evaluate A 2
evaluate B 3
body
report 2 3
```

两个参数中的函数调用依源码顺序执行，得到 `2` 和 `3`；再保存 `report` 与这两个实参；最后才继续普通函数体。不要将这个例子推广为“任意复杂表达式的所有读操作都严格从左到右”，更一般的求值次序应直接查看 [Order of evaluation](https://go.dev/ref/spec#Order_of_evaluation)。

<a id="section-6-6-2"></a>

#### 6.6.2 返回函数的函数：两对括号分别在何时执行

**示例 D-12：开始动作现在做，收尾动作稍后做。** 改编自 [Effective Go 的 trace/untrace 思路](https://go.dev/doc/effective_go#defer)，使用函数返回值组织成对行为。

```go
// 运行示例：D-12
package main

import "fmt"

func begin(name string) func() {
	fmt.Println("begin", name)
	return func() {
		fmt.Println("end", name)
	}
}

func main() {
	defer begin("job")()
	fmt.Println("body")
}
```

输出：

```text
begin job
body
end job
```

`begin("job")` 先运行并返回一个函数值，末尾的 `()` 表示调用这个返回的函数，而这一调用被 defer 延迟。可以将其理解为以下等价片段，片段需放在函数中，不能单独运行：

```go
finish := begin("job")
defer finish()
```

对比 `defer func() { begin("job")() }()`：它先保存一个匿名函数；真正退出时才执行匿名函数体，所以 begin 和 end 都发生在退出阶段。判断括号层级比背诵“defer 后的都最后执行”更可靠。

<a id="section-6-6-3"></a>

#### 6.6.3 实参计算时 panic，该次调用尚未登记完成

**示例 D-13：后面的清理语句不等于已经生效。**

```go
// 运行示例：D-13
package main

import "fmt"

func argument() int {
	fmt.Println("evaluate argument")
	panic("argument failed")
}

func work() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered", r)
		}
	}()
	defer fmt.Println("registered cleanup")
	defer fmt.Println("not registered", argument())
}

func main() {
	work()
}
```

输出：

```text
evaluate argument
registered cleanup
recovered argument failed
```

计算最后一个 defer 的实参时，`argument()` 已经 panic，外层 `Println` 未能完成登记。因此不会出现 `not registered`。此前登记成功的两个调用仍按逆序执行。这说明资源清理要考虑**是否已经执行到登记语句并完成其求值**，不能只看源码里是否写了一个 defer。

<a id="section-6-6-4"></a>

#### 6.6.4 方法接收者也需要在登记时确定

**示例 D-14：值接收者保留副本，指针接收者保留指向对象的指针。** 根据规范的方法调用和 defer 规则编写。

```go
// 运行示例：D-14
package main

import "fmt"

type Counter struct {
	n int
}

func (c Counter) Snapshot() {
	fmt.Println("value receiver", c.n)
}

func (c *Counter) Latest() {
	fmt.Println("pointer receiver", c.n)
}

func main() {
	c := Counter{n: 1}
	defer c.Snapshot()
	defer c.Latest()
	c.n = 9
	fmt.Println("body", c.n)
}
```

输出：

```text
body 9
pointer receiver 9
value receiver 1
```

登记 `Snapshot` 时复制了结构体，副本中的 `n` 为 `1`；登记 `Latest` 时保存了 `&c`，实际执行时通过这个指针读取已改为 `9` 的对象。两次调用仍遵循 LIFO。若结构体含有切片、map 或指针字段，值接收者的副本也可能共享字段关联的数据，所以这里的“副本”仍不是通用深拷贝。[官方依据：Method values](https://go.dev/ref/spec#Method_values)、[Calls](https://go.dev/ref/spec#Calls)。

<a id="section-6-6-5"></a>

#### 6.6.5 本章速查

| 需要判断的情况 | 推导规则 |
| --- | --- |
| 多个 defer | 已完成登记的调用按逆序执行 |
| defer 在代码块或循环内 | 等当前函数退出，不在普通花括号结束时执行 |
| defer 的普通实参 | 执行到 defer 时求值并保存；共享对象仍可能变化 |
| defer 的匿名函数体 | 实际调用时执行；捕获变量与显式参数要分别看待 |
| return 与 defer | 先计算并设置结果，再执行 defer，最后返回调用方 |
| 具名结果参数 | 延迟函数若访问同一变量，就能修改最终结果 |
| panic 与 defer | 同 goroutine 中沿调用链展开，并执行已登记的 defer |
| recover 后的执行位置 | 恢复边界函数返回调用方，不回到 panic 位置 |
| defer 的参数本身会 panic | 该次登记可能未完成；此前成功登记的仍会展开 |
| `defer factory()()` | factory 现在执行，返回的函数稍后调用 |

本章内容对应 [Go 规范的 Defer statements](https://go.dev/ref/spec#Defer_statements)、[Return statements](https://go.dev/ref/spec#Return_statements)、[Handling panics](https://go.dev/ref/spec#Handling_panics)。先按顺序推演，再运行程序核对，能比单独记忆输出更稳定地掌握这些规则。

<a id="appendix-a"></a>

## 附录 A：综合示例——读取文件并统计非负整数

<a id="section-a-1"></a>

### A.1 将函数、指针方法、error 和 defer 放到同一流程

下面的程序创建两份临时文本，逐行读取非负整数，返回数量与总和。如果遇到负数，返回已经统计的部分结果和带行号的错误；调用方用 `errors.Is` 识别错误类别。文件在读取函数结束时关闭，临时目录在演示函数结束时清理。

这个例子的结构由本笔记组织，逐行读取的方式参考 [bufio.Scanner 的 Lines 示例](https://pkg.go.dev/bufio#example-Scanner-Lines)，错误处理依据 [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)、[fmt.Errorf](https://pkg.go.dev/fmt#Errorf) 和 [errors.Join](https://pkg.go.dev/errors#Join)。程序只使用标准库，无须准备外部文件。

```go
// 运行示例：I-01
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrNegative = errors.New("negative value")

type Stats struct {
	Count int
	Sum   int
}

func (s *Stats) Add(n int) {
	s.Count++
	s.Sum += n
}

func parseValue(text string) (int, error) {
	n, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", text, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("%w: %d", ErrNegative, n)
	}
	return n, nil
}

// readStats 返回有效前缀的统计结果；遇到无效行时同时返回 error。
func readStats(path string) (result Stats, err error) {
	f, err := os.Open(path)
	if err != nil {
		return result, fmt.Errorf("open input: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close input: %w", closeErr))
		}
	}()

	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		n, parseErr := parseValue(text)
		if parseErr != nil {
			return result, fmt.Errorf("line %d: %w", line, parseErr)
		}
		result.Add(n)
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return result, fmt.Errorf("scan input: %w", scanErr)
	}
	return result, nil
}

func run() (err error) {
	dir, err := os.MkdirTemp("", "go-ch04-")
	if err != nil {
		return err
	}
	defer func() {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			err = errors.Join(err, fmt.Errorf("remove demo directory: %w", removeErr))
		}
	}()

	cases := []struct {
		name string
		data string
	}{
		{"good", "4\n5\n6\n"},
		{"bad", "4\n-2\n"},
	}
	for _, item := range cases {
		path := filepath.Join(dir, item.name+".txt")
		if writeErr := os.WriteFile(path, []byte(item.data), 0600); writeErr != nil {
			return writeErr
		}
		stats, readErr := readStats(path)
		fmt.Printf("%s: count=%d sum=%d error=%t\n",
			item.name, stats.Count, stats.Sum, readErr != nil)
		fmt.Println("negative:", errors.Is(readErr, ErrNegative))
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println("demo setup or cleanup failed:", err)
	}
}
```

在正常可读写的临时目录中，输出：

```text
good: count=3 sum=15 error=false
negative: false
bad: count=1 sum=4 error=true
negative: true
```

<a id="section-a-2"></a>

### A.2 沿着一次失败调用阅读程序

以 `bad` 输入为例：第一行 `4` 转换成功，`result.Add(4)` 使用指针接收者更新当前统计；第二行 `-2` 转换为整数后被业务规则拒绝，返回包装了 `ErrNegative` 的错误；`readStats` 再包装行号，然后返回 `result` 和包装后的错误。

此时结果参数已经得到“数量 1、总和 4”和带上下文的错误。关闭文件的 defer 随后执行；如果关闭也失败，就把两类错误合并。最后调用方获得部分统计与错误，`errors.Is` 仍能穿过包装找到 `ErrNegative`。错误发生在第二行，并不意味着第一行的统计自动回滚；“返回有效前缀”已经写进函数注释，是这个接口的约定。

| 本讲知识点 | 综合例子中的位置 |
| --- | --- |
| 多返回值 | `parseValue` 返回数值与错误，`readStats` 返回统计与错误 |
| 指针接收者 | `(*Stats).Add` 修改调用方的统计对象 |
| 具名结果与闭包 | defer 通过闭包访问 `err`，把关闭错误加入结果 |
| 获取成功后登记清理 | `os.Open` 成功后登记 `Close`，临时目录创建成功后登记删除 |
| 错误上下文 | `%w` 保存原因，同时附加操作名称或行号 |
| 错误分类 | `errors.Is(readErr, ErrNegative)` 不依赖完整错误字符串 |
| 循环中的资源生命周期 | 每次 `readStats` 调用结束都先关闭文件，再进入下一轮 |

`scanner.Scan()` 返回 false 既可能表示输入结束，也可能表示读取失败，因此循环后必须检查 `scanner.Err()`；`Scanner` 默认按行划分，并有 token 长度上限。例子只处理少量小整数，扩展到大型输入时，还需要为单行大小和总和溢出制定明确规则。[官方依据：Scanner.Scan](https://pkg.go.dev/bufio#Scanner.Scan)、[Scanner.Buffer](https://pkg.go.dev/bufio#Scanner.Buffer)、[整数溢出](https://go.dev/ref/spec#Integer_overflow)。

<a id="appendix-b"></a>

## 附录 B：练习与参考答案

<a id="section-b-1"></a>

### B.1 练习题

先写出原因，再运行相关示例核对，避免只记忆某一个输出。

1. `func change(p *int) { p = new(int); *p = 99 }` 接收 `&n` 后，能否把调用方的 `n` 改成 `99`？怎样修改才能做到？
2. `new(map[string]int)` 与 `make(map[string]int)` 的返回类型分别是什么？前者为什么不能直接向所指 map 中写入条目？
3. 把同一个闭包返回的函数赋给两个变量，会不会产生两份独立计数器？调用两次“创建闭包”的函数又会怎样？
4. 有方法 `func (c *Counter) Add(int)`，为什么变量 `c` 可以写 `c.Add(1)`，却不表示值 `c` 一定满足要求 `Add(int)` 的接口？
5. 在 D-02 中，把 `defer func() { fmt.Println("closure", n) }()` 改成带 `n` 参数的匿名函数，会改变哪一行输出？
6. 在 D-06 中，将最后的 `return 3` 改成裸 `return`，最终返回多少？请写出两个 defer 执行后的中间值。
7. 某函数通过 defer 恢复了它调用的子函数产生的 panic，能否继续执行该子函数调用后面的普通语句？
8. `var p *MyError = nil; var err error = p` 中，`err == nil` 的结果是什么？实现 `Error()` 方法时检查 `p == nil` 能否改变这一比较结果？
9. 循环打开很多文件，每次都在循环体直接写 `defer f.Close()`，为什么可能同时保留很多打开的文件？如何调整代码结构？
10. 在综合示例中，把 `%w` 改成 `%v` 只保留字符串后，为什么 `errors.Is` 可能无法识别 `ErrNegative`？如果关闭文件也失败，应如何保留两种错误？

<a id="section-b-2"></a>

### B.2 参考答案与推导

| 题号 | 参考答案 | 对应规则 |
| --- | --- | --- |
| 1 | 不能。`p` 是指针副本，重新赋值后改动的是新变量。要改 `n`，对传入的非 nil 指针执行 `*p = 99`，不要先重新绑定它 | 第四章、5.1；[Calls](https://go.dev/ref/spec#Calls) |
| 2 | 分别为 `*map[string]int` 与 `map[string]int`。前者指向的 map 仍为 nil，需先给 `*p` 赋一个初始化的 map；通常直接使用 make 更清楚 | 3.1—3.3；[Allocation](https://go.dev/ref/spec#Allocation) |
| 3 | 复制同一个函数值不会复制其捕获的状态；分别调用创建函数则可得到分别捕获不同局部变量的闭包 | 2.5；[Function literals](https://go.dev/ref/spec#Function_literals) |
| 4 | 可寻址变量上的调用允许按规则自动取地址，而接口满足关系依据类型的方法集；`Counter` 与 `*Counter` 的方法集不同 | 5.2；[Method sets](https://go.dev/ref/spec#Method_sets) |
| 5 | 原来的 `closure 8` 变为 `closure 1`，因为登记时复制了整数参数；LIFO 顺序不变 | 6.1；[Defer statements](https://go.dev/ref/spec#Defer_statements) |
| 6 | 裸 return 使用当前结果参数 `100`，后登记的 defer 加一为 `101`，先登记的 defer 乘二为 `202`，最后返回 `202` | 6.3；[Return statements](https://go.dev/ref/spec#Return_statements) |
| 7 | 不能。恢复后，该恢复边界函数的剩余 defer 执行，然后函数返回调用方；不会回到中断的普通语句 | 5.3.1、6.4；[Handling panics](https://go.dev/ref/spec#Handling_panics) |
| 8 | 为 false。接口包含动态类型 `*MyError`；`Error()` 的实现只影响调用方法的行为，不会把已装入接口的类型化 nil 变成 nil 接口 | 5.3.3；[FAQ — nil error](https://go.dev/doc/faq#nil_error) |
| 9 | defer 绑定到当前函数，而非每次循环。将单文件操作提取为函数，或者在明确时点显式关闭并处理错误，可以缩短资源持有时间 | 5.3.5、6.1；[Effective Go — Defer](https://go.dev/doc/effective_go#defer) |
| 10 | `%v` 仅形成文本，不建立这个包装关系。用 `%w` 保留原错误，并在确有关闭错误时使用 `errors.Join` 等策略保留两者 | 5.3.3、附录 A；[errors](https://pkg.go.dev/errors) |

<a id="appendix-c"></a>

## 附录 C：官方文档与后续阅读

<a id="section-c-1"></a>

### C.1 官方资料按用途索引

| 想核对的内容 | 官方入口 |
| --- | --- |
| 函数签名、声明、多返回值 | [Function types](https://go.dev/ref/spec#Function_types)、[Function declarations](https://go.dev/ref/spec#Function_declarations)、[Return statements](https://go.dev/ref/spec#Return_statements) |
| 匿名函数与闭包 | [Function literals](https://go.dev/ref/spec#Function_literals) |
| 调用与可变参数 | [Calls](https://go.dev/ref/spec#Calls)、[Passing arguments to ... parameters](https://go.dev/ref/spec#Passing_arguments_to_..._parameters) |
| make、new 与零值 | [Making slices, maps and channels](https://go.dev/ref/spec#Making_slices_maps_and_channels)、[Allocation](https://go.dev/ref/spec#Allocation)、[The zero value](https://go.dev/ref/spec#The_zero_value) |
| Go 1.26 的 new 表达式扩展 | [Go 1.26 Release Notes](https://go.dev/doc/go1.26#language) |
| 指针、取地址、解引用 | [Pointer types](https://go.dev/ref/spec#Pointer_types)、[Address operators](https://go.dev/ref/spec#Address_operators) |
| 方法声明、方法集、方法值 | [Method declarations](https://go.dev/ref/spec#Method_declarations)、[Method sets](https://go.dev/ref/spec#Method_sets)、[Method values](https://go.dev/ref/spec#Method_values) |
| 值传递、nil error、栈与堆 | [Go FAQ](https://go.dev/doc/faq) |
| 惯用函数设计及资源管理 | [Effective Go](https://go.dev/doc/effective_go) |
| defer、panic、recover 的正式语义 | [Defer statements](https://go.dev/ref/spec#Defer_statements)、[Handling panics](https://go.dev/ref/spec#Handling_panics) |
| defer 三条规则与分层恢复示例 | [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)，Andrew Gerrand，2010-08-04 |
| error 的设计与表达 | [Error handling and Go](https://go.dev/blog/error-handling-and-go) |
| 错误包装与检查 | [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)、[errors 标准库](https://pkg.go.dev/errors) |
| Go 1.22 循环变量变化 | [Go 1.22 Release Notes](https://go.dev/doc/go1.22#language) |
| Go 1.21 panic(nil) 变化 | [Go 1.21 Release Notes](https://go.dev/doc/go1.21#language) |

<a id="section-c-2"></a>

### C.2 中文辅助资料

以下保留相关的阅读入口。正文的语言规则和标准库行为以官方资料为核对依据，中文资料可作为换一种表达方式的辅助阅读。其中变量与常量一文主要用于基础知识衔接。

- [C 语言中文网：Go 语言入门教程](https://m.biancheng.net/golang/)
- [Golang 中文学习文档：基本语法](https://golang.halfiisland.com/essential/base/1.grammer.html)
- [李文周：Go 语言基础之变量和常量](https://liwenzhou.com/posts/go/var-and-const/)

<a id="section-c-3"></a>

### C.3 版本和示例验证说明

本次共验证 **55 个完整独立程序**，均编译运行成功，实际标准输出与文中预期输出逐一一致。实际工具链为 `go version go1.27.3 linux/amd64`。全部 55 个程序统一采用 `go 1.27.0` 模块语言设置，`M-06` 用于验证 `new(值表达式)`。

| 示例组 | 内容 | 完整程序数 |
| --- | --- | --- |
| F | 函数基础 | 15 |
| M | make 与 new | 6 |
| P | 指针 | 7 |
| E | 高阶函数机制与错误处理 | 12 |
| D | defer 的执行机制 | 14 |
| I | 文件统计综合例子 | 1 |

验证范围为标有“运行示例”编号的完整程序。语法示意、故意错误和依赖上下文的局部片段不作为独立程序运行。所有完整程序只依赖 Go 标准库；文件例子使用自行创建的临时文件。这些结果用于确认示例在上述环境中的行为，不构成对所有操作系统、历史工具链或异常 I/O 条件的穷举测试。

本书保留版本差异是为了帮助读者理解旧文章和新工具链：Go 1.22 的循环变量语义、Go 1.26 的 `new(值表达式)` 不应直接套用到更早的语言版本。具体特性以所在小节和官方发布说明为准；编译器版本、模块 `go` 指令、标准库可用 API 是相关但不同的概念。

<a id="section-c-4"></a>

### C.4 来源、改编与许可

本笔记的中文组织、执行推导和多数完整程序为围绕该目录编写的讲解内容。对于明确标为“改编”的例子，已在示例附近链接到原始官方页面，并说明增加了完整入口、调整输入、组合对照或打印轨迹等修改；这不表示 Go 项目为本笔记背书。

Go 官方站点的 [Copyright 页面](https://go.dev/copyright) 说明：除另有注明外，站点内容按 [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) 许可，代码按 [Go BSD 许可](https://go.dev/LICENSE) 发布。文档归属其原作者及 Go 项目；本笔记引用规则并重新组织中文讲解。为保留改编官方代码所涉及的许可信息，附 Go 代码许可全文：

```text
Copyright 2009 The Go Authors.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google LLC nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```
