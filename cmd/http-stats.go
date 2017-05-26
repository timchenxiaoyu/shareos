/*
 * Minio Cloud Storage, (C) 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"time"
"net/http"
	"go.uber.org/atomic"
)

// ConnStats - Network statistics
// Count total input/output transferred bytes during
// the server's life.
type ConnStats struct {
	totalInputBytes  atomic.Uint64
	totalOutputBytes atomic.Uint64
}

// Increase total input bytes
func (s *ConnStats) incInputBytes(n int) {
	s.totalInputBytes.Add(uint64(n))
}

// Increase total output bytes
func (s *ConnStats) incOutputBytes(n int) {
	s.totalOutputBytes.Add(uint64(n))
}

// Return total input bytes
func (s *ConnStats) getTotalInputBytes() uint64 {
	return s.totalInputBytes.Load()
}

// Return total output bytes
func (s *ConnStats) getTotalOutputBytes() uint64 {
	return s.totalOutputBytes.Load()
}


// Prepare new ConnStats structure
func newConnStats() *ConnStats {
	return &ConnStats{}
}

// HTTPMethodStats holds statistics information about
// a given HTTP method made by all clients
type HTTPMethodStats struct {
	Counter  atomic.Uint64
	Duration atomic.Float64
}

// HTTPStats holds statistics information about
// HTTP requests made by all clients
type HTTPStats struct {
	// HEAD request stats.
	totalHEADs   HTTPMethodStats
	successHEADs HTTPMethodStats

	// GET request stats.
	totalGETs   HTTPMethodStats
	successGETs HTTPMethodStats

	// PUT request stats.
	totalPUTs   HTTPMethodStats
	successPUTs HTTPMethodStats

	// POST request stats.
	totalPOSTs   HTTPMethodStats
	successPOSTs HTTPMethodStats

	// DELETE request stats.
	totalDELETEs   HTTPMethodStats
	successDELETEs HTTPMethodStats
}

func durationStr(totalDuration, totalCount float64) string {
	return fmt.Sprint(time.Duration(totalDuration/totalCount) * time.Second)
}

// Update statistics from http request and response data
func (st *HTTPStats) updateStats(r *http.Request, w *httpResponseRecorder, durationSecs float64) {
	// A successful request has a 2xx response code
	successReq := (w.respStatusCode >= 200 && w.respStatusCode < 300)
	// Update stats according to method verb
	switch r.Method {
	case "HEAD":
		st.totalHEADs.Counter.Inc()
		st.totalHEADs.Duration.Add(durationSecs)
		if successReq {
			st.successHEADs.Counter.Inc()
			st.successHEADs.Duration.Add(durationSecs)
		}
	case "GET":
		st.totalGETs.Counter.Inc()
		st.totalGETs.Duration.Add(durationSecs)
		if successReq {
			st.successGETs.Counter.Inc()
			st.successGETs.Duration.Add(durationSecs)
		}
	case "PUT":
		st.totalPUTs.Counter.Inc()
		st.totalPUTs.Duration.Add(durationSecs)
		if successReq {
			st.successPUTs.Counter.Inc()
			st.totalPUTs.Duration.Add(durationSecs)
		}
	case "POST":
		st.totalPOSTs.Counter.Inc()
		st.totalPOSTs.Duration.Add(durationSecs)
		if successReq {
			st.successPOSTs.Counter.Inc()
			st.totalPOSTs.Duration.Add(durationSecs)
		}
	case "DELETE":
		st.totalDELETEs.Counter.Inc()
		st.totalDELETEs.Duration.Add(durationSecs)
		if successReq {
			st.successDELETEs.Counter.Inc()
			st.successDELETEs.Duration.Add(durationSecs)
		}
	}
}


// Prepare new HTTPStats structure
func newHTTPStats() *HTTPStats {
	return &HTTPStats{}
}
