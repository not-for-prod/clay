package mwhttp

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const (
	notFoundRouteTemplate = "not_found"
	unknownRouteTemplate  = "unknown"
)

type routeTemplateContextKey struct{}

type routeTemplateState struct {
	template string
	onSet    func(string)
}

func withRouteTemplateState(r *http.Request, onSet func(string)) (*http.Request, *routeTemplateState) {
	state := &routeTemplateState{onSet: onSet}
	ctx := context.WithValue(r.Context(), routeTemplateContextKey{}, state)

	return r.WithContext(ctx), state
}

func (s *routeTemplateState) set(template string) {
	if template == "" || s.template != "" {
		return
	}

	s.template = template
	if s.onSet != nil {
		s.onSet(template)
		s.onSet = nil
	}
}

func routeTemplateStateFromContext(ctx context.Context) (*routeTemplateState, bool) {
	state, ok := ctx.Value(routeTemplateContextKey{}).(*routeTemplateState)

	return state, ok
}

func routeTemplate(r *http.Request, state *routeTemplateState, statusCode int) string {
	if state.template != "" {
		return state.template
	}

	if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
		template := routeContext.RoutePattern()
		if template != "" {
			return template
		}
	}

	if statusCode == http.StatusNotFound {
		return notFoundRouteTemplate
	}

	return unknownRouteTemplate
}

func matchedChiRouteTemplate(r *http.Request) string {
	routeContext := chi.RouteContext(r.Context())
	if routeContext == nil || routeContext.Routes == nil {
		return ""
	}

	routePath := routeContext.RoutePath
	if routePath == "" {
		if r.URL.RawPath != "" {
			routePath = r.URL.RawPath
		} else {
			routePath = r.URL.Path
		}
		if routePath == "" {
			routePath = "/"
		}
	}

	routeMethod := routeContext.RouteMethod
	if routeMethod == "" {
		routeMethod = r.Method
	}

	matchContext := chi.NewRouteContext()
	template := routeContext.Routes.Find(matchContext, routeMethod, routePath)
	if template == "/*" {
		return ""
	}

	return template
}

// GRPCGatewayRouteTemplateOption captures the matched grpc-gateway route
// template so metrics use bounded route labels instead of concrete URL paths.
func GRPCGatewayRouteTemplateOption() runtime.ServeMuxOption {
	return runtime.WithMiddlewares(func(next runtime.HandlerFunc) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			state, hasState := routeTemplateStateFromContext(r.Context())
			if hasState && state.template == "" {
				if pattern, ok := runtime.HTTPPattern(r.Context()); ok {
					state.set(pattern.String())
				}
			}

			next(w, r, pathParams)
		}
	})
}
