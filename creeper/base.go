package creeper

type Content struct {
	Type       string
	Content    string
	ImageUrl   []string
	Url        string
	Summery    string
	Extra      string //额外的一些信息
	Title      string
	Match      string   //比赛
	Predict    string   //预测谁赢
	Conditions []string //条件 让球 1.25
	time       string
	league     string
}
type Creeper interface {
	Creep() []Content
	Key() string
}
