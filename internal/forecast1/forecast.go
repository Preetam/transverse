// Package forecast contains version 1 forecasting code.
package forecast1

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

import "math"

const alpha = 0.3
const beta = 0.3

type Model struct {
	Steps     int
	Level     float32
	Trend     float32
	Forecasts []float32
	ErrEWMA   float32
	first     bool
}

func NewModel(steps int) *Model {
	forecasts := make([]float32, steps)
	for i := range forecasts {
		forecasts[i] = float32(math.NaN())
	}
	return &Model{
		Steps:     steps,
		Forecasts: forecasts,
		Level:     float32(math.NaN()),
		Trend:     float32(math.NaN()),
	}
}

func (m *Model) Update(v, level, trend float32) {
	if math.IsNaN(float64(m.Level)) {
		m.Level = level
	}
	if math.IsNaN(float64(m.Trend)) {
		m.Trend = trend
	}
	if !math.IsNaN(float64(m.Forecasts[0])) {
		m.ErrEWMA = (m.Forecasts[0]-v)*(m.Forecasts[0]-v)*alpha + (1-alpha)*m.ErrEWMA
	}
	if m.Steps == 1 {
		newLevel := v*alpha + (1-alpha)*(m.Level+m.Trend)
		newTrend := (newLevel-m.Level)*beta + (1-beta)*m.Trend
		m.Forecasts[0] = newLevel + newTrend
		level = newLevel
		trend = newTrend
		m.Level = level
		m.Trend = trend
	}
	m.Forecasts = m.Forecasts[1:]
	m.Forecasts = append(m.Forecasts, level+float32(m.Steps)*trend)
}

func (m *Model) Forecast() float32 {
	return m.Forecasts[m.Steps-1]
}

type Range [2]float64

func (r Range) Subdivide(parts int) []float64 {
	result := []float64{}
	intervalSize := (r[1] - r[0]) / float64(parts)
	for i := 0; i <= parts; i++ {
		result = append(result, r[0]+float64(i)*intervalSize)
	}
	return result
}
