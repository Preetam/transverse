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
var parse = require("./parse")
var Chart = require("./chart")
var ChartState = require("./chartState")
var Goal = require("./goals").Goal
var Spinner = require("./spinner")

function getRelativeTimeString(d) {
  var now = new Date();
  if (Math.floor((now.getTime() - now.getTimezoneOffset()*60*1000)/(86400*1000)) ==
    Math.floor((d.getTime() - d.getTimezoneOffset()*60*1000)/(86400*1000))) {
    return "today";
  }
  return moment(d).fromNow();
}

var SummaryTable = function(goal) {
  this.view = function() {
    var components = [];

    components.push(m("div", [
      m("span.tv-summary-label", "last value:"),
      " ",
      m("span.summary-text", goal.end.value + " from " + getRelativeTimeString(goal.end.ts))
    ]));

    components.push(m("div", [
      m("span.tv-summary-label", "min:"),
      " ",
      m("span.summary-text", goal.min.value + " from " + getRelativeTimeString(goal.min.ts))
    ]));

    components.push(m("div", [
      m("span.tv-summary-label", "max:"),
      " ",
      m("span.summary-text", goal.max.value + " from " + getRelativeTimeString(goal.max.ts))
    ]));

    if (goal.data.series.length > 1) {
      components.push(m("div", [
        m("span.tv-summary-label", "duration:"),
        " ",
        m("span.summary-text", Math.floor((goal.end.ts.getTime() - goal.data_start.getTime())/(86400*1000)) + " days")
      ]));
    }

    if (goal.slope) {
      components.push(m("div", [
        m("span.tv-summary-label", "current rate:"),
        " ",
        m("span.summary-text", goal.slope.toFixed(2) + " per day")
      ]));
    }

    if (goal.eta) {
      components.push(m("div", [
        m("span.tv-summary-label", "eta:"),
        " ",
        m("span.summary-text", "~" + goal.eta + " days " + "(" + moment().add(goal.eta, "days").format("LL") + ")")
      ]));
    }
    return m("div.summary-table", [
      components
    ])
  }
}

var GoalDetailsPage = {
  oninit: function(vnode) {
    vnode.state.goal = new Goal({id: m.route.param("id")});
    vnode.state.goal.error = "";
    Goal.get(vnode.state.goal).then(function() {
      vnode.state.goalAddDataPage = new GoalAddDataPage(vnode.state.goal);
      return Goal.getData(vnode.state.goal)
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
  view: function(vnode) {
    if (vnode.state.goal.user === "" && vnode.state.goal.error === "") {
      return m(Spinner)
    }
    if (vnode.state.goal.error != "") {
      return m("div", [
        m("p", "Oops. Something went wrong."),
        m("p", "Nerdy details:"),
        m("pre", vnode.state.goal.error)
      ])
    }

    if (!vnode.state.goal.data) {
      // No data
      return m("div", [
        m("h1.tv-goal-title", vnode.state.goal.name),
        vnode.state.goal.updated > 0 ? m("div.tv-goal-updated", [
          "Updated ",
          m("strong", moment(""+vnode.state.goal.updated, "X").fromNow())
        ]) : "",
        m("div.row", [
          m("div.col-sm-9", {style: "padding-top: 0.5rem;"}, [
            m("h3.tv-goal-details-section-title", "Summary"),
            m("p", "You haven't added any data yet.")
          ])
        ]),
        m("hr.tv-section-separator"),
        m("div.row", [
          m(vnode.state.goalAddDataPage)
        ]),
        m("hr.tv-section-separator"),
        m("div.row", [
          m(new GoalSettingsPage(vnode.state.goal))
        ])
      ])
    }

    return m("div", [
      m("h1.tv-goal-title", vnode.state.goal.name),
      vnode.state.goal.updated > 0 ? m("div.tv-goal-updated", [
        "Updated ",
        m("strong", moment(""+vnode.state.goal.updated, "X").fromNow())
      ]) : "",
      m("div.row", [
        m("div.col-sm-9", {style: "padding-top: 0.5rem;"}, [
          m("h3.tv-goal-details-section-title", "Summary"),
          m(new Chart(new ChartState(100, 200, vnode.state.goal.data_start, vnode.state.goal.data_end, {
            lines: {
              "series": vnode.state.goal.data.series,
              "target": [{ts: vnode.state.goal.data_start, value: vnode.state.goal.target},{ts: vnode.state.goal.data_end, value: vnode.state.goal.target}],
              "prediction": vnode.state.goal.data.prediction,
              "low": vnode.state.goal.data.low,
              "high": vnode.state.goal.data.high
            }
          }, "line", function(){}))),
          m(new SummaryTable(vnode.state.goal))
        ])
      ]),
      m("hr.tv-section-separator"),
      m("div.row", [
        m(vnode.state.goalAddDataPage)
      ]),
      m("hr.tv-section-separator"),
      m("div.row", [
        m(new GoalSettingsPage(vnode.state.goal))
      ])
    ])
  }
}

// Add data page

var AddOrSetForm = function(goal) {
  this.oninit = function(vnode) {
    vnode.state.newValue = 0.0;
    vnode.state.goal = goal;
    vnode.state.submitError = "";
    vnode.state.add = (function() {
      var state = this;
      Goal.addDataAddPoint(this.goal, this.newValue)
        .then(function() {location.reload()}).catch(function(e) {
          state.submitError = "Something went wrong.";
        })
    }).bind(vnode.state);
    vnode.state.set = (function() {
      var state = this;
      Goal.addDataSetPoint(this.goal, this.newValue)
        .then(function() {location.reload()}).catch(function(e) {
          state.submitError = "Something went wrong.";
        })
    }).bind(vnode.state);
  }
  this.view = function(vnode) {
    return m("form", {
        class: "pure-form pure-form-stacked"
      },
      [
      m("div.form-group", [
        m("label", {for: "value"}, "Value"),
        m("input.form-control",
          {
            type: "number",
            pattern: "[0-9]+([.][0-9]+)?",
            oninput: function(ev) { vnode.state.newValue = parseFloat(ev.target.value); },
            name: "value",
            style: "width: 100px"
          })
      ]),
      m("button", {type: "button", class: "pure-button", onclick: vnode.state.add}, "Add"),
      " ",
      m("button", {type: "button", class: "pure-button", onclick: vnode.state.set}, "Set"),
      (vnode.state.submitError != "" ? (m("div", {class: "alert alert-warning", role: "alert", style: "margin: 1rem 0;"}, vnode.state.submitError)) : "")
    ])
  }
}

var GoalAddDataPage = function(goal) {
  this.oninit = function(vnode) {
    vnode.state.goal = goal;
    Goal.getRawData(vnode.state.goal).catch(function(e) {
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
    }).then(function() {
      vnode.state.loaded = false;
    })
    vnode.state.parsedEvents = [];
    vnode.state.submitError = "";
    vnode.state.parsedOK = true;

    vnode.state.addOrSetForm = new AddOrSetForm(vnode.state.goal);
  }

  this.view = function(vnode) {
    if (vnode.state.goal.error === "" && vnode.state.goal.user === "") {
      return m(Spinner)
    }
    if (vnode.state.submitError != "") {
      return m("div", [
        m("h4", "An error occurred."),
        m("p", vnode.state.submitError)
      ])
    }
    var fieldContent = null;
    if (vnode.state.loaded === false) {
      fieldContent = "";
      if (vnode.state.goal.rawData) {
        var series = vnode.state.goal.rawData.series;
        for (var i in series) {
          fieldContent += series[i].ts.toLocaleDateString('en-US') + ", " + series[i].value + "\n";
        }
      }
      if (fieldContent != "") {
        vnode.state.parsedEvents = parse.csv(fieldContent);
      }
      vnode.state.loaded = true;
    }
    var events = [];
    vnode.state.parsedEvents.forEach(function(e) {
      events.push(m("li", [
        m("strong", (new Date(e.ts.getTime() + e.ts.getTimezoneOffset()*60*1000)).toDateString()),
        ": ",
        m("span", e.value)
      ]))
    })

    var finalComponents = [];
    if (vnode.state.parsedOK) {
      finalComponents.push(m("div", {style: {marginTop: "10px"}}, "Looks good."))
      finalComponents.push(m("details", {style: "margin: 1rem 0;"}, m("ul", events)))
      finalComponents.push(m("button", {class: "pure-button", onclick: function() {
        Goal.addData(vnode.state.goal, vnode.state.parsedEvents)
          .then(function() {location.reload()}).catch(function(e) {
            vnode.state.submitError = e;
          })
      }}, "Submit"))
    } else {
      finalComponents.push(m("div", {style: {marginTop: "10px"}}, "There’s something wrong with your data."))
    }

    return m("div", [
      m("div", {style: "padding-top: 0.5rem;"}, [
        m("h3.tv-goal-details-section-title", "Add Data"),
        m("h4", "Quick add"),
        m("p", m("strong", "Add to"), " or ", m("strong", "set"), " today’s value."),
        m(this.addOrSetForm),
        m("hr.tv-section-separator"),
        m("h4", "Raw data input"),
        m("form.pure-form", m("textarea", {
          oninput: function() {
            if (this.value.length > 0) {
              try {
                vnode.state.parsedEvents = parse.csv(this.value);
                vnode.state.parsedOK = true;
              } catch(e) {
                vnode.state.parsedOK = false;
              }
            }
          },
          style: {
            maxWidth: "80%",
            width: "300px",
            resize: "both",
            height: "170px"
          }
        }, fieldContent)),
        finalComponents
      ])
    ])
  }
}

// Settings

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

var GoalSettingsPage = function(goal) {
  this.oninit = function(vnode) {
    vnode.state.goal = goal;
    vnode.state.goalUpdateForm = m(UpdateGoalForm, {
      goal: vnode.state.goal,
      update: function(goal) {
        Goal.update(goal).then(function() {location.reload()})
      }
    })
  }

  this.view = function(vnode) {
    var archiveAction = (vnode.state.goal.archived ? "unarchive" : "archive");
    var archiveActionButtonText = (vnode.state.goal.archived ? "Unarchive" : "Archive");
    if (vnode.state.goal.error === "" && vnode.state.goal.user === "") {
      return m(Spinner)
    }
    return m("div", [
      m("div.col-sm-9", {style: "padding-top: 0.5rem;"}, [
        m("h3.tv-goal-details-section-title", "Settings"),
        m("div", {style: "margin-bottom: 1rem;"}, vnode.state.goalUpdateForm),
        m("hr.tv-section-separator"),
        m("form", {
          class: "pure-form pure-form-stacked"
        },
        [
          m("button", {
            class: "pure-button tv-archive-button",
            onclick: function(e) {
              if (confirm("Are you sure you want to "+archiveAction+" this goal?")) {
                if (!vnode.state.goal.archived) {
                  Goal.archive(vnode.state.goal).then(function() {m.route.set("/goals/")})
                } else {
                  Goal.unarchive(vnode.state.goal).then(function() {m.route.set("/goals/")})
                }
              }
              return false;
            }
          }, archiveActionButtonText+" this goal"),
          " ",
          m("button", {
            class: "pure-button tv-delete-button",
            onclick: function(e) {
              if (confirm("Are you sure you want to delete this goal?")) {
                Goal.delete(vnode.state.goal).then(function() {m.route.set("/goals/")})
              }
              return false;
            }
          }, "Delete this goal")
        ])
      ])
    ])
  }
}

module.exports = GoalDetailsPage;
