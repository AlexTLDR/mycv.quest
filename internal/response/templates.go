package response

import (
	"context"
	"net/http"

	templates "github.com/AlexTLDR/mycv.quest/assets/templates/templ"
	"github.com/a-h/templ"
)

// Component renders a templ component with the given status code
func Component(w http.ResponseWriter, status int, component templ.Component) error {
	return ComponentWithHeaders(w, status, component, nil)
}

// ComponentWithHeaders renders a templ component with the given status code and headers
func ComponentWithHeaders(w http.ResponseWriter, status int, component templ.Component, headers http.Header) error {
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.WriteHeader(status)
	return component.Render(context.Background(), w)
}

// Page renders a page using the base layout with the given data
func Page(w http.ResponseWriter, status int, data templates.PageData, content templ.Component) error {
	return PageWithHeaders(w, status, data, nil, content)
}

// PageWithHeaders renders a page using the base layout with the given data and headers
func PageWithHeaders(w http.ResponseWriter, status int, data templates.PageData, headers http.Header, content templ.Component) error {
	page := templates.Base("Page", data, nil, content)
	return ComponentWithHeaders(w, status, page, headers)
}

// PageWithTitle renders a page using the base layout with a custom title
func PageWithTitle(w http.ResponseWriter, status int, title string, data templates.PageData, content templ.Component) error {
	return PageWithTitleAndHeaders(w, status, title, data, nil, content)
}

// PageWithTitleAndHeaders renders a page using the base layout with a custom title and headers
func PageWithTitleAndHeaders(w http.ResponseWriter, status int, title string, data templates.PageData, headers http.Header, content templ.Component) error {
	page := templates.Base(title, data, nil, content)
	return ComponentWithHeaders(w, status, page, headers)
}

// PageWithMeta renders a page using the base layout with custom meta tags
func PageWithMeta(w http.ResponseWriter, status int, title string, data templates.PageData, meta templ.Component, content templ.Component) error {
	return PageWithMetaAndHeaders(w, status, title, data, meta, content, nil)
}

// PageWithMetaAndHeaders renders a page using the base layout with custom meta tags and headers
func PageWithMetaAndHeaders(w http.ResponseWriter, status int, title string, data templates.PageData, meta templ.Component, content templ.Component, headers http.Header) error {
	page := templates.Base(title, data, meta, content)
	return ComponentWithHeaders(w, status, page, headers)
}
