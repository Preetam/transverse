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
var req = require("./req")
var User = require("./user")

var PrimaryNav = {
  view: function(vnode) {
    var user = User.get();

    var genLiItems = function() {
      var liItems = [
        m("li",
          m("a", {href: "#!/goals"}, "Goals")
        ),
        m("li",
          m("a", {href: "#!/profile"}, "Profile")
        )
      ]

      if (user.id) {
        liItems.push(
          m("li.tv-login-info", [
            "Logged in as ",
            m("strong", user.name),
            ". ",
            m("a[href='/logout']", "Log out")
          ])
        )
      }
      return liItems;
    }

    return [
      m("div.tv-menu-top", [
        m("h1.tv-site-logo", m("a[href='/app']", "Transverse")),
        m("div#hw-widget", ""),
        m("div",
          {
            class: "tv-hamburger",
            id: "tv-dropdown-hamburger",
            onclick: function() {
              document.getElementById("tv-dropdown-items").classList.toggle("active");
              document.getElementById("tv-dropdown-hamburger").classList.toggle("active");
            }
          },
          [
            m("div.tv-hamburger-bar1"),
            m("div.tv-hamburger-bar2")
          ]
        ),
        m("div.tv-menu-items", m("ul", genLiItems()))
      ]),
      m("div#tv-dropdown-items", m("ul", genLiItems()))
    ];
  }
};

module.exports = PrimaryNav;
