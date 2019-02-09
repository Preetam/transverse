/**
 * Copyright (C) 2019 Preetam Jinka
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
var moment = require("moment")
var Goals = require("./goals").Goals

var NoGoalsNotice = {
  view: function(vnode) {
    return m("div", {style: "text-align: center;"}, [
      m("p", "You haven’t added any goals yet."),
      m("a", {class: "pure-button", style: "margin: 0.5rem", href: "/create-goal", oncreate: m.route.link}, "Add a goal")
    ])
  }
}

class ActiveGoalToggle {
  constructor(cb) {
    this.activeOnly = true;
    this.oninit = function (vnode) {
      vnode.state.activeOnly = true;
      vnode.state.callback = cb;
    };
    this.view = function(vnode) {
      return m("div.tv-goals-list-filter", [
        "Showing ",
        m("a", {
          onclick: function() {
            vnode.state.activeOnly = !vnode.state.activeOnly;
            vnode.state.callback(vnode.state.activeOnly);
          },
          style: {
            fontWeight: "bold",
            cursor: "pointer"
          }
        }, vnode.state.activeOnly ? "active" : "all"),
        " goals."
      ])
    }
  }
}

var GoalsListPage = {
  oninit: function(vnode) {
    vnode.state.goals = new Goals;
    vnode.state.archived = false;
    vnode.state.loading = true;
    vnode.state.error = "";
    vnode.state.activeToggle = new ActiveGoalToggle(function(activeOnly) {
      Goals.get(vnode.state.goals, !activeOnly).then(function() {vnode.state.loading = false}).catch(function(e) {
        vnode.state.loading = false;
      })
    });
    Goals.get(vnode.state.goals, vnode.state.archived).then(function() {vnode.state.loading = false}).catch(function(e) {
      vnode.state.loading = false;
    })
  },
  view: function(vnode) {
    if (vnode.state.loading) {
      return m("p", "Loading...")
    }
    if (vnode.state.error != "") {
      return m("div", [
        m("p", "Oops. Something went wrong."),
        m("p", "Nerdy details:"),
        m("pre", vnode.state.goal.error)
      ])
    }
    if (!vnode.state.goals.data || Object.keys(vnode.state.goals.data).length === 0) {
      return m("div", [
        m("h1.tv-page-title", "Goals"),
        m(NoGoalsNotice)
      ]);
    }
    return m("div", [
      m("h1.tv-page-title", "Goals"),
      m(vnode.state.activeToggle),
      m("table", {class: "table table-sm goals-list-table"}, [
        m("tbody", [
          (function() {
            var goals = [];
            for (var i in vnode.state.goals.data) {
              goals.push(this.goals.data[i]);
            }
            goals.sort(function(a, b) {
              if (a.name < b.name) {
                return -1;
              }
              if (a.name > b.name) {
                return 1;
              }
              return 0;
            })
            var rows = [];
            for (var i in goals) {
              var goal = goals[i];
              rows.push(m("tr", [
                m("td", m("div.tv-goal-list-item", [
                  m("div", m("a.tv-goal-list-name[href=/goals/"+goal.id+"]", {oncreate: m.route.link}, goal.name + (goal.archived ? " (archived)" : ""))),
                  m("div.tv-list-created-updated", "Updated ", m("strong", moment(""+goal.updated, "X").fromNow())),
                  m("div.tv-list-created-updated", "Created ", m("strong", moment(""+goal.created, "X").fromNow()))
                ]))
              ]))
            }
            return rows;
          }).bind(this)()
        ])
      ]),
      m("a[href=/create-goal]", {class: "pure-button tv-add-another-goal-button", oncreate: m.route.link}, "Add another goal")
    ])
  }
}

module.exports = GoalsListPage;
