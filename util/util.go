package util

import (
	"fmt"
	"strconv"
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
func GetPLWFinishedTime() time.Time {
	now := time.Now()
	var dateEnd string

	dateEnd = fmt.Sprintf("%d-%s-%s 21:20:00", now.Year(), getNum(int(now.Month())), getNum(int(now.Day())))
	time, err := time.ParseInLocation("2006-01-02 15:04:05", dateEnd, time.Local)

	if err != nil {
		fmt.Println(err)
	}

	return time
}

func getNum(num int) string {

	if num < 10 {
		return fmt.Sprintf("0%d", num)
	} else {
		return strconv.Itoa(num)
	}

}

// 09 和 9 对比校验
func PaddingZeroCompare(buyNum string, releaseNum string) bool {
	num, err := strconv.Atoi(buyNum)
	rnum, err := strconv.Atoi(releaseNum)
	if err != nil {
		fmt.Println("号码对比失败!")
		return false
	}
	if num == rnum {
		return true
	} else {
		return false
	}

}

func GetTodayYYHHMMSS() string {
	now := time.Now()
	dateEnd := fmt.Sprintf("%d-%s-%s", now.Year(), getNum(int(now.Month())), getNum(int(now.Day())))
	return dateEnd

}

func Permute(nums []int) [][]int {
	res := make([][]int, 0)
	track := make([]int, 0)
	used := make([]bool, len(nums))
	backtrackA(&res, &track, nums, used)
	return res
}

func backtrackA(res *[][]int, track *[]int, nums []int, used []bool) {
	// base case，到达叶子节点
	if len(*track) == len(nums) {
		// 收集叶子节点上的值
		tmp := make([]int, len(*track))
		copy(tmp, *track)
		*res = append(*res, tmp)
		return
	}

	// 回溯算法标准框架
	for i := 0; i < len(nums); i++ {
		// 已经存在 track 中的元素，不能重复选择
		if used[i] {
			continue
		}
		// 做选择
		used[i] = true
		*track = append(*track, nums[i])
		// 进入下一层回溯树
		backtrackA(res, track, nums, used)
		// 取消选择
		*track = (*track)[:len(*track)-1]
		used[i] = false
	}
}

func GetCombine3(arr []int) [][]int {
	combines := make([][]int, 0)
	if arr[0] == arr[1] {
		var child = []int{arr[2], arr[0], arr[0]}
		var child2 = []int{arr[0], arr[2], arr[0]}
		var child3 = []int{arr[0], arr[0], arr[2]}
		combines = append(combines, child, child2, child3)
	} else {
		if arr[0] == arr[2] {
			var child = []int{arr[1], arr[0], arr[0]}
			var child2 = []int{arr[0], arr[1], arr[0]}
			var child3 = []int{arr[0], arr[0], arr[1]}
			combines = append(combines, child, child2, child3)
		} else {
			var child = []int{arr[2], arr[2], arr[0]}
			var child2 = []int{arr[0], arr[2], arr[2]}
			var child3 = []int{arr[2], arr[0], arr[2]}
			combines = append(combines, child, child2, child3)
		}
	}
	return combines
}

func CovertStrArrToInt(arr []string) []int {
	arry := make([]int, 0)
	for _, s := range arr {
		num, _ := strconv.Atoi(s)
		arry = append(arry, num)
	}
	return arry
}

func GetPaddingId(id uint) string {
	strId := strconv.Itoa(int(id))
	if len(strId) < 6 {
		size := 6 - len(strId)
		var padding = make([]byte, size)
		for i := 0; i < size; i++ {
			padding[i] = '0'
		}
		return fmt.Sprintf("%s%s", string(padding), strId)
	}
	return "000000"
}
