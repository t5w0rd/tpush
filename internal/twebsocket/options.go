package twebsocket

import "time"

type Options struct {
	mux         *ServeMux
	recvTimeout time.Duration
	sendTimeout time.Duration

	upgradeHandler UpgradeHandler
	openHandler    OpenHandler
	closeHandler   CloseHandler
}

type Option func(opt *Options)

func WithServeMux(mux *ServeMux) Option {
	return func(opt *Options) {
		opt.mux = mux
	}
}

func WithRecvTimeout(timeout time.Duration) Option {
	return func(opt *Options) {
		opt.recvTimeout = timeout
	}
}

func WithSendTimeout(timeout time.Duration) Option {
	return func(opt *Options) {
		opt.sendTimeout = timeout
	}
}

func WithUpgradeHandler(handler UpgradeHandler) Option {
	return func(opt *Options) {
		opt.upgradeHandler = handler
	}
}

func WithOpenHandler(handler OpenHandler) Option {
	return func(opt *Options) {
		opt.openHandler = handler
	}
}

func WithCloseHandler(handler CloseHandler) Option {
	return func(opt *Options) {
		opt.closeHandler = handler
	}
}
