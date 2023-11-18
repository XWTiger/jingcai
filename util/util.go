package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 从n里面取出k个组合C(n,k)
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

	dateEnd = fmt.Sprintf("%d-%s-%s 20:55:00", now.Year(), getNum(int(now.Month())), getNum(int(now.Day())))
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

func GetTodayYYHHMMSSFrom(dataTime time.Time) string {
	dateEnd := dataTime.Format("2006-01-02 15:04:05")
	return dateEnd

}

// 获取组合6的所有类型
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

func GetZxGsb(index int, all [][]string, sb *[]byte, childs *[]string) {

	if index == 3 {
		var str = string((*sb)[0 : len((*sb))-1])
		*childs = append(*childs, str)
		return
	}

	for i := 0; i < len(all[0]); i++ {
		fmt.Println(index)
		num := all[index][i]
		var b = []byte(fmt.Sprintf("%s%s", num, " "))
		*sb = append(*sb, b...)
		GetZxGsb(index+1, all, sb, childs)
		*sb = (*sb)[:len(*sb)-2]
	}

}

func GetSpaceStr(num string) []string {
	var arr = make([]string, 0)
	for i := 0; i < len(num); i++ {
		arr = append(arr, fmt.Sprintf("%c", num[i]))
	}
	return arr
}

// 从数组里面抽出A（n,m）
func PermuteAnm(arr []int, m int) [][]int {
	result := [][]int{}
	permHelper(arr, m, []int{}, &result)
	return result
}

func permHelper(arr []int, m int, current []int, result *[][]int) {
	if m == 0 {
		*result = append(*result, current)
		return
	}

	for i := 0; i < len(arr); i++ {
		next := make([]int, len(current)+1)
		copy(next, current)
		next[len(current)] = arr[i]

		remaining := make([]int, len(arr)-1)
		copy(remaining[:i], arr[:i])
		copy(remaining[i:], arr[i+1:])

		permHelper(remaining, m-1, next, result)
	}
}

// 二同排列5
func Get2SamePlW(arr []int, sameNum int) []string {
	result := PermuteAnm(arr, 5)
	var sum = 0

	var turns = make([]string, 0)
	var duplicate = make([][]int, 0)
	for _, ints := range result {
		var exist = false
		for _, value := range duplicate {
			tmp := fmt.Sprintf("%d%d%d%d%d", value[0], value[1], value[2], value[3], value[4])
			tmp2 := fmt.Sprintf("%d%d%d%d%d", ints[0], ints[1], ints[2], ints[3], ints[4])
			//fmt.Println(tmp2)
			if strings.Compare(tmp2, tmp) == 0 {
				exist = true
				break
			}
		}
		if !exist {
			duplicate = append(duplicate, ints)
		}
	}
	fmt.Println(len(duplicate))
	for _, ints := range duplicate {
		var count = 0
		for _, val := range ints {
			if val == sameNum {
				count++
				if count == 2 {
					break
				}
			}

		}
		if count == 2 {

			sum++
			turns = append(turns, fmt.Sprintf("%d %d %d %d %d", ints[0], ints[1], ints[2], ints[3], ints[4]))
		}
	}
	fmt.Println(sum)
	return turns
}
