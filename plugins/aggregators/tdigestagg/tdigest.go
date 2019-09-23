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

// Package tdigest provides an implementation of Ted Dunning's t-digest, an
// approximate histogram for online, distributed applications. For more details,
// refer to Dunning's paper and the reference implementations.
//
// https://github.com/tdunning/t-digest/blob/master/docs/t-digest-paper/histo.pdf
//
// https://github.com/tdunning/t-digest/blob/master/src/main/java/com/tdunning/math/stats/
//
// This implementation is largely based on https://github.com/stripe/veneur/tree/master/tdigest
package tdigestagg

import (
	"bytes"
	"encoding/gob"
	"math"
	"math/rand"
	"sort"
)

// A t-digest using the merging implementation. MergingDigest is not safe for
// use by multiple goroutines simultaneously, and its methods must not be
// invoked concurrently (including Quantile and CDF).
type TDigest struct {
	compression float64

	// main list of centroids
	mainCentroids []Centroid
	// total weight of main centroids
	mainWeight float64

	// centroids that have been added but not yet merged into main list
	tempCentroids []Centroid
	// total weight of unmerged centroids
	tempWeight float64

	min float64
	max float64
}

// fields must be exported to allow encoding
type Centroid struct {
	Mean   float64
	Weight float64
}

type TDigestOut struct {
	Compression float64

	// main list of centroids
	Centroids []Centroid
	// total weight of main centroids
	Weight float64

	Min float64
	Max float64
}

func (td *TDigest) ForJson() TDigestOut {
	td.mergeTemps()

	return TDigestOut{
		td.compression,
		td.mainCentroids,
		td.mainWeight,
		td.min,
		td.max,
	}
}

var _ sort.Interface = centroidList{}

// sort centroids by their mean
type centroidList []Centroid

func (cl centroidList) Len() int {
	return len(cl)
}
func (cl centroidList) Less(i, j int) bool {
	return cl[i].Mean < cl[j].Mean
}
func (cl centroidList) Swap(i, j int) {
	cl[i], cl[j] = cl[j], cl[i]
}

// Initializes a new t-digest using the given compression parameter.
// Lower compression values result in reduced memory consumption and less
// precision, especially at the median. Values from 20 to 1000 are recommended
// in Dunning's paper.
func NewTDigest(compression float64) *TDigest {
	// this is a provable upper bound on the size of the centroid list
	// TODO: derive it myself
	sizeBound := int((math.Pi * compression / 2) + 0.5)

	return &TDigest{
		compression:   compression,
		mainCentroids: make([]Centroid, 0, sizeBound),
		tempCentroids: make([]Centroid, 0, estimateTempBuffer(compression)),
		min:           math.Inf(+1),
		max:           math.Inf(-1),
	}
}

func estimateTempBuffer(compression float64) int {
	// this heuristic comes from Dunning's paper
	// 925 is the maximum point of this quadratic equation
	// TODO: let's derive and justify this heuristic
	tempCompression := math.Min(925, math.Max(20, compression))
	return int(7.5 + 0.37*tempCompression - 2e-4*tempCompression*tempCompression)
}

// Adds a new value to the t-digest, with a given weight that must be positive.
// Infinities and NaN cannot be added.
func (td *TDigest) Add(value float64, weight float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) || weight <= 0 {
		panic("invalid value added")
	}

	if len(td.tempCentroids) == cap(td.tempCentroids) {
		td.mergeTemps()
	}

	td.min = math.Min(td.min, value)
	td.max = math.Max(td.max, value)

	next := Centroid{
		Mean:   value,
		Weight: weight,
	}
	td.tempCentroids = append(td.tempCentroids, next)
	td.tempWeight += weight
}

// combine the mainCentroids and tempCentroids in-place into mainCentroids
func (td *TDigest) mergeTemps() {
	// this optimization is really important! if you remove it, the main list
	// will get merged into itself every time this is called
	if len(td.tempCentroids) == 0 {
		return
	}

	// we iterate over both centroid lists from least to greatest mean, so first
	// we have to sort this one
	sort.Sort(centroidList(td.tempCentroids))
	tempIndex := 0

	// total weight that the final t-digest will have, after everything is merged
	totalWeight := td.mainWeight + td.tempWeight
	// how much weight has been merged so far
	mergedWeight := 0.0
	// the index of the last quantile to be merged into the previous centroid
	// this value gets updated each time we split a new centroid out instead of
	// merging into the current one
	lastMergedIndex := 0.0
	// since we will be merging in-place into td.mainCentroids, we need to keep
	// track of the indices of the remaining elements
	actualMainCentroids := td.mainCentroids
	td.mainCentroids = td.mainCentroids[:0]
	// to facilitate the in-place merge, we will need a place to store the main
	// centroids that would be overwritten - we will use space from the start
	// of tempCentroids for this
	swappedCentroids := td.tempCentroids[:0]

	for len(actualMainCentroids)+len(swappedCentroids) != 0 || tempIndex < len(td.tempCentroids) {
		nextTemp := Centroid{
			Mean:   math.Inf(+1),
			Weight: 0,
		}
		if tempIndex < len(td.tempCentroids) {
			nextTemp = td.tempCentroids[tempIndex]
		}

		nextMain := Centroid{
			Mean:   math.Inf(+1),
			Weight: 0,
		}
		if len(swappedCentroids) != 0 {
			nextMain = swappedCentroids[0]
		} else if len(actualMainCentroids) != 0 {
			nextMain = actualMainCentroids[0]
		}

		if nextMain.Mean < nextTemp.Mean {
			if len(actualMainCentroids) != 0 {
				if len(swappedCentroids) != 0 {
					// if this came from swap, before merging, we have to save
					// the next main centroid at the end
					// this copy is probably the most expensive part of the
					// in-place merge, compared to merging into a separate buffer
					copy(swappedCentroids, swappedCentroids[1:])
					swappedCentroids[len(swappedCentroids)-1] = actualMainCentroids[0]
				}
				actualMainCentroids = actualMainCentroids[1:]
			} else {
				// the real main has been completely exhausted, so we're just
				// cleaning out swapped mains now
				swappedCentroids = swappedCentroids[1:]
			}

			lastMergedIndex = td.mergeOne(mergedWeight, totalWeight, lastMergedIndex, nextMain)
			mergedWeight += nextMain.Weight
		} else {
			// before merging, we have to save the next main centroid somewhere
			// else, so that we don't overwrite it
			if len(actualMainCentroids) != 0 {
				swappedCentroids = append(swappedCentroids, actualMainCentroids[0])
				actualMainCentroids = actualMainCentroids[1:]
			}
			tempIndex++

			lastMergedIndex = td.mergeOne(mergedWeight, totalWeight, lastMergedIndex, nextTemp)
			mergedWeight += nextTemp.Weight
		}
	}

	td.tempCentroids = td.tempCentroids[:0]
	td.tempWeight = 0
	td.mainWeight = totalWeight
}

// merges a single centroid into the mergedCentroids list
// note that "merging" sometimes creates a new centroid in the list, however
// the length of the list has a strict upper bound (see constructor)
func (td *TDigest) mergeOne(beforeWeight, totalWeight, beforeIndex float64, next Centroid) float64 {
	// compute the quantile index of the element we're about to merge
	nextIndex := td.indexEstimate((beforeWeight + next.Weight) / totalWeight)

	if nextIndex-beforeIndex > 1 || len(td.mainCentroids) == 0 {
		// the new index is far away from the last index of the current centroid
		// therefore we cannot merge into the current centroid or it would
		// become too wide, so we will append a new centroid
		td.mainCentroids = append(td.mainCentroids, next)
		// return the last index that was merged into the previous centroid
		return td.indexEstimate(beforeWeight / totalWeight)
	} else {
		// the new index fits into the range of the current centroid, so we
		// combine it into the current centroid's values
		// this computation is known as welford's method, the order matters
		// weight must be updated before mean
		td.mainCentroids[len(td.mainCentroids)-1].Weight += next.Weight
		td.mainCentroids[len(td.mainCentroids)-1].Mean += (next.Mean - td.mainCentroids[len(td.mainCentroids)-1].Mean) * next.Weight / td.mainCentroids[len(td.mainCentroids)-1].Weight
		// we did not create a new centroid, so the trailing index of the previous
		// centroid remains
		return beforeIndex
	}
}

// given a quantile, estimate the index of the centroid that contains it using
// the given compression
func (td *TDigest) indexEstimate(quantile float64) float64 {
	// TODO: a polynomial approximation of arcsine should be a lot faster
	return td.compression * ((math.Asin(2*quantile-1) / math.Pi) + 0.5)
}

// Returns the approximate percentage of values in td that are below value (ie
// the cumulative distribution function). Returns NaN if the digest is empty.
func (td *TDigest) CDF(value float64) float64 {
	td.mergeTemps()

	if len(td.mainCentroids) == 0 {
		return math.NaN()
	}
	if value <= td.min {
		return 0
	}
	if value >= td.max {
		return 1
	}

	weightSoFar := 0.0
	lowerBound := td.min
	for i, c := range td.mainCentroids {
		upperBound := td.centroidUpperBound(i)
		if value < upperBound {
			// the value falls inside the bounds of this centroid
			// based on the assumed uniform distribution, we calculate how much
			// of this centroid's weight is below the value
			weightSoFar += c.Weight * (value - lowerBound) / (upperBound - lowerBound)
			return weightSoFar / td.mainWeight
		}

		// the value is above this centroid, so sum the weight and carry on
		weightSoFar += c.Weight
		lowerBound = upperBound
	}

	// should never be reached, since the final loop comparison is value < td.max
	return math.NaN()
}

// Returns a value such that the fraction of values in td below that value is
// approximately equal to quantile. Returns NaN if the digest is empty.
func (td *TDigest) Quantile(quantile float64) float64 {
	if quantile < 0 || quantile > 1 {
		panic("quantile out of bounds")
	}
	td.mergeTemps()

	// add up the weights of centroids in ascending order until we reach a
	// centroid that pushes us over the quantile
	q := quantile * td.mainWeight
	weightSoFar := 0.0
	lowerBound := td.min
	for i, c := range td.mainCentroids {
		upperBound := td.centroidUpperBound(i)
		if q <= weightSoFar+c.Weight {
			// the target quantile is somewhere inside this centroid
			// we compute how much of this centroid's weight falls into the quantile
			proportion := (q - weightSoFar) / c.Weight
			// and interpolate what value that corresponds to inside a uniform
			// distribution
			return lowerBound + (proportion * (upperBound - lowerBound))
		}

		// the quantile is above this centroid, so sum the weight and carry on
		weightSoFar += c.Weight
		lowerBound = upperBound
	}

	// should never be reached unless empty, since the final comparison is
	// q <= td.mainWeight
	return math.NaN()
}

func (td *TDigest) Min() float64 {
	return td.min
}
func (td *TDigest) Max() float64 {
	return td.max
}
func (td *TDigest) Count() float64 {
	return td.mainWeight + td.tempWeight
}

// we assume each centroid contains a uniform distribution of values
// the lower bound of the distribution is the midpoint between this centroid and
// the previous one (or the minimum, if this is the lowest centroid)
// similarly, the upper bound is the midpoint between this centroid and the
// next one (or the maximum, if this is the greatest centroid)
// this function returns the position of the upper bound (the lower bound is
// equal to the upper bound of the previous centroid)
// this assumption is justified empirically in dunning's paper
// TODO: does this assumption actually apply to our implementation?
func (td *TDigest) centroidUpperBound(i int) float64 {
	if i != len(td.mainCentroids)-1 {
		return (td.mainCentroids[i+1].Mean + td.mainCentroids[i].Mean) / 2
	} else {
		return td.max
	}
}

// Merge another digest into this one. Neither td nor other can be shared
// concurrently during the execution of this method.
func (td *TDigest) Merge(other *TDigest) {
	shuffledIndices := rand.Perm(len(other.mainCentroids))

	for _, i := range shuffledIndices {
		td.Add(other.mainCentroids[i].Mean, other.mainCentroids[i].Weight)
	}

	// we did not merge other's temps, so we need to add those too
	// they're unsorted so there's no need to shuffle them
	for i := range other.tempCentroids {
		td.Add(other.tempCentroids[i].Mean, other.tempCentroids[i].Weight)
	}
}

var _ gob.GobEncoder = &TDigest{}
var _ gob.GobDecoder = &TDigest{}

func (td *TDigest) GobEncode() ([]byte, error) {
	td.mergeTemps()

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(td.mainCentroids); err != nil {
		return nil, err
	}
	if err := enc.Encode(td.compression); err != nil {
		return nil, err
	}
	if err := enc.Encode(td.min); err != nil {
		return nil, err
	}
	if err := enc.Encode(td.max); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (td *TDigest) GobDecode(b []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(b))

	if err := dec.Decode(&td.mainCentroids); err != nil {
		return err
	}
	if err := dec.Decode(&td.compression); err != nil {
		return err
	}
	if err := dec.Decode(&td.min); err != nil {
		return err
	}
	if err := dec.Decode(&td.max); err != nil {
		return err
	}

	// reinitialize the remaining variables
	td.mainWeight = 0
	for _, c := range td.mainCentroids {
		td.mainWeight += c.Weight
	}
	td.tempWeight = 0
	if tempSize := estimateTempBuffer(td.compression); cap(td.tempCentroids) != tempSize {
		td.tempCentroids = make([]Centroid, 0, tempSize)
	} else {
		// discard any unmerged centroids if we didn't reallocate
		td.tempCentroids = td.tempCentroids[:0]
	}

	return nil
}
