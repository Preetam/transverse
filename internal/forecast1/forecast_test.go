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

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestForecast(t *testing.T) {
	models := []*Model{}
	for i := 1; i <= 50; i++ {
		models = append(models, NewModel(i))
	}

	for i := 0; i < 350; i++ {
		val := 10*rand.Float32() + float32(i)*0.01*float32(math.Sin(float64(i)))
		fmt.Println(i+1, val)
		if i > 300 {
			continue
		}
		models[0].Update(val, 0, 0)
		for j, m := range models {
			if j == 0 {
				continue
			}
			m.Update(val, models[0].Level, models[0].Trend)
		}
	}

	for _, m := range models {
		prediction := m.Forecast()
		sqrtErr := float32(math.Sqrt(float64(m.ErrEWMA)))
		low := prediction - 2*sqrtErr
		high := prediction + 2*sqrtErr
		_, _ = low, high
		fmt.Println(low, prediction, high)
	}
}

func TestRange(t *testing.T) {
	r := Range{0, 1}
	t.Log(r.Subdivide(2))
}
