package forecast3

/**
 * Copyright (C) 2018 Preetam Jinka
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"math"

	"github.com/Preetam/transverse/internal/forecast"
)

type Forecaster struct {
}

func NewForecaster() *Forecaster {
	return &Forecaster{}
}

func (f *Forecaster) Forecast(data []float64) []forecast.ForecastPoint {
	result := []forecast.ForecastPoint{}

	maxInitPoints := len(data)
	if maxInitPoints > 3 && false {
		maxInitPoints = 3
	}

	if len(data) > 60 {
		data = data[len(data)-60 : len(data)]
	}

	// Find the best double exponential smoothing model.
	var minModel *simpleModel
	for alpha := 0.001; alpha < 0.3; alpha += 0.001 {
		for beta := 0.001; beta < 0.3; beta += 0.001 {
			model := newSimpleModel(alpha, beta)
			model.Initialize(data[:maxInitPoints])
			model.level = data[0]
			for _, v := range data {
				model.AddPoint(v)
			}
			if minModel == nil {
				minModel = model
			}
			if minModel.SquareError() > model.SquareError() {
				minModel = model
			}
		}
	}

	slope := minModel.trend
	lastPointValue := data[len(data)-1]

	stepsToForecastAhead := len(data) - 1
	if stepsToForecastAhead > 30 {
		stepsToForecastAhead = 30
	}

	prevErrBelow := 0.0
	prevErrAbove := 0.0
	for nStep := 1; nStep <= stepsToForecastAhead; nStep++ {
		errorAbove := 0.0
		errorBelow := 0.0
		pointsForecasted := 0.0

		forecastPointsForErr := make([]float64, nStep, nStep)
		// Initialize forecastPointsForErr to NaN.
		for i := range forecastPointsForErr {
			forecastPointsForErr[i] = math.NaN()
		}

		forecastPointsForCoverage := map[int]float64{}

		model := newSimpleModel(minModel.alpha, minModel.beta)
		model.Initialize(data[:maxInitPoints])
		model.level = data[0]

		monotonicIncreasing := true
		monotonicDecreasing := true
		for i, point := range data {
			if i >= 2 {
				if data[i] > data[i-1] {
					monotonicDecreasing = false
				}
				if data[i] < data[i-1] {
					monotonicIncreasing = false
				}
			}
			if !math.IsNaN(forecastPointsForErr[0]) {
				// Compare our forecast for this point with the true value.
				predicted := forecastPointsForErr[0]
				predictedDiff := predicted - point
				if point > predicted {
					// new point is above prediction
					errorAbove += math.Abs(predictedDiff)
				} else {
					errorBelow += math.Abs(predictedDiff)
				}
				pointsForecasted++
				forecastPointsForCoverage[i] = predicted
			}
			model.AddPoint(point)
			forecastPointsForErr = forecastPointsForErr[1:]
			forecastPointsForErr = append(forecastPointsForErr, model.Forecast(nStep))
		}

		avgErrAbove := errorAbove / pointsForecasted
		avgErrBelow := errorBelow / pointsForecasted

		var multiplierBelow float64
		for multiplierBelow = 1.0; multiplierBelow < 3; multiplierBelow += 0.001 {
			covered := []float64{}
			for i, predicted := range forecastPointsForCoverage {
				if i >= len(data) {
					continue
				}
				value := data[i]
				low := predicted - multiplierBelow*avgErrBelow

				if value > low {
					covered = append(covered, 1)
				} else {
					covered = append(covered, 0)
				}
			}

			if sum(covered)/float64(len(covered)) >= 0.9 {
				// Have >90% coverage
				break
			}

			if monotonicIncreasing {
				predicted := lastPointValue + float64(nStep-1)*slope
				errBelow := multiplierBelow * avgErrBelow
				if errBelow < prevErrBelow {
					errBelow = prevErrBelow
				}
				if nStep > 1 {
					if predicted-errBelow < result[len(result)-1].Low {
						break
					}
				} else if predicted-errBelow < lastPointValue {
					break
				}
			}
		}

		var multiplierAbove float64
		for multiplierAbove = 1.0; multiplierAbove < 3; multiplierAbove += 0.001 {
			covered := []float64{}
			for i, predicted := range forecastPointsForCoverage {
				if i >= len(data) {
					continue
				}
				value := data[i]
				high := predicted + multiplierAbove*avgErrAbove

				if value < high {
					covered = append(covered, 1)
				} else {
					covered = append(covered, 0)
				}
			}

			if sum(covered)/float64(len(covered)) >= 0.9 {
				// Have >90% coverage
				break
			}

			if monotonicDecreasing {
				predicted := lastPointValue + float64(nStep-1)*slope
				errAbove := multiplierAbove * avgErrAbove
				if errAbove < prevErrAbove {
					errAbove = prevErrAbove
				}
				if nStep > 1 {
					if predicted+errAbove > result[len(result)-1].High {
						break
					}
				} else if predicted+errAbove > lastPointValue {
					break
				}
			}
		}

		// Add the nth step ahead forecast (finally).
		predicted := lastPointValue + float64(nStep-1)*slope

		errBelow := multiplierBelow * avgErrBelow
		errAbove := multiplierAbove * avgErrAbove

		if errBelow < prevErrBelow {
			errBelow = prevErrBelow
		}
		if errAbove < prevErrAbove {
			errAbove = prevErrAbove
		}
		prevErrBelow = errBelow
		prevErrAbove = errAbove
		if nStep == 1 {
			errBelow = 0
			errAbove = 0
		}
		result = append(result, forecast.ForecastPoint{
			Predicted: predicted,
			Low:       predicted - errBelow,
			High:      predicted + errAbove,
		})
	}

	return result
}

var _ forecast.Forecaster = &Forecaster{}
