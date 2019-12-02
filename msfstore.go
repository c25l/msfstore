package MSFStore

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
	"strconv"
)

// Histogram is, for now, a map store implementing something akin to the
// HDRHistogram idea, meaning it's high-accuracy and has few restrictions.
// It can be an issue for space, as it will get bigger as more keys are added.
type Histogram struct {
	// This will be used as an int, but it's a float to satisfy upstream
	// dependencies
	Resolution int
	// using strings as keys because I've used floats long enough
	// to not compare them for equality.
	Registers map[string]float64
}

// NewHistogram returns a new map store, not very exciting
func New(Resolution int) Histogram {
	var x Histogram
	s := make(map[string]float64)
	x.Resolution = Resolution
	x.Registers = s
	return x
}

// project a value for storage in the histogram
func (m Histogram) project(x float64) string {
	return fmt.Sprintf("%."+strconv.Itoa(m.Resolution)+"e", x)
}

// Insert a count at a certain value into a histogram
func (m Histogram) Insert(val float64, count float64) Histogram {
	dest := m.project(val)
	if m.Registers == nil {
		m.Registers = make(map[string]float64)
	}
	current, ok := m.Registers[dest]
	if !ok {
		m.Registers[dest] = 0.0
	}
	m.Registers[dest] = current + count
	return m
}

// Read a value out of a histogram
func (m Histogram) Read(val float64) float64 {
	dest := m.project(val)
	output := m.Registers[dest]
	return output
}

// Min returns the pointwise min of the two histograms.
func (m Histogram) Min(o Histogram) Histogram {
	out := New(m.Resolution)
	for key, val := range m.Registers {
		useval := val
		if val, ok := o.Registers[key]; ok {
			if val < useval {
				useval = val
			}
			out.Registers[key] = useval
		}
	}
	return out
}

// RawHist contains a gonum-compatible array pair.
type RawHist struct {
	Location, Weight []float64
}

// ToRawHist gives an interchangeable format for histograms as a pair of arrays of sorted values and weights.
func (m Histogram) ToRawHist() RawHist {
	locs := make([]float64, len(m.Registers))
	wts := make([]float64, len(m.Registers))
	index := 0
	for key := range m.Registers {
		parsed, err := strconv.ParseFloat(key, 64)
		if err != nil {
			continue
		}
		locs[index] = parsed
		index++
	}
	sort.Float64s(locs)
	for ii, xx := range locs {
		wts[ii] = m.Registers[m.project(xx)]
	}

	return RawHist{locs, wts}
}

// FromRawHist takes an interchange pair and regenerates a hist.
func FromRawHist(x RawHist, Resolution int) Histogram {
	m := New(Resolution)
	locs := x.Location
	wts := x.Weight
	for ii, xx := range locs {
		m.Insert(xx, wts[ii])
	}
	return m
}

// Combine adds two hists together.
func (m Histogram) Combine(x Histogram) Histogram {
	out := New(m.Resolution)
	for key, val := range x.Registers {
		usekey, _ := strconv.ParseFloat(key, 64)
		out.Insert(usekey, val)
	}
	for key, val := range m.Registers {
		usekey, _ := strconv.ParseFloat(key, 64)
		out.Insert(usekey, val)
	}
	return out
}

// Total gives the total of all elements in the histogram.
func (m Histogram) Total() float64 {
	output := 0.0
	for _, val := range m.Registers {
		output += val
	}
	return output
}

// Cancel subtracts the argument from the base. It's called cancel and not diff or
// subtract because enjoy having negative values in your hists if you're not careful.
func (m Histogram) Cancel(x Histogram) Histogram {
	out := New(m.Resolution)
	for key, val := range m.Registers {
		usekey, _ := strconv.ParseFloat(key, 64)
		out.Insert(usekey, val)
	}
	for key, val := range x.Registers {
		usekey, _ := strconv.ParseFloat(key, 64)
		out.Insert(usekey, -val)
	}
	return out
}

// Serialize turns a Histogram into bytes.
func (m Histogram) Serialize() []byte {
	var outbytes bytes.Buffer
	enc := gob.NewEncoder(&outbytes)
	_ = enc.Encode(m.ToRawHist())
	return outbytes.Bytes()
}

// Deserialize is the inverse of serialize.
func Deserialize(input []byte, res int) (Histogram, error) {
	var inbytes bytes.Buffer
	var h RawHist
	inbytes.Write(input)
	dec := gob.NewDecoder(&inbytes)
	err := dec.Decode(&h)
	out := FromRawHist(h, res)
	return out, err
}
