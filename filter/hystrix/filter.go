/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package hystrix provides hystrix filter.
// To use hystrix, you need to configure commands using hystrix-go API:
//
//	import "github.com/afex/hystrix-go/hystrix"
//
//	// Resource name format: dubbo:consumer:InterfaceName:group:version:Method
//	// Example: dubbo:consumer:com.example.GreetService:::Greet
//	hystrix.ConfigureCommand("dubbo:consumer:com.example.GreetService:::Greet", hystrix.CommandConfig{
//	    Timeout:                1000,
//	    MaxConcurrentRequests:  20,
//	    RequestVolumeThreshold: 20,
//	    SleepWindow:            5000,
//	    ErrorPercentThreshold:  50,
//	})

package hystrix

import (
	"context"
	"fmt"
	"strings"
)

import (
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/filter"
	"dubbo.apache.org/dubbo-go/v3/protocol/base"
	"dubbo.apache.org/dubbo-go/v3/protocol/result"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/dubbogo/gost/log/logger"
)

const (
	// Filter keys
	HystrixConsumerFilterKey = "hystrix_consumer"
	HystrixProviderFilterKey = "hystrix_provider"

	// Prefixes for resource naming
	DefaultProviderPrefix = "dubbo:provider:"
	DefaultConsumerPrefix = "dubbo:consumer:"
)

func init() {
	extension.SetFilter(HystrixConsumerFilterKey, newFilterConsumer)
	extension.SetFilter(HystrixProviderFilterKey, newFilterProvider)
}

// FilterError implements error interface
type FilterError struct {
	err         error
	circuitOpen bool
}

func (hfError *FilterError) Error() string {
	return hfError.err.Error()
}

// CircuitOpen returns whether the circuit is open
func (hfError *FilterError) CircuitOpen() bool {
	return hfError.circuitOpen
}

// NewHystrixFilterError return a FilterError instance
func NewHystrixFilterError(err error, circuitOpen bool) error {
	return &FilterError{
		err:         err,
		circuitOpen: circuitOpen,
	}
}

// Filter for Hystrix
type Filter struct {
	isConsumer bool // true for consumer, false for provider
}

// Invoke is an implementation of filter, provides Hystrix pattern latency and fault tolerance
func (f *Filter) Invoke(ctx context.Context, invoker base.Invoker, invocation base.Invocation) result.Result {
	cmdName := getResourceName(invoker, invocation, f.isConsumer)

	var res result.Result
	err := hystrix.Do(cmdName, func() error {
		res = invoker.Invoke(ctx, invocation)
		return res.Error()
	}, func(err error) error {
		// Return fallback error
		_, isCircuitOpen := err.(hystrix.CircuitError)
		if isCircuitOpen {
			logger.Debugf("[Hystrix Filter] Circuit opened for %s", cmdName)
		} else {
			logger.Debugf("[Hystrix Filter] Hystrix fallback for %s: %v", cmdName, err)
		}
		res = &result.RPCResult{}
		res.SetResult(nil)
		res.SetError(NewHystrixFilterError(err, isCircuitOpen))
		return err
	})

	if err != nil {
		return res
	}
	return res
}

// OnResponse dummy process, returns the result directly
func (f *Filter) OnResponse(ctx context.Context, result result.Result, invoker base.Invoker, invocation base.Invocation) result.Result {
	return result
}

// newFilterConsumer returns Filter instance for consumer
func newFilterConsumer() filter.Filter {
	return &Filter{isConsumer: true}
}

// newFilterProvider returns Filter instance for provider
func newFilterProvider() filter.Filter {
	return &Filter{isConsumer: false}
}

func getResourceName(invoker base.Invoker, invocation base.Invocation, isConsumer bool) string {
	var sb strings.Builder

	if isConsumer {
		sb.WriteString(DefaultConsumerPrefix)
	} else {
		sb.WriteString(DefaultProviderPrefix)
	}

	// Format: interface:group:version:method
	sb.WriteString(getColonSeparatedKey(invoker.GetURL()))
	sb.WriteString(":")
	sb.WriteString(invocation.MethodName())

	return sb.String()
}

func getColonSeparatedKey(url *common.URL) string {
	return fmt.Sprintf("%s:%s:%s",
		url.Service(),
		url.GetParam(constant.GroupKey, ""),
		url.GetParam(constant.VersionKey, ""))
}
