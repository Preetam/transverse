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

 // Profile page
var m = require("mithril");
var User = require("./user");

var ProfilePage = {
  oninit: function(vnode) {
    vnode.state.user = User.get();

    vnode.state.submit = (function() {
      var state = this;
      this.error = "";
      this.updated = false;
      User.updatePassword(this.newPassword).then(function() {
        state.updated = true;
      }).catch(function(e) {
        state.updated = false;
        if (e.status === 400) { // bad request
          state.error = "Invalid password."
          return false;
        } else {
          state.error = "Something went wrong."
        }
      });
      return false;
    }).bind(vnode.state);

    vnode.state.newPassword = "";

    vnode.state.error = "";
    vnode.state.updated = false;
  },
  view: function(vnode) {
    var updated = (vnode.state.updated ? m("p", "Updated!") : "");
    var err = (vnode.state.error ? m("p", vnode.state.error) : "");
    return m("div", [
      m("h1.tv-page-title", "Profile"),
      m("form", {class: "pure-form pure-form-stacked", style: {width: "400px"}}, [
        m("fieldset", [
          m("label", {for: "email"}, "Email address"),
          m("input",
            {
              type: "email",
              disabled: true,
              name: "email",
              value: vnode.state.user.email
            }
          ),
          m("label", {for: "password1"}, "Password"),
          m("input",
            {
              type: "password",
              oninput: m.withAttr("value", function(v) { vnode.state.newPassword = v; }),
              name: "password1"
            },
            "Password"
          ),
          m("button", {style: {marginTop: "10px"}, class: "pure-button", onclick: vnode.state.submit}, "Update password")
        ]),
      ]),
      m("form", {
        class: "pure-form pure-form-stacked",
        style: {width: "400px"},
        action: "/api/v1/user/data",
        method: "get",
        target: "_blank"
        }, [
        m("fieldset", [
          m("button", {
            "type": "submit",
            style: {marginTop: "10px"},
            class: "pure-button",
          }, "Download data")
        ]),
      ]),
      m("form", {class: "pure-form pure-form-stacked", style: {width: "400px"}}, [
        m("fieldset", [
          m("button", {
            style: {marginTop: "10px"},
            class: "pure-button tv-delete-button",
            onclick: function(e) {
              if (confirm("Are you sure you want to delete this account? This can't be undone!")) {
                User.delete().then(function() {m.route.set("/logout")})
              }
              return false;
            }
          }, "Delete account")
        ]),
      ]),
      m("div", [
        updated,
        err
      ])
    ])
  }
}

module.exports = {
  ProfilePage: ProfilePage
}
