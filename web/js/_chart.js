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

var m = require("mithril");
var d3 = require("d3");

var Chart = function(chartState) {
  this.oninit = function(vnode) {
    vnode.state.chartState = chartState;
  };
  this.view = function(vnode) {
    // Resize
    var resize = function(vnode) {
      var spacing = Math.abs(this.chartState.minVal - this.chartState.maxVal) * 0.15;
      var chart = vnode.dom;
      var width = parseInt(d3.select(chart).style("width"));
      var data = this.chartState.data,
               w = width,
               h = this.chartState.height,
               margin = 35,
               y = d3.scaleLinear().domain([ this.chartState.minVal-spacing, this.chartState.maxVal+spacing ]).range([ h - margin, 10 ]),
               x = d3.scaleTime().domain([ this.chartState.start, this.chartState.end ]).range([ 0 + 2*margin, w - margin ]);
      var yAxis = d3.axisLeft(y).ticks(4).tickFormat(d3.format("5.3s"));
      var xAxis = d3.axisBottom(x).ticks(Math.round(width/175)+1).tickFormat(d3.timeFormat("%m/%d"));
      // Remove existing paths
      d3.select(chart).selectAll("path").remove();

      // Remove existing text
      d3.select(chart).select(".lineGroup").selectAll("text").remove();

      var target = 0;
      // Draw paths
      for (i in data.lines) {
        if (i == "target") {
          target = data.lines[i][0].value;
        }
        var lineData = data.lines[i];
        var line = d3.line().x(function(d, i) {
          return x(new Date(d.ts));
        }).y(function(d) {
          return y(d.value);
        });
        d3.select(chart).select(".lineGroup").append("path")
          .attr("d", line(lineData))
          .attr("fill", "none")
          .attr("stroke", i == "target" ? "#0075ed" : (i == "low" || i == "high" ? "#ddd" : "#aaa"))
          .attr("stroke-width", i == "series" ? "2px" : "1px")
          .attr("stroke-dasharray", i == "target" || i == "prediction" || i == "low" || i == "high" ? "5, 5" : "none")
          .attr("class", i);
      }

      d3.select(chart).select(".lineGroup").append("text")
        .attr("fill", "#0075ed")
        .attr("style", "font-size: 12px; -webkit-font-smoothing: antialiased")
        .attr("x", width-margin*2)
        .attr("y", y(target)-3)
        .text(d3.format("5.3s")(target))

      // Draw axes
      d3.select(chart).select(".y-axis").attr("transform", "translate(" + (2*margin-10) + ", 0)").call(yAxis);
      d3.select(chart).select(".x-axis").attr("transform", "translate(0, " + (h - margin + 10) + ")").call(xAxis);
    }.bind(this);
    // Draw
    var draw = function(vnode) {
      d3.select(window).on("resize." + this.chartState.name, resize.bind(null, vnode));
      resize(vnode);
    }.bind(this);
    // Elements
    return m("div.chart", [
      m("svg", {
        width: "100%",
        height: this.chartState.height,
        oncreate: draw.bind(this)
      },
      m("g", [ m("g.lineGroup"), m("g.x-axis"), m("g.y-axis") ]))
    ]);
  };
};

module.exports = Chart;
