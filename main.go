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
		Sum    float64
		N      int
		Values []float64
	}

	statistics, count, previous := make(map[int]Statistic), 0, ""
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
				irradiance, err := strconv.ParseFloat(parts[MeasureTypeIrradiance], 64)
				if err != nil {
					panic(err)
				}
				statistic, found := statistics[int(minWavelength)]
				if !found {
					statistics[int(minWavelength)] = Statistic{}
				}
				statistic.Sum += irradiance
				statistic.N++
				for i := len(statistic.Values); i < count; i++ {
					statistic.Values = append(statistic.Values, irradiance)
				}
				statistics[int(minWavelength)] = statistic
			}
		}
		line, err = reader.ReadString('\n')
	}

	for minWavelength, statistic := range statistics {
		values = append(values, plotter.XY{X: float64(minWavelength), Y: statistic.Sum / float64(statistic.N)})
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

	for key, value := range statistics {
		frequency := fft.FFTReal(value.Values)
		values = values[:0]
		for i, f := range frequency {
			values = append(values, plotter.XY{X: float64(i), Y: cmplx.Abs(f)})
		}

		p = plot.New()
		p.Title.Text = "Spectrum"
		histogram, err = plotter.NewHistogram(values, len(values))
		if err != nil {
			panic(err)
		}
		p.Add(histogram)
		err = p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("frequency_%d.png", key))
		if err != nil {
			panic(err)
		}
	}
}
