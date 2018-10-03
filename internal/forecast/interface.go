package forecast

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

import "fmt"

// ForecastPoint is a single forecasted point.
type ForecastPoint struct {
	Low       float64
	Predicted float64
	High      float64
}

type Forecaster interface {
	Forecast(data []float64) []ForecastPoint
}

func (p ForecastPoint) String() string {
	return fmt.Sprintf("(%0.3f, %0.3f, %0.3f)", p.Low, p.Predicted, p.High)
}
