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

var ChartState = function(width, height, start, end, data, name, brushEnd) {
  this.width = width;
  this.height = height;
  this.start = start;
  this.end = end;
  this.data = data;
  this.name = name;
  this.maxVal = 10;
  this.minVal = 1e100;
  this.brushEnd = brushEnd
  for (var i in data.lines) {
    var lineData = data.lines[i];
    if (lineData.length == 1) {
      // Only one point, so skip it.
      continue
    }
    for (var j in lineData) {
      if (lineData[j].value > this.maxVal) {
        this.maxVal = lineData[j].value;
      }
      if (lineData[j].value < this.minVal) {
        this.minVal = lineData[j].value;
      }
    }
  }
};

module.exports = ChartState;
