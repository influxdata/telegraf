//Copyright (c) 2019, Groupon, Inc.
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without
//modification, are permitted provided that the following conditions are
//met:
//
//Redistributions of source code must retain the above copyright notice,
//this list of conditions and the following disclaimer.
//
//Redistributions in binary form must reproduce the above copyright
//notice, this list of conditions and the following disclaimer in the
//documentation and/or other materials provided with the distribution.
//
//Neither the name of GROUPON nor the names of its contributors may be
//used to endorse or promote products derived from this software without
//specific prior written permission.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
//IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
//TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
//PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
//HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
//SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
//TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
//PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
//LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
//NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package tdigestagg

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMergingDigest(t *testing.T) {
	rand.Seed(time.Now().Unix())

	td := NewTDigest(1000)

	for i := 0; i < 1000000; i++ {
		td.Add(rand.Float64(), 1.0)
	}
	validateMergingDigest(t, td)

	// we don't bother testing the CDF here, it's not very precise at the median
	assert.InEpsilon(t, 0.5, td.Quantile(0.5), 0.02, "median was %v, not 0.5", td.Quantile(0.5))
	assert.True(t, td.Min() >= 0, "minimum was %v, expected non-negative", td.Min())
	assert.True(t, td.Max() < 1, "maximum was %v, expected below 1", td.Max())
}

func TestMergeSparseDigest(t *testing.T) {
	td := NewTDigest(1000)
	td.Add(-200000, 1)
	other := NewTDigest(1000)
	other.Add(200000, 1)

	td.Merge(other)
	validateMergingDigest(t, td)

	assert.InEpsilon(t, 0.5, td.CDF(0), 0.02, "cdf below 0 was %v, not ~50%", td.CDF(0))
	// epsilon-style ULP comparisons do not work on zero
	assert.InDelta(t, 0, td.Quantile(0.5), 0.02, "median was %v, not 0", td.Quantile(0.5))
	assert.InEpsilon(t, td.Min(), td.Quantile(0), 0.02, "minimum was %v", td.Quantile(0))
	assert.InEpsilon(t, td.Max(), td.Quantile(1), 0.02, "maximum was %v", td.Quantile(1))
}

// check the basic validity of a merging t-digest
// are its centroids within the sizing bound?
// do its weights add up?
func validateMergingDigest(t *testing.T, td *TDigest) {
	td.mergeTemps()

	index := 0.0
	quantile := 0.0
	runningWeight := 0.0
	for i, c := range td.mainCentroids {
		nextIndex := td.indexEstimate(quantile + c.Weight/td.mainWeight)
		// avoid checking the first and last centroids
		// they're under the strictest expectations so they often fail
		if i != 0 && i != len(td.mainCentroids)-1 {
			assert.True(t, nextIndex-index <= 1 || c.Weight == 1, "centroid is oversized: ", c)
		}

		quantile += c.Weight / td.mainWeight
		index = nextIndex
		runningWeight += c.Weight
	}

	assert.Equal(t, td.mainWeight, runningWeight, "total weights didn't add up")
}

func TestGobEncoding(t *testing.T) {
	rand.Seed(time.Now().Unix())

	td := NewTDigest(1000)
	for i := 0; i < 1000; i++ {
		td.Add(rand.Float64(), 1.0)
	}
	validateMergingDigest(t, td)

	buf, err := td.GobEncode()
	assert.NoError(t, err, "should have encoded successfully")

	td2 := NewTDigest(1000)
	assert.NoError(t, td2.GobDecode(buf), "should have decoded successfully")

	assert.InEpsilon(t, td.Count(), td2.Count(), 0.02, "counts did not match")
	assert.InEpsilon(t, td.Min(), td2.Min(), 0.02, "minimums did not match")
	assert.InEpsilon(t, td.Max(), td2.Max(), 0.02, "maximums did not match")
	assert.InEpsilon(t, td.Quantile(0.5), td2.Quantile(0.5), 0.02, "50%% quantiles did not match")
}

func BenchmarkAdd(b *testing.B) {
	rand.Seed(time.Now().Unix())
	td := NewTDigest(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		td.Add(rand.NormFloat64(), 1.0)
	}
}

func BenchmarkQuantile(b *testing.B) {
	rand.Seed(time.Now().Unix())
	td := NewTDigest(1000)
	for i := 0; i < b.N; i++ {
		td.Add(rand.NormFloat64(), 1.0)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		td.Quantile(rand.Float64())
	}
}
