package web

import (
	"fmt"
	"strings"
)

const (
	continueKeyword string = "继续"
)

func detectContinue(input string) bool {
	return strings.EqualFold(input, continueKeyword)
}

func getContinueHint() string {
	return fmt.Sprintf("处理时间过长，请稍后回复%s获取上一轮对话结果", continueKeyword)
}

func getContinueEmptyHint() string {
	return "上一轮对话结果为空，请重试对话"
}
