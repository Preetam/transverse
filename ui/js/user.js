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

 // user

var m = require("mithril")
var req = require("./req")

var User = function() {
  this.data = {}
  this.get = function() {
    return this.data
  }.bind(this)
  this.refresh = function() {
    return req({
      method: "GET",
      url: "/api/v1/user",
    }).then(function(d) {
      this.data = d.data;
    }.bind(this));
  }.bind(this)
  this.refresh()
};

User.updatePassword = function(password) {
  return req({
    method: "PUT",
    url: "/api/v1/user/password",
    data: password
  })
}

User.delete = function() {
  return req({
    method: "DELETE",
    url: "/api/v1/user"
  })
}

var user = new User();
user.updatePassword = User.updatePassword;
user.delete = User.delete;

module.exports = user;
