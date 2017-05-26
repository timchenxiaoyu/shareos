/*
 * Minio Cloud Storage, (C) 2016, 2017, 2017 Minio, Inc.
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
"runtime"
	"time"
)

const (
	// Erasure related constants.
	erasureAlgorithmKlauspost = "klauspost/reedsolomon/vandermonde"
)

// objectPartInfo Info of each part kept in the multipart metadata
// file after CompleteMultipartUpload() is called.
type objectPartInfo struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
	ETag   string `json:"etag"`
	Size   int64  `json:"size"`
}

// byObjectPartNumber is a collection satisfying sort.Interface.
type byObjectPartNumber []objectPartInfo

func (t byObjectPartNumber) Len() int           { return len(t) }
func (t byObjectPartNumber) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byObjectPartNumber) Less(i, j int) bool { return t[i].Number < t[j].Number }

// checkSumInfo - carries checksums of individual scattered parts per disk.
type checkSumInfo struct {
	Name      string   `json:"name"`
	Algorithm HashAlgo `json:"algorithm"`
	Hash      string   `json:"hash"`
}

// HashAlgo - represents a supported hashing algorithm for bitrot
// verification.
type HashAlgo string

const (
	// HashBlake2b represents the Blake 2b hashing algorithm
	HashBlake2b HashAlgo = "blake2b"
	// HashSha256 represents the SHA256 hashing algorithm
	HashSha256 HashAlgo = "sha256"
)

// isValidHashAlgo - function that checks if the hash algorithm is
// valid (known and used).
func isValidHashAlgo(algo HashAlgo) bool {
	switch algo {
	case HashSha256, HashBlake2b:
		return true
	default:
		return false
	}
}

// Constant indicates current bit-rot algo used when creating objects.
// Depending on the architecture we are choosing a different checksum.
var bitRotAlgo = getDefaultBitRotAlgo()

// Get the default bit-rot algo depending on the architecture.
// Currently this function defaults to "blake2b" as the preferred
// checksum algorithm on all architectures except ARM64. On ARM64
// we use sha256 (optimized using sha2 instructions of ARM NEON chip).
func getDefaultBitRotAlgo() HashAlgo {
	switch runtime.GOARCH {
	case "arm64":
		// As a special case for ARM64 we use an optimized
		// version of hash i.e sha256. This is done so that
		// blake2b is sub-optimal and slower on ARM64.
		// This would also allows erasure coded writes
		// on ARM64 servers to be on-par with their
		// counter-part X86_64 servers.
		return HashSha256
	default:
		// Default for all other architectures we use blake2b.
		return HashBlake2b
	}
}

// erasureInfo - carries erasure coding related information, block
// distribution and checksums.
type erasureInfo struct {
	Algorithm    HashAlgo       `json:"algorithm"`
	DataBlocks   int            `json:"data"`
	ParityBlocks int            `json:"parity"`
	BlockSize    int64          `json:"blockSize"`
	Index        int            `json:"index"`
	Distribution []int          `json:"distribution"`
	Checksum     []checkSumInfo `json:"checksum,omitempty"`
}

// AddCheckSum - add checksum of a part.
func (e *erasureInfo) AddCheckSumInfo(ckSumInfo checkSumInfo) {
	for i, sum := range e.Checksum {
		if sum.Name == ckSumInfo.Name {
			e.Checksum[i] = ckSumInfo
			return
		}
	}
	e.Checksum = append(e.Checksum, ckSumInfo)
}

// GetCheckSumInfo - get checksum of a part.
func (e erasureInfo) GetCheckSumInfo(partName string) (ckSum checkSumInfo) {
	// Return the checksum.
	for _, sum := range e.Checksum {
		if sum.Name == partName {
			return sum
		}
	}
	return checkSumInfo{Algorithm: bitRotAlgo}
}

// statInfo - carries stat information of the object.
type statInfo struct {
	Size    int64     `json:"size"`    // Size of the object `xl.json`.
	ModTime time.Time `json:"modTime"` // ModTime of the object `xl.json`.
}

// A xlMetaV1 represents `xl.json` metadata header.
type xlMetaV1 struct {
	Version string   `json:"version"` // Version of the current `xl.json`.
	Format  string   `json:"format"`  // Format of the current `xl.json`.
	Stat    statInfo `json:"stat"`    // Stat of the current object `xl.json`.
	// Erasure coded info for the current object `xl.json`.
	Erasure erasureInfo `json:"erasure"`
	// Minio release tag for current object `xl.json`.
	Minio struct {
		Release string `json:"release"`
	} `json:"minio"`
	// Metadata map for current object `xl.json`.
	Meta map[string]string `json:"meta,omitempty"`
	// Captures all the individual object `xl.json`.
	Parts []objectPartInfo `json:"parts,omitempty"`
}

// XL metadata constants.
const (
	// XL meta version.
	xlMetaVersion = "1.0.1"

	// XL meta version.
	xlMetaVersion100 = "1.0.0"

	// XL meta format string.
	xlMetaFormat = "xl"

	// Add new constants here.
)
