// MIGRATED: reflect/caller_test.go on 2024-10-01
package main

import (
	"fmt"
	"runtime"
)

func main() {
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		funcInfo := runtime.FuncForPC(pc)
		funcName := funcInfo.Name()
		fmt.Printf("[main] Called from function: %s\nFile: %s\nLine: %d\n", funcName, file, line)
	}

	func1()
}

func func1() {
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		funcInfo := runtime.FuncForPC(pc)
		funcName := funcInfo.Name()
		fmt.Printf("[func1] Called from function: %s\nFile: %s\nLine: %d\n", funcName, file, line)
	}

	func2()
}

func func2() {
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		funcInfo := runtime.FuncForPC(pc)
		funcName := funcInfo.Name()
		fmt.Printf("[func2] Called from function: %s\nFile: %s\nLine: %d\n", funcName, file, line)
	}

	func3()
}

func func3() {
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		funcInfo := runtime.FuncForPC(pc)
		funcName := funcInfo.Name()
		fmt.Printf("[func3] Called from function: %s\nFile: %s\nLine: %d\n", funcName, file, line)
	}
}
