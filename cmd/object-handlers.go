/*
 * Minio Cloud Storage, (C) 2015 Minio, Inc.
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
	//"net"
	"encoding/hex"
	"net/url"
)

// supportedGetReqParams - supported request parameters for GET presigned request.
var supportedGetReqParams = map[string]string{
	"response-expires":             "Expires",
	"response-content-type":        "Content-Type",
	"response-cache-control":       "Cache-Control",
	"response-content-encoding":    "Content-Encoding",
	"response-content-language":    "Content-Language",
	"response-content-disposition": "Content-Disposition",
}

func setGetRespHeaders(w http.ResponseWriter, reqParams url.Values) {
	for k, v := range reqParams {
		if header, ok := supportedGetReqParams[k]; ok {
			w.Header()[header] = v
		}
	}
}

type funcToWriter func([]byte) (int, error)

func (f funcToWriter) Write(p []byte) (int, error) {
	return f(p)
}

func (api objectAPIHandlers) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	var object, bucket string
	vars := mux.Vars(r)
	bucket = vars["bucket"]
	object = vars["object"]

	// Fetch object stat info.
	objectAPI := api.ObjectAPI()
	if objectAPI == nil {
		writeErrorResponse(w, ErrServerNotInitialized, r.URL)
		return
	}

	//if s3Error := checkRequestAuthType(r, bucket, "s3:GetObject", serverConfig.GetRegion()); s3Error != ErrNone {
	//	writeErrorResponse(w, s3Error, r.URL)
	//	return
	//}
	//
	//// Lock the object before reading.
	//objectLock := globalNSMutex.NewNSLock(bucket, object)
	//objectLock.RLock()
	//defer objectLock.RUnlock()

	objInfo, err := objectAPI.GetObjectInfo(bucket, object)
	if err != nil {
		println(err, "Unable to fetch object info.")
		apiErr := toAPIErrorCode(err)
		if apiErr == ErrNoSuchKey {
			apiErr = errAllowableObjectNotFound(bucket, r)
		}
		writeErrorResponse(w, apiErr, r.URL)
		return
	}

	// Get request range.
	var hrange *httpRange
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		if hrange, err = parseRequestRange(rangeHeader, objInfo.Size); err != nil {
			// Handle only errInvalidRange
			// Ignore other parse error and treat it as regular Get request like Amazon S3.
			if err == errInvalidRange {
				writeErrorResponse(w, ErrInvalidRange, r.URL)
				return
			}

			// log the error.
			println(err, "Invalid request range")
		}
	}

	// Validate pre-conditions if any.
	//if checkPreconditions(w, r, objInfo) {
	//	return
	//}

	// Get the object.
	var startOffset int64
	length := objInfo.Size
	if hrange != nil {
		startOffset = hrange.offsetBegin
		length = hrange.getLength()
	}

	// Indicates if any data was written to the http.ResponseWriter
	dataWritten := false
	// io.Writer type which keeps track if any data was written.
	writer := funcToWriter(func(p []byte) (int, error) {
		if !dataWritten {
			// Set headers on the first write.
			// Set standard object headers.
			setObjectHeaders(w, objInfo, hrange)

			// Set any additional requested response headers.
			setGetRespHeaders(w, r.URL.Query())

			dataWritten = true
		}
		return w.Write(p)
	})

	// Reads the object at startOffset and writes to mw.
	if err = objectAPI.GetObject(bucket, object, startOffset, length, writer); err != nil {
		println(err, "Unable to write to client.")
		if !dataWritten {
			// Error response only if no data has been written to client yet. i.e if
			// partial data has already been written before an error
			// occurred then no point in setting StatusCode and
			// sending error XML.
			writeErrorResponse(w, toAPIErrorCode(err), r.URL)
		}
		return
	}
	if !dataWritten {
		// If ObjectAPI.GetObject did not return error and no data has
		// been written it would mean that it is a 0-byte object.
		// call wrter.Write(nil) to set appropriate headers.
		writer.Write(nil)
	}
	//
	//// Get host and port from Request.RemoteAddr.
	//host, port, err := net.SplitHostPort(r.RemoteAddr)
	//if err != nil {
	//	host, port = "", ""
	//}

	// Notify object accessed via a GET request.
	//eventNotify(eventData{
	//	Type:      ObjectAccessedGet,
	//	Bucket:    bucket,
	//	ObjInfo:   objInfo,
	//	ReqParams: extractReqParams(r),
	//	UserAgent: r.UserAgent(),
	//	Host:      host,
	//	Port:      port,
	//})
}

func (api objectAPIHandlers) HeadObjectHandler(w http.ResponseWriter, r *http.Request) {
	var object, bucket string
	vars := mux.Vars(r)
	bucket = vars["bucket"]
	object = vars["object"]

	objectAPI := api.ObjectAPI()
	if objectAPI == nil {
		writeErrorResponseHeadersOnly(w, ErrServerNotInitialized)
		return
	}

	//if s3Error := checkRequestAuthType(r, bucket, "s3:GetObject", serverConfig.GetRegion()); s3Error != ErrNone {
	//	writeErrorResponseHeadersOnly(w, s3Error)
	//	return
	//}

	// Lock the object before reading.
	//objectLock := globalNSMutex.NewNSLock(bucket, object)
	//objectLock.RLock()
	//defer objectLock.RUnlock()

	objInfo, err := objectAPI.GetObjectInfo(bucket, object)
	if err != nil {
		println(err, "Unable to fetch object info.")
		apiErr := toAPIErrorCode(err)
		if apiErr == ErrNoSuchKey {
			apiErr = errAllowableObjectNotFound(bucket, r)
		}
		writeErrorResponseHeadersOnly(w, apiErr)
		return
	}

	// Validate pre-conditions if any.
	//if checkPreconditions(w, r, objInfo) {
	//	return
	//}

	// Set standard object headers.
	setObjectHeaders(w, objInfo, nil)

	// Successful response.
	w.WriteHeader(http.StatusOK)

	// Get host and port from Request.RemoteAddr.
	//host, port, err := net.SplitHostPort(r.RemoteAddr)
	//if err != nil {
	//	host, port = "", ""
	//}
	//
	//// Notify object accessed via a HEAD request.
	//eventNotify(eventData{
	//	Type:      ObjectAccessedHead,
	//	Bucket:    bucket,
	//	ObjInfo:   objInfo,
	//	ReqParams: extractReqParams(r),
	//	UserAgent: r.UserAgent(),
	//	Host:      host,
	//	Port:      port,
	//})
}



// ListObjectPartsHandler - List object parts
func (api objectAPIHandlers) ListObjectPartsHandler(w http.ResponseWriter, r *http.Request) {


	//response := generateListPartsResponse(nil)
	//encodedSuccessResponse := encodeResponse(response)

	// Write success response.
	writeSuccessResponseXML(w, nil)
}


func (api objectAPIHandlers) PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	objectAPI := api.ObjectAPI()
	if objectAPI == nil {
		writeErrorResponse(w, ErrServerNotInitialized, r.URL)
		return
	}

	// X-Amz-Copy-Source shouldn't be set for this call.
	if _, ok := r.Header["X-Amz-Copy-Source"]; ok {
		writeErrorResponse(w, ErrInvalidCopySource, r.URL)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	// Get Content-Md5 sent by client and verify if valid
	md5Bytes, err := checkValidMD5(r.Header.Get("Content-Md5"))
	if err != nil {
		println(err, "Unable to validate content-md5 format.")
		writeErrorResponse(w, ErrInvalidDigest, r.URL)
		return
	}

	/// if Content-Length is unknown/missing, deny the request
	size := r.ContentLength
	//rAuthType := getRequestAuthType(r)
	//if rAuthType == authTypeStreamingSigned {
	//	sizeStr := r.Header.Get("x-amz-decoded-content-length")
	//	size, err = strconv.ParseInt(sizeStr, 10, 64)
	//	if err != nil {
	//		errorIf(err, "Unable to parse `x-amz-decoded-content-length` into its integer value", sizeStr)
	//		writeErrorResponse(w, toAPIErrorCode(err), r.URL)
	//		return
	//	}
	//}
	if size == -1 {
		writeErrorResponse(w, ErrMissingContentLength, r.URL)
		return
	}

	/// maximum Upload size for objects in a single operation
	if isMaxObjectSize(size) {
		writeErrorResponse(w, ErrEntityTooLarge, r.URL)
		return
	}

	// Extract metadata to be saved from incoming HTTP header.
	metadata := extractMetadataFromHeader(r.Header)
	//if rAuthType == authTypeStreamingSigned {
	//	if contentEncoding, ok := metadata["content-encoding"]; ok {
	//		contentEncoding = trimAwsChunkedContentEncoding(contentEncoding)
	//		if contentEncoding != "" {
	//			// Make sure to trim and save the content-encoding
	//			// parameter for a streaming signature which is set
	//			// to a custom value for example: "aws-chunked,gzip".
	//			metadata["content-encoding"] = contentEncoding
	//		} else {
	//			// Trimmed content encoding is empty when the header
	//			// value is set to "aws-chunked" only.
	//
	//			// Make sure to delete the content-encoding parameter
	//			// for a streaming signature which is set to value
	//			// for example: "aws-chunked"
	//			delete(metadata, "content-encoding")
	//		}
	//	}
	//}

	// Make sure we hex encode md5sum here.
	metadata["etag"] = hex.EncodeToString(md5Bytes)

	sha256sum := ""

	// Lock the object.
	//objectLock := globalNSMutex.NewNSLock(bucket, object)
	//objectLock.Lock()
	//defer objectLock.Unlock()

	var objInfo ObjectInfo
	objInfo, err = objectAPI.PutObject(bucket, object, size, r.Body, metadata, sha256sum)
	//switch rAuthType {
	//default:
	//	// For all unknown auth types return error.
	//	writeErrorResponse(w, ErrAccessDenied, r.URL)
	//	return
	//case authTypeAnonymous:
	//	// http://docs.aws.amazon.com/AmazonS3/latest/dev/using-with-s3-actions.html
	//	if s3Error := enforceBucketPolicy(bucket, "s3:PutObject", r.URL.Path,
	//		r.Referer(), r.URL.Query()); s3Error != ErrNone {
	//		writeErrorResponse(w, s3Error, r.URL)
	//		return
	//	}
	//	// Create anonymous object.
	//	objInfo, err = objectAPI.PutObject(bucket, object, size, r.Body, metadata, sha256sum)
	//case authTypeStreamingSigned:
	//	// Initialize stream signature verifier.
	//	reader, s3Error := newSignV4ChunkedReader(r)
	//	if s3Error != ErrNone {
	//		errorIf(errSignatureMismatch, dumpRequest(r))
	//		writeErrorResponse(w, s3Error, r.URL)
	//		return
	//	}
	//	objInfo, err = objectAPI.PutObject(bucket, object, size, reader, metadata, sha256sum)
	//case authTypeSignedV2, authTypePresignedV2:
	//	s3Error := isReqAuthenticatedV2(r)
	//	if s3Error != ErrNone {
	//		errorIf(errSignatureMismatch, dumpRequest(r))
	//		writeErrorResponse(w, s3Error, r.URL)
	//		return
	//	}
	//	objInfo, err = objectAPI.PutObject(bucket, object, size, r.Body, metadata, sha256sum)
	//case authTypePresigned, authTypeSigned:
	//	if s3Error := reqSignatureV4Verify(r, serverConfig.GetRegion()); s3Error != ErrNone {
	//		errorIf(errSignatureMismatch, dumpRequest(r))
	//		writeErrorResponse(w, s3Error, r.URL)
	//		return
	//	}
	//	if !skipContentSha256Cksum(r) {
	//		sha256sum = r.Header.Get("X-Amz-Content-Sha256")
	//	}
	//	// Create object.
	//	objInfo, err = objectAPI.PutObject(bucket, object, size, r.Body, metadata, sha256sum)
	//}
	if err != nil {
		println(err, "Unable to create an object. %s", r.URL.Path)
		writeErrorResponse(w, toAPIErrorCode(err), r.URL)
		return
	}
	w.Header().Set("ETag", "\""+objInfo.ETag+"\"")
	writeSuccessResponseHeadersOnly(w)

	// Get host and port from Request.RemoteAddr.
	//host, port, err := net.SplitHostPort(r.RemoteAddr)
	//if err != nil {
	//	host, port = "", ""
	//}

	// Notify object created event.
	//eventNotify(eventData{
	//	Type:      ObjectCreatedPut,
	//	Bucket:    bucket,
	//	ObjInfo:   objInfo,
	//	ReqParams: extractReqParams(r),
	//	UserAgent: r.UserAgent(),
	//	Host:      host,
	//	Port:      port,
	//})
}


func errAllowableObjectNotFound(bucket string, r *http.Request) APIErrorCode {
	if getRequestAuthType(r) == authTypeAnonymous {
		//we care about the bucket as a whole, not a particular resource
		//resource := "/" + bucket
		//if s3Error := enforceBucketPolicy(bucket, "s3:ListBucket", resource,
		//	r.Referer(), r.URL.Query()); s3Error != ErrNone {
		//	return ErrAccessDenied
		//}
	}
	return ErrNoSuchKey
}