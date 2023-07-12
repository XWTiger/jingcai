package util

import (
	"fmt"
	"time"
)

func Combine(n int, k int) [][]int {
	res := [][]int{}
	// 记录回溯算法的递归路径
	track := []int{}
	backtrack(1, n, k, &res, &track)
	return res
}

func backtrack(start int, n int, k int, res *[][]int, track *[]int) {
	// base case
	if k == len(*track) {
		// 遍历到了第 k 层，收集当前节点的值
		temp := make([]int, len(*track))
		copy(temp, *track)
		*res = append(*res, temp)
		return
	}

	// 回溯算法标准框架
	for i := start; i <= n; i++ {
		// 选择
		*track = append(*track, i)
		// 通过 start 参数控制树枝的遍历，避免产生重复的子集
		backtrack(i+1, n, k, res, track)
		// 撤销选择
		*track = (*track)[:len(*track)-1]
	}
}

func AddTwoHToTime(date time.Time) time.Time {
	return date.Add(time.Hour * 2)
}

func StrToTime(timeStr string) (time.Time, error) {
	date, error := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
	if error != nil {
		fmt.Println(error)
		return AddTwoHToTime(time.Now()), error
	} else {
		return date, nil
	}
}
