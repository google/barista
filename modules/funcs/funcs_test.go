// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package funcs

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/soumya92/barista/bar"
	"github.com/soumya92/barista/outputs"
	testBar "github.com/soumya92/barista/testing/bar"
)

var count = int64(0)

func doFunc(s bar.Sink) {
	newCount := atomic.AddInt64(&count, 1)
	if newCount < 4 {
		s.Output(outputs.Textf("%d", newCount))
	} else {
		s.Error(fmt.Errorf("something"))
	}
}

func TestOneShot(t *testing.T) {
	require := require.New(t)
	testBar.New(t)
	atomic.StoreInt64(&count, 0)

	module := Once(doFunc)
	require.Equal(int64(0), atomic.LoadInt64(&count),
		"Function isn't called until module starts streaming")

	testBar.Run(module)
	testBar.NextOutput().AssertText(
		[]string{"1"}, "Function is never called again")

	testBar.AssertNoOutput("No output is sent")
	require.Equal(int64(1), atomic.LoadInt64(&count),
		"Function is never called again")
}

func TestOnClick(t *testing.T) {
	testBar.New(t)
	require := require.New(t)
	atomic.StoreInt64(&count, 0)

	module := OnClick(doFunc)
	require.Equal(int64(0), atomic.LoadInt64(&count),
		"Function isn't called until module starts streaming")

	testBar.Run(module)
	testBar.NextOutput().AssertText(
		[]string{"1"}, "Function called when streaming")
	testBar.AssertNoOutput("Function is not called again")
	testBar.Tick()
	testBar.AssertNoOutput("Function is not called again")

	testBar.Click(0)
	testBar.LatestOutput().AssertText(
		[]string{"2"}, "Function called again on click")

	testBar.Click(0)
	testBar.LatestOutput().AssertText(
		[]string{"3"}, "Function called again on click")
}

func TestRepeated(t *testing.T) {
	require := require.New(t)
	testBar.New(t)
	atomic.StoreInt64(&count, 0)

	module := Every(time.Minute, doFunc)
	require.Equal(int64(0), atomic.LoadInt64(&count),
		"Function isn't called until module starts streaming")

	testBar.Run(module)
	testBar.NextOutput().AssertText(
		[]string{"1"}, "Function called when streaming")
	testBar.AssertNoOutput("Function is not called again until next tick")
	testBar.Tick()
	testBar.NextOutput().AssertText(
		[]string{"2"}, "Function is called on next tick")
	testBar.Tick()
	testBar.NextOutput().AssertText(
		[]string{"3"}, "Function is called on next tick")
	testBar.Tick()
	testBar.NextOutput().AssertError("When function calls Error(...)")
	atomic.StoreInt64(&count, 0)
	testBar.Tick()
	testBar.LatestOutput().AssertText(
		[]string{"1"}, "Function is called on tick after Error")
	testBar.Tick()
	testBar.NextOutput().AssertText(
		[]string{"2"}, "Function is called on next tick")
	testBar.Tick()
	testBar.NextOutput().AssertText(
		[]string{"3"}, "Function is called on next tick")
}
