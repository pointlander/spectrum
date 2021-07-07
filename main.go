// Copyright 2021 The Spectrum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"math/cmplx"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mjibson/go-dsp/fft"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// MeasureType is a type of measure
type MeasureType int

const (
	// MeasureTypeDate is a date
	MeasureTypeDate MeasureType = iota
	// MeasureTypeDateJDN is a jdn date
	MeasureTypeDateJDN
	// MeasureTypeMinWavelength is the min wavelength
	MeasureTypeMinWavelength
	// MeasureTypeMaxWavelength is the max wavelength
	MeasureTypeMaxWavelength
	// MeasureTypeInstrumentModeID is the instrument mode id
	MeasureTypeInstrumentModeID
	// MeasureTypeDataVersion is the data version
	MeasureTypeDataVersion
	// MeasureTypeIrradiance is the irradiance
	MeasureTypeIrradiance
	// MeasureTypeIrradianceUncertainty is the irradiance uncertainty
	MeasureTypeIrradianceUncertainty
	// MeasureTypeQuality is the quality
	MeasureTypeQuality
	// MeasureTypeCount is the number of measures
	MeasureTypeCount
)

func main() {
	space := regexp.MustCompile(`[\s]+`)

	input, err := os.Open("sorce_L3_combined_c24h_20030225_20200225.txt")
	if err != nil {
		panic(err)
	}
	defer input.Close()

	type Statistic struct {
		Sum           float64
		N             int
		Values        []float64
		MinWavelength float64
		MaxWavelength float64
	}

	statistics, count, previous := make(map[string]Statistic), 0, ""
	values := make(plotter.XYs, 0, 1024)
	reader := bufio.NewReader(input)
	line, err := reader.ReadString('\n')
	for err == nil {
		if !strings.HasPrefix(line, ";") {
			parts := space.Split(strings.TrimSpace(line), -1)
			if len(parts) == int(MeasureTypeCount) {
				if previous != parts[MeasureTypeDate] {
					previous = parts[MeasureTypeDate]
					count++
				}

				minWavelength, err := strconv.ParseFloat(parts[MeasureTypeMinWavelength], 64)
				if err != nil {
					panic(err)
				}
				maxWavelength, err := strconv.ParseFloat(parts[MeasureTypeMaxWavelength], 64)
				if err != nil {
					panic(err)
				}
				irradiance, err := strconv.ParseFloat(parts[MeasureTypeIrradiance], 64)
				if err != nil {
					panic(err)
				}
				key := fmt.Sprintf("%f-%f", minWavelength, maxWavelength)
				statistic, found := statistics[key]
				if !found {
					statistics[key] = Statistic{}
					statistic.MinWavelength = minWavelength
					statistic.MaxWavelength = maxWavelength
				}
				statistic.Sum += irradiance
				statistic.N++
				for i := len(statistic.Values); i < count; i++ {
					statistic.Values = append(statistic.Values, irradiance)
				}
				statistics[key] = statistic
			}
		}
		line, err = reader.ReadString('\n')
	}

	for key, statistic := range statistics {
		length := len(statistic.Values)
		for i := length; i < count; i++ {
			statistic.Values = append(statistic.Values, statistic.Values[length-1])
		}
		statistics[key] = statistic
		fmt.Println(key, statistic.N, len(statistic.Values))
		values = append(values, plotter.XY{X: float64(statistic.MinWavelength), Y: statistic.Sum / float64(statistic.N)})
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].X < values[j].X
	})
	fmt.Println(len(values))

	p := plot.New()
	p.Title.Text = "Spectrum"
	histogram, err := plotter.NewHistogram(values, len(values))
	if err != nil {
		panic(err)
	}
	p.Add(histogram)
	err = p.Save(8*vg.Inch, 8*vg.Inch, "spectrum.png")
	if err != nil {
		panic(err)
	}

	type Value struct {
		Index int
		Value float64
	}

	ranks, length := make(map[string][]Value), len(statistics["0.000000-1.000000"].Values)
	sum := make([]float64, length)
	for key, value := range statistics {
		frequency := fft.FFTReal(value.Values)
		values = values[:0]
		frequencies := make([]Value, len(frequency))
		for i, f := range frequency {
			values = append(values, plotter.XY{X: float64(i), Y: cmplx.Abs(f)})
			frequencies[i].Index = i
			frequencies[i].Value = cmplx.Abs(f)
			sum[i] += cmplx.Abs(f)
		}
		sort.Slice(frequencies, func(i, j int) bool {
			return frequencies[i].Value > frequencies[j].Value
		})
		ranks[key] = frequencies

		p = plot.New()
		p.Title.Text = "Spectrum"
		histogram, err = plotter.NewHistogram(values, len(values))
		if err != nil {
			panic(err)
		}
		p.Add(histogram)
		err = p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("plots/frequency_%s.png", key))
		if err != nil {
			panic(err)
		}
	}

	averages := make([]float64, length)
	for _, value := range ranks {
		for i, rank := range value {
			averages[rank.Index] += float64(i)
		}
	}

	v := make([]Value, 0, 1024)
	values = values[:0]
	for i, f := range sum {
		values = append(values, plotter.XY{X: float64(i), Y: f / float64(len(sum))})
		v = append(v, Value{
			Index: i,
			Value: f / float64(len(sum)),
		})
	}

	p = plot.New()
	p.Title.Text = "Spectrum"
	histogram, err = plotter.NewHistogram(values, len(values))
	if err != nil {
		panic(err)
	}
	p.Add(histogram)
	err = p.Save(8*vg.Inch, 8*vg.Inch, "frequency.png")
	if err != nil {
		panic(err)
	}

	sort.Slice(v, func(i, j int) bool {
		return v[i].Value < v[j].Value
	})
	for _, value := range v {
		fmt.Println(value.Index, value.Value)
	}

	v = v[:0]
	for i, average := range averages {
		v = append(v, Value{
			Index: i,
			Value: average / float64(length),
		})
	}
	sort.Slice(v, func(i, j int) bool {
		return v[i].Value > v[j].Value
	})
	for _, value := range v {
		fmt.Println(value.Index, value.Value)
	}
}
