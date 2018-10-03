var o = require("mithril/ospec/ospec")

o.spec("math", function() {
  o("addition works", function() {
    o(1 + 2).equals(3)
  })
})
