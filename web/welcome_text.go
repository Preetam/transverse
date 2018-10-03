package main

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

var welcomeHTML = `

<hr>

<p>Welcome to Transverse!</p>

<h4>How it works &amp; best practices</h4>

<p>Transverse uses daily time series points to determine a trend and forecast future values. It's just drawing a line.</p>

<p>You enter data as a CSV (or TSV; the parser does some delimiter detection). Often times you can just copy+paste stuff out of a spreadsheet. Example:</p>

<pre>
11/1/2017, 10
11/2/2017, 11
11/3/2017, 12
</pre>

<p>You can enter data in manually or use the "quick add" form.</p>

<p>It's OK to skip days. Transverse will fill in gaps, but it uses linear interpolation so depending on your units, you have to manually enter in "0" values.</p>

<p>Transverse only works with daily points. It is also tuned for short-term forecasts (within a month). Anything more is unsupported for now. Transverse will also not generate forecasts longer than the period of time you've added data for.</p>
`

var welcomePlaintext = `

-------

Welcome to Transverse!

How it works & best practices
-----------------------------

Transverse uses daily time series points to determine a trend and forecast future values.
It's just drawing a line.

You enter data as a CSV (or TSV; the parser does some delimiter detection).
Often times you can just copy+paste stuff out of a spreadsheet. Example:

11/1/2017, 10
11/2/2017, 11
11/3/2017, 12

You can enter data in manually or use the "quick add" form.

It's OK to skip days. Transverse will fill in gaps, but it uses linear interpolation so
depending on your units, you have to manually enter in "0" values.

Transverse only works with daily points. It is also tuned for short-term forecasts
(within a month). Anything more is unsupported for now. Transverse will also not generate
forecasts longer than the period of time you've added data for.
`
