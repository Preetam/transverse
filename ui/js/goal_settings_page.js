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
var Goal = require("./goals").Goal
var Spinner = require("./spinner")

var GoalSettingsPage = {
  oninit: function (vnode) {
    vnode.state.goal = new Goal({id: m.route.param("id")});
    vnode.state.goal.error = "";
    Goal.get(vnode.state.goal).then(function() {
      vnode.state.goalUpdateForm = m(UpdateGoalForm, {
        goal: vnode.state.goal,
        update: function (goal) {
          Goal.update(goal).then(function () { location.reload() })
        }
      })
    }).catch(function(e) {
      if (e.code === 404) {
        return;
      }
      if (e.code) {
        vnode.state.goal.error = "Status " + e.code + " " + e.message + "\n";
        vnode.state.goal.error += e.message;
        if (e.response) {
          vnode.state.goal.error += e.response;
        }
      }
    })
  },

  view: function (vnode) {
    if (vnode.state.goal.user === "" && vnode.state.goal.error === "") {
      return m(Spinner)
    }
    var archiveAction = (vnode.state.goal.archived ? "unarchive" : "archive");
    var archiveActionButtonText = (vnode.state.goal.archived ? "Unarchive" : "Archive");
    if (vnode.state.goal.error === "" && vnode.state.goal.user === "") {
      return m(Spinner)
    }
    return m("div", [
      m("h1.tv-goal-title", [
        vnode.state.goal.name,
        " ",
        m("a.tv-settings-back-link", {
          href: "#!/goals/"+vnode.state.goal.id,
          style: {
            float: "right",
          }
        }, "Back")]),
      m("div.tv-goal-updated", "Editing goal settings"),
      m("div", { style: "margin-bottom: 1rem;" }, vnode.state.goalUpdateForm),
      m("hr.tv-section-separator"),
      m("form", {
        class: "pure-form pure-form-stacked"
      },
        [
          m("button", {
            class: "pure-button tv-archive-button",
            onclick: function (e) {
              if (confirm("Are you sure you want to " + archiveAction + " this goal?")) {
                if (!vnode.state.goal.archived) {
                  Goal.archive(vnode.state.goal).then(function () { m.route.set("/goals/") })
                } else {
                  Goal.unarchive(vnode.state.goal).then(function () { m.route.set("/goals/") })
                }
              }
              return false;
            }
          }, archiveActionButtonText + " this goal"),
          " ",
          m("button", {
            class: "pure-button tv-delete-button",
            onclick: function (e) {
              if (confirm("Are you sure you want to delete this goal?")) {
                Goal.delete(vnode.state.goal).then(function () { m.route.set("/goals/") })
              }
              return false;
            }
          }, "Delete this goal")
        ])
    ])
  }
}


var UpdateGoalForm = {
  oninit: function(vnode) {
    vnode.state.goal = vnode.attrs.goal;
  },
  view: function(vnode) {
    return m("form", {
      class: "pure-form pure-form-stacked",
      onsubmit: function() {
        vnode.attrs.update(vnode.state.goal);
        return false;
      }
    },
    m("fieldset", [
      m("div.pure-control-group", [
        m("label", {for: "form-goal-name"}, "Name"),
        m("input#form-goal-name", {
          class: "form-control",
          onchange: function(ev) {vnode.state.goal.name = ev.target.value},
          value: vnode.state.goal.name
        }),
        m("div#form-goal-name-help", {class: "pure-form-message-inline"}, "Give your goal a name.")
      ]),
      m("div.pure-control-group", [
        m("label", {for: "form-goal-target"}, "Target"),
        m("input#form-goal-target", {
          class: "form-control",
          onchange: function(ev) {vnode.state.goal.target = parseFloat(ev.target.value)},
          value: vnode.state.goal.target
        }),
        m("div#form-goal-target-help", {class: "pure-form-message-inline"},
          "Set a target value. This has to be a number. If you want to make it \"473 pages\", just put \"473\".")
      ]),
      m("button", {style: {marginTop: "10px"}, class: "pure-button"}, "Update goal")
    ]))
  }
}

module.exports = GoalSettingsPage;
