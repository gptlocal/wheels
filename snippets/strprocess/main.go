package main

import (
	"fmt"
	"strings"
)

// splitPath 解析输入路径，返回基础路径和特性标识
func splitPath(input string) (string, string) {
	// 查找最后一个 ':' 的位置
	colonIndex := strings.LastIndex(input, ":")

	// 如果没有找到 ':', 返回原路径和空字符串
	if colonIndex == -1 {
		return input, ""
	}

	// 找到最后一个 ':'，截取基础路径和特性部分
	basePath := input[:colonIndex]
	feature := input[colonIndex+1:]

	return basePath + "/", feature
}

func main() {
	// 测试用例
	inputs := []string{
		"/live/service-name/",
		"/live/service-name:feature1/",
		"/live/service-name:hello:feature1/",
	}

	for _, input := range inputs {
		basePath, feature := splitPath(input)
		fmt.Printf("Input: %s\nBase Path: %s\nFeature: %s\n\n", input, basePath, feature)
	}
}
