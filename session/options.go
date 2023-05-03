package session

type options struct {
	//默认开启
	notNeedHistory bool

	//默认关闭
	needHook bool
}

// optionFunc统一开关函数
type optionFunc func(*options)

func (s *Session) Options(opts ...optionFunc) {
	for _, opt := range opts {
		opt(&s.opts)
	}
}

func CloseHook() optionFunc {
	return func(o *options) {
		o.needHook = false
	}
}

func OpenHook() optionFunc {
	return func(o *options) {
		o.needHook = true
	}
}

func CloseHistory() optionFunc {
	return func(o *options) {
		o.notNeedHistory = true
	}
}

func OpenHistory() optionFunc {
	return func(o *options) {
		o.notNeedHistory = false
	}
}

