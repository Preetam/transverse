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

var Goals = function(data) {
  var data = data || {};
  this.data = data;
}

Goals.get = function(goals, archived) {
  return req({
    method: "GET",
    url: "/api/v1/goals?showArchived="+(archived ? "true" : "false")
  }).then(function(d) {
    var d = d.data;
    goals.data = d;
  })
}

var Goal = function(data) {
  var data = data || {};
  this.id = data.id || "";
  this.name = data.name || "";
  this.user = data.user || "";
  this.description = data.description || "";
  this.target = data.target || 0;
  this.archived = data.archived || false;
  this.created = data.created || 0;
  this.updated = data.updated || 0;
  this.deleted = data.deleted || 0;

  this.setName = function(v) {
    this.name = v;
  };

  this.setTarget = function(v) {
    this.target = v;
    var parsed = parseFloat(v);
    if (parsed == parsed) {
      this.target = parsed;
    }
  };
}

Goal.get = function(goal) {
  return req({
    method: "GET",
    url: "/api/v1/goals/"+goal.id
  }).then(function(d) {
    var d = d.data;
    goal.name = d.name;
    goal.user = d.user;
    goal.description = d.description;
    goal.target = d.target;
    goal.archived = d.archived;
    goal.created = d.created;
    goal.updated = d.updated;
    goal.deleted = d.deleted;
  })
}

Goal.create = function(goal) {
  return req({
    method: "POST",
    url: "/api/v1/goals",
    data: goal
  })
}

Goal.update = function(goal) {
  return req({
    method: "PUT",
    url: "/api/v1/goals/"+goal.id,
    data: goal
  })
}

Goal.delete = function(goal) {
  return req({
    method: "DELETE",
    url: "/api/v1/goals/"+goal.id
  })
}

Goal.archive = function(goal) {
  goal.archived = true;
  return req({
    method: "PUT",
    url: "/api/v1/goals/"+goal.id,
    data: goal
  })
}

Goal.unarchive = function(goal) {
  goal.archived = false;
  return req({
    method: "PUT",
    url: "/api/v1/goals/"+goal.id,
    data: goal
  })
}

Goal.getData = function(goal) {
  return req({
    method: "GET",
    url: "/api/v1/goals/"+goal.id+"/data"
  }).then(function(d) {
    var d = d.data;
    for (var i in d.series) {
      d.series[i].ts = new Date(d.series[i].ts)
      d.series[i].ts = new Date(d.series[i].ts.getTime() + d.series[i].ts.getTimezoneOffset()*60*1000)
      if (!goal.data_start || d.series[i].ts < goal.data_start) {
        goal.data_start = d.series[i].ts;
      }
      if (!goal.data_end || d.series[i].ts > goal.data_end) {
        goal.data_end = d.series[i].ts;
      }
      goal.end = d.series[i];

      if (!goal.min) {
        goal.min = d.series[i];
      }
      if (d.series[i].value < goal.min.value) {
        goal.min = d.series[i];
      }

      if (!goal.max) {
        goal.max = d.series[i];
      }
      if (d.series[i].value > goal.max.value) {
        goal.max = d.series[i];
      }
    }
    for (var i in d.prediction) {
      d.prediction[i].ts = new Date(d.prediction[i].ts)
      d.prediction[i].ts = new Date(d.prediction[i].ts.getTime() + d.prediction[i].ts.getTimezoneOffset()*60*1000)

      if (!goal.data_start || d.prediction[i].ts < goal.data_start) {
        goal.data_start = d.prediction[i].ts;
      }
      if (!goal.data_end || d.prediction[i].ts > goal.data_end) {
        goal.data_end = d.prediction[i].ts;
      }
      if (d.low[i]) {
        d.low[i].ts = new Date(d.low[i].ts)
        d.low[i].ts = new Date(d.low[i].ts.getTime() + d.low[i].ts.getTimezoneOffset()*60*1000)
      }
      if (d.high[i]) {
        d.high[i].ts = new Date(d.high[i].ts)
        d.high[i].ts = new Date(d.high[i].ts.getTime() + d.high[i].ts.getTimezoneOffset()*60*1000)
      }
    }
    if (d.prediction.length > 1) {
      goal.slope = d.prediction[1].value - d.prediction[0].value;
    }
    if (d.eta) {
      if (d.eta > 0) {
        goal.eta = d.eta;
      }
    }
    goal.data = d;
  })
}

Goal.getETA = function(goal) {
  return req({
    method: "GET",
    url: "/api/v1/goals/"+goal.id+"/eta"
  }).then(function(d) {
    var d = d.data;
    if (d.eta) {
      goal.eta = d.eta;
    }
  })
}

Goal.getRawData = function(goal) {
  return req({
    method: "GET",
    url: "/api/v1/goals/"+goal.id+"/raw-data"
  }).then(function(d) {
    var d = d.data;
    for (var i in d.series) {
      d.series[i].ts = new Date(d.series[i].ts)
      d.series[i].ts = new Date(d.series[i].ts.getTime() + d.series[i].ts.getTimezoneOffset()*60*1000)
    }
    goal.rawData = d;
  })
}

Goal.addData = function(goal, data) {
  return req({
    method: "POST",
    url: "/api/v1/goals/"+goal.id+"/data",
    data: data
  })
}

Goal.addDataAddPoint = function(goal, value) {
  var now = new Date();
  var normalized = new Date(now.getTime() - (now.getTimezoneOffset()*60*1000))
  return req({
    method: "POST",
    url: "/api/v1/goals/"+goal.id+"/data/single?add=true",
    data: {
      value: value,
      date: normalized
    }
  })
}

Goal.addDataSetPoint = function(goal, value) {
  var now = new Date();
  var normalized = new Date(now.getTime() - (now.getTimezoneOffset()*60*1000))
  return req({
    method: "POST",
    url: "/api/v1/goals/"+goal.id+"/data/single?add=false",
    data: {
      value: value,
      date: normalized
    }
  })
}

module.exports = {
  Goals: Goals,
  Goal: Goal,
}
