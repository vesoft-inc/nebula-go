/* Copyright (c) 2019 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package nebula_client

import "time"

type Options struct {
	Timeout time.Duration
}

type Option func(*Options)

var defaultOptions = Options{
	Timeout: 30 * time.Second,
}

func WithTimeout(duration time.Duration) Option {
	return func(options *Options) {
		options.Timeout = duration
	}
}
