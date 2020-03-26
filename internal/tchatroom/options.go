package tchatroom

type Options struct {
	distribute Distribute
}

type Option func(opt *Options)

func WithDistribute(distribute Distribute) Option {
	return func(opt *Options) {
		opt.distribute = distribute
	}
}
