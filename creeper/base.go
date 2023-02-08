package creeper

type Content struct {
	Type     string
	Content  string
	ImageUrl []string
	Url      string
	Summery  string
	Extra    string //额外的一些信息
	Title    string
}
type Creeper interface {
	Creep() []Content
	Key() string
}
