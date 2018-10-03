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

 var papaparse = require("papaparse")

function parseCSV(str) {
  str = str.trim();
  var events = [];

  var lines = str.split("\n")
  if (lines[0].startsWith("Activity")) {
    str = lines.slice(1).join("\n")
  }

  var parsed = papaparse.parse(str, {dynamicTyping: true});
  if (parsed.errors.length > 0) {
    throw "invalid";
  }

  var now = new Date();
  var normalized = new Date(now.getTime() - (now.getTimezoneOffset()*60*1000))

  for (var i in parsed.data) {
    var d = new Date(parsed.data[i][0]);
    var normalized = new Date(d.getTime() - (d.getTimezoneOffset()*60*1000))
    var event = {
      ts: normalized,
      value: parseFloat(parsed.data[i][1])
    }
    // Validate
    if (event.ts.toString() == "Invalid Date") {
      throw "invalid date"
    }
    events.push(event)
  }

  return events
}

var parse = {
  csv: parseCSV
}

module.exports = parse;
