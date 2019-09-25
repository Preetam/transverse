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

 var m = require("mithril")

// Request wrapper
var req = function(opts) {
  opts.config = function(xhr) {
    xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");
  };
  opts.extract = function(xhr) {
    var result = {};
    try {
      result = JSON.parse(xhr.response);
    } catch (e) {
    }
    if (xhr.status > 299) {
      result.status = xhr.status;
      result.statusText = xhr.statusText;
      result.response = xhr.response;
    }
    return result;
  }
  return m.request(opts);
};

module.exports = req
