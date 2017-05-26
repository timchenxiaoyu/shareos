/*
 * Minio Cloud Storage, (C) 2015, 2016 Minio, Inc.
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
	"net/http"

	router "github.com/gorilla/mux"
)

func newObjectLayerFn() (layer ObjectLayer) {
	globalObjLayerMutex.RLock()
	layer = globalObjectAPI
	globalObjLayerMutex.RUnlock()
	return
}

// configureServer handler returns final handler for the http server.
func configureServerHandler(endpoints EndpointList) (http.Handler, error) {
	// Initialize router. `SkipClean(true)` stops gorilla/mux from
	// normalizing URL path minio/minio#3256
	mux := router.NewRouter().SkipClean(true)

	// Initialize distributed NS lock.

	// Add API router.
	registerAPIRouter(mux)

	var handlerFns = []HandlerFunc{}

	// Register rest of the handlers.
	return registerHandlers(mux, handlerFns...), nil
}
