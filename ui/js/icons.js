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

var m = require("mithril");

var cog = {
  view: function (vnode) {
    return m("svg", { "width": "16px", "height": "16px", "viewBox": "0 0 16 16", "version": "1.1", "xmlns": "http://www.w3.org/2000/svg", "xmlns:xlink": "http://www.w3.org/1999/xlink" },
      m("g", { "stroke": "none", "stroke-width": "1", "fill": "none", "fill-rule": "evenodd" },
        m("g", { "fill": "#1D1D1D", "fill-rule": "nonzero" },
          m("path", { "d": "M13.95,8.78 C13.98,8.53 14,8.27 14,8 C14,7.73 13.98,7.47 13.94,7.22 L15.63,5.9 C15.78,5.78 15.82,5.56 15.73,5.39 L14.13,2.62 C14.03,2.44 13.82,2.38 13.64,2.44 L11.65,3.24 C11.23,2.92 10.79,2.66 10.3,2.46 L10,0.34 C9.97,0.14 9.8,1.77635684e-15 9.6,1.77635684e-15 L6.4,1.77635684e-15 C6.2,1.77635684e-15 6.04,0.14 6.01,0.34 L5.71,2.46 C5.22,2.66 4.77,2.93 4.36,3.24 L2.37,2.44 C2.19,2.37 1.98,2.44 1.88,2.62 L0.28,5.39 C0.18,5.57 0.22,5.78 0.38,5.9 L2.07,7.22 C2.03,7.47 2,7.74 2,8 C2,8.26 2.02,8.53 2.06,8.78 L0.37,10.1 C0.22,10.22 0.18,10.44 0.27,10.61 L1.87,13.38 C1.97,13.56 2.18,13.62 2.36,13.56 L4.35,12.76 C4.77,13.08 5.21,13.34 5.7,13.54 L6,15.66 C6.04,15.86 6.2,16 6.4,16 L9.6,16 C9.8,16 9.97,15.86 9.99,15.66 L10.29,13.54 C10.78,13.34 11.23,13.07 11.64,12.76 L13.63,13.56 C13.81,13.63 14.02,13.56 14.12,13.38 L15.72,10.61 C15.82,10.43 15.78,10.22 15.62,10.1 L13.95,8.78 L13.95,8.78 Z M8,11 C6.35,11 5,9.65 5,8 C5,6.35 6.35,5 8,5 C9.65,5 11,6.35 11,8 C11,9.65 9.65,11 8,11 Z" })
        )
      )
    )
  }
};

module.exports = {
  cog: cog,
}
