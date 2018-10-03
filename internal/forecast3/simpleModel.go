// Package forecast3 contains version 3 forecasting code.
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

// simpleModel is a simple EWMA model.
type simpleModel struct {
	alpha float64
	beta  float64

	level float64
	trend float64

	errorScore float64
}

func newSimpleModel(alpha, beta float64) *simpleModel {
	return &simpleModel{
		alpha: alpha,
		beta:  beta,
	}
}

// Initialize initializes level and trend based on the first few points.
func (m *simpleModel) Initialize(initData []float64) {
	m.level = sum(initData) / float64(len(initData))
	if len(initData) > 1 {
		m.trend = sum(diff(initData)) /
			float64(len(initData)-1)
	}
}

func (m *simpleModel) AddPoint(point float64) {
	// Get the 1-step-ahead forecast
	predicted := m.Forecast(1)
	predictedDiff := predicted - point
	m.errorScore += predictedDiff * predictedDiff

	newLevel := point*m.alpha + (1-m.alpha)*(m.level+m.trend)
	newTrend := (newLevel-m.level)*m.beta + (1-m.beta)*m.trend

	m.level = newLevel
	m.trend = newTrend
}

// Forecast returns an n-step-ahead forecast for a double exponential smoothing model.
func (m *simpleModel) Forecast(steps int) float64 {
	return m.level + float64(steps)*m.trend
}

// SquareError returns the EWMA square error for the model.
func (m *simpleModel) SquareError() float64 {
	return m.errorScore
}

func diff(vals []float64) []float64 {
	differences := []float64{}
	for i := 1; i < len(vals); i++ {
		differences = append(differences, vals[i]-vals[i-1])
	}
	return differences
}

func sum(vals []float64) float64 {
	result := 0.0
	for _, v := range vals {
		result += v
	}
	return result
}
