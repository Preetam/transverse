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

var AddGoalForm = {
  oninit: function(vnode) {
    vnode.state.goal = new Goal;
  },
  view: function(vnode) {
    return m("form", {
      class: "pure-form pure-form-stacked",
      onsubmit: function() {
        vnode.attrs.create(vnode.state.goal);
        return false;
      }
    },
    m("fieldset", [
      m("div.pure-control-group", [
        m("label", {for: "form-goal-name"}, "Name"),
        m("input#form-goal-name", {
          class: "form-control",
          oninput: m.withAttr("value", vnode.state.goal.setName.bind(vnode.state.goal)),
          placeholder: "Example: My reading goal"
        }),
        m("div#form-goal-name-help", {class: "pure-form-message-inline"}, "Give your goal a name.")
      ]),
      m("div.pure-control-group", [
        m("label", {for: "form-goal-target"}, "Target"),
        m("input#form-goal-target", {
          class: "form-control",
          oninput: m.withAttr("value", vnode.state.goal.setTarget.bind(vnode.state.goal)),
          placeholder: "Example: 473"
        }),
        m("div#form-goal-target-help", {class: "pure-form-message-inline"},
          "Set a target value. This has to be a number. If you want to make it \"473 pages\", just put \"473\".")
      ]),
      m("button", {style: {marginTop: "10px"}, class: "pure-button"}, "Add goal")
    ]))
  }
}

var GoalAddPage = {
  oninit: function(vnode) {
    vnode.state.error = "";
    vnode.state.create = function(goal) {
      Goal.create(goal).then(function() {m.route.set("/goals/")}, function(e) {
        vnode.state.error = e.statusText;
      })
    }
  },
  view: function(vnode) {
    return m("div", [
      m("h1.tv-page-title", "Add a new goal"),
      m(AddGoalForm, {create: vnode.state.create}),
      (vnode.state.error != "" ?
        m("div", {class: "alert alert-warning", role: "alert", style: "margin: 1rem 0"}, vnode.state.error) :
        "")
    ])
  }
}

module.exports = GoalAddPage;
