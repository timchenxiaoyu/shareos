/*
 * Minio Cloud Storage, (C) 2015, 2016, 2017 Minio, Inc.
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
	mux "github.com/gorilla/mux"

)


// ListBucketsHandler - GET Service.
// -----------
// This implementation of the GET operation returns a list of all buckets
// owned by the authenticated sender of the request.
func (api objectAPIHandlers) ListBucketsHandler(w http.ResponseWriter, r *http.Request) {
	objectAPI := api.ObjectAPI()


	// Invoke the list buckets.
	bucketsInfo, err := objectAPI.ListBuckets()
	if err != nil {

		return
	}

	// Generate response.
	response := generateListBucketsResponse(bucketsInfo)
	encodedSuccessResponse := encodeResponse(response)

	// Write response.
	writeSuccessResponseXML(w, encodedSuccessResponse)
}

func (api objectAPIHandlers) PutBucketHandler(w http.ResponseWriter, r *http.Request) {
	objectAPI := api.ObjectAPI()
	if objectAPI == nil {
		writeErrorResponse(w, ErrServerNotInitialized, r.URL)
		return
	}



	vars := mux.Vars(r)
	bucket := vars["bucket"]

	// Parse incoming location constraint.
	//location, s3Error := parseLocationConstraint(r)
	//if s3Error != ErrNone {
	//	writeErrorResponse(w, s3Error, r.URL)
	//	return
	//}

	// Validate if location sent by the client is valid, reject
	// requests which do not follow valid region requirements.
	//if !isValidLocation(location) {
	//	writeErrorResponse(w, ErrInvalidRegion, r.URL)
	//	return
	//}

	//bucketLock := globalNSMutex.NewNSLock(bucket, "")
	//bucketLock.Lock()
	//defer bucketLock.Unlock()

	// Proceed to creating a bucket.
	err := objectAPI.MakeBucket(bucket)
	if err != nil {
		println(err, "Unable to create a bucket.")
		writeErrorResponse(w, toAPIErrorCode(err), r.URL)
		return
	}

	// Make sure to add Location information here only for bucket
	w.Header().Set("Location", getLocation(r))

	writeSuccessResponseHeadersOnly(w)
}