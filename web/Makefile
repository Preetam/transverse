CSS_SRC := css/style.css

static/css/style.min.css: $(CSS_SRC)
	cleancss $(CSS_SRC) -o $@

all: static/css/style.min.css

.PHONY: all
