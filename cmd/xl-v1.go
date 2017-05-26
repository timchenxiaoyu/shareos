/*
 * Minio Cloud Storage, (C) 2016 Minio, Inc.
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

	"sync"

	humanize "github.com/dustin/go-humanize"

	"github.com/minio/minio/pkg/objcache"
)

// XL constants.
const (
	// Format config file carries backend format specific details.
	formatConfigFile = "format.json"

	// Format config tmp file carries backend format.
	formatConfigFileTmp = "format.json.tmp"

	// XL metadata file carries per object metadata.
	xlMetaJSONFile = "xl.json"

	// Uploads metadata file carries per multipart object metadata.
	uploadsJSONFile = "uploads.json"

	// Represents the minimum required RAM size to enable caching.
	minRAMSize = 24 * humanize.GiByte

	// Maximum erasure blocks.
	maxErasureBlocks = 16

	// Minimum erasure blocks.
	minErasureBlocks = 4
)

// xlObjects - Implements XL object layer.
type xlObjects struct {
	mutex        *sync.Mutex
	storageDisks []StorageAPI // Collection of initialized backend disks.
	dataBlocks   int          // dataBlocks count caculated for erasure.
	parityBlocks int          // parityBlocks count calculated for erasure.
	readQuorum   int          // readQuorum minimum required disks to read data.
	writeQuorum  int          // writeQuorum minimum required disks to write data.

	// ListObjects pool management.
	listPool *treeWalkPool

	// Object cache for caching objects.
	objCache *objcache.Cache

	// Object cache enabled.
	objCacheEnabled bool
}

// list of all errors that can be ignored in tree walk operation in XL
var xlTreeWalkIgnoredErrs = append(baseIgnoredErrs, errDiskAccessDenied, errVolumeNotFound, errFileNotFound)

