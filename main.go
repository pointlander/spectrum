// Copyright 2021 The Spectrum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

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

	values := make(plotter.XYs, 0, 1024)
	reader := bufio.NewReader(input)
	line, err := reader.ReadString('\n')
	for err == nil {
		if !strings.HasPrefix(line, ";") {
			parts := space.Split(strings.TrimSpace(line), -1)
			if len(parts) == int(MeasureTypeCount) {
				if parts[MeasureTypeDate] == "20030225.5" {
					minWavelength, err := strconv.ParseFloat(parts[MeasureTypeMinWavelength], 64)
					if err != nil {
						panic(err)
					}
					irradiance, err := strconv.ParseFloat(parts[MeasureTypeIrradiance], 64)
					if err != nil {
						panic(err)
					}
					fmt.Println(minWavelength, irradiance)
					values = append(values, plotter.XY{X: minWavelength, Y: irradiance})
				}
			}
		}
		line, err = reader.ReadString('\n')
	}

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
}
