package service

// 比较答案是否正确
func compareAnswers(correctAnswer string, userAnswer []string) bool {
	// 将正确答案字符串转换为字符数组进行比较
	// 例如："ABC" 转换为 ["A", "B", "C"]
	var correctAnswers []string
	for _, c := range correctAnswer {
		correctAnswers = append(correctAnswers, string(c))
	}

	// 判断数组长度是否相同
	if len(correctAnswers) != len(userAnswer) {
		return false
	}

	// 创建一个 map 来跟踪用户答案中的每个字符
	userAnswerMap := make(map[string]bool)
	for _, a := range userAnswer {
		userAnswerMap[a] = true
	}

	// 检查正确答案中的每个字符是否都在用户答案中
	for _, a := range correctAnswers {
		if !userAnswerMap[a] {
			return false
		}
	}

	return true
}
