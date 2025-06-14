package biz

// 函数式选项模式
// 配置项较多, 且可选

// 闭包函数 - 记住外部变量的值和状态, 状态函数（匿名函数）
// 因为匿名函数使用的参数 是外部参数 - 从而会把外部参数放到堆上, 不会随函数结束而消失（栈）

type ListOptions struct {
	Limit       int64
	Offset      int64
	Tag         string
	Author      string
	FavoritedBy string
	CurrentUid  uint
}

func NewListOptions(opts ...ListOption) *ListOptions {
	// 默认值暂无
	options := &ListOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

type ListOption func(*ListOptions)

func WithLimit(limit int) ListOption {
	return func(o *ListOptions) {
		o.Limit = int64(limit)
	}
}

func WithOffset(offset int) ListOption {
	return func(o *ListOptions) {
		o.Offset = int64(offset)
	}
}

func WithTag(tag string) ListOption {
	return func(o *ListOptions) {
		o.Tag = tag
	}
}

func WithAuthor(author string) ListOption {
	return func(o *ListOptions) {
		o.Author = author
	}
}

func WithFavoritedBy(favoritedBy string) ListOption {
	return func(o *ListOptions) {
		o.FavoritedBy = favoritedBy
	}
}

func WithCurrentUid(currentUid uint) ListOption {
	return func(o *ListOptions) {
		o.CurrentUid = currentUid
	}
}
