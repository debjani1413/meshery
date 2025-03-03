package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/layer5io/meshery/handlers"
	"github.com/layer5io/meshery/models"
)

// Router represents Meshery router
type Router struct {
	S    *mux.Router
	port int
}

// NewRouter returns a new ServeMux with app routes.
func NewRouter(ctx context.Context, h models.HandlerInterface, port int, g http.Handler, gp http.Handler) *Router {
	gMux := mux.NewRouter()

	gMux.Handle("/api/system/graphql/query", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GraphqlMiddleware(g))))).Methods("GET", "POST")
	gMux.Handle("/api/system/graphql/playground", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GraphqlMiddleware(gp))))).Methods("GET", "POST")

	gMux.HandleFunc("/api/system/version", h.ServerVersionHandler).
		Methods("GET")

	gMux.HandleFunc("/api/provider", h.ProviderHandler)
	gMux.HandleFunc("/api/providers", h.ProvidersHandler).
		Methods("GET")
	gMux.PathPrefix("/api/provider/extension").
		Handler(h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.ProviderComponentsHandler)))).
		Methods("GET", "POST", "OPTIONS", "PUT", "DELETE")
	gMux.Handle("/api/provider/capabilities", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.ProviderCapabilityHandler)))).
		Methods("GET")
	gMux.PathPrefix("/provider").
		Handler(http.HandlerFunc(h.ProviderUIHandler)).
		Methods("GET")
	// gMux.PathPrefix("/provider/").
	// 	Handler(http.StripPrefix("/provider/", http.FileServer(http.Dir("../provider-ui/out/")))).
	// 	Methods("GET")

	gMux.Handle("/api/system/sync", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.SessionSyncHandler)))).
		Methods("GET")

	gMux.Handle("/api/user", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.UserHandler)))).
		Methods("GET")
	gMux.Handle("/api/user/prefs", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.UserPrefsHandler)))).
		Methods("GET", "POST")

	gMux.Handle("/api/user/prefs/perf", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.UserTestPreferenceHandler)))).
		Methods("GET", "POST", "DELETE")

	gMux.Handle("/api/k8sconfig", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.K8SConfigHandler)))).
		Methods("POST", "DELETE")
	gMux.Handle("/api/k8sconfig/contexts", h.ProviderMiddleware(h.AuthMiddleware(http.HandlerFunc(h.GetContextsFromK8SConfig)))).
		Methods("POST")
	gMux.Handle("/api/k8sconfig/ping", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.KubernetesPingHandler))))

	gMux.Handle("/api/perf/profile", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.LoadTestHandler)))).
		Methods("GET", "POST")
	gMux.Handle("/api/perf/profile/result", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.FetchAllResultsHandler)))).
		Methods("GET")
	gMux.Handle("/api/perf/profile/result/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GetResultHandler)))).
		Methods("GET")
	gMux.Handle("/api/mesh", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GetSMPServiceMeshes)))).
		Methods("GET")

	gMux.Handle("/api/smi/results", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.FetchSmiResultsHandler)))).
		Methods("GET")

	gMux.Handle("/api/mesh/manage", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.MeshAdapterConfigHandler))))
	gMux.Handle("/api/mesh/ops", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.MeshOpsHandler)))).
		Methods("POST")
	gMux.Handle("/api/mesh/adapters", h.ProviderMiddleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		providerI := req.Context().Value(models.ProviderCtxKey)
		provider, ok := providerI.(models.Provider)
		if !ok {
			http.Redirect(w, req, "/provider", http.StatusFound)
			return
		}
		h.GetAllAdaptersHandler(w, req, provider)
	})))
	gMux.Handle("/api/mesh/adapter/ping", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.AdapterPingHandler)))).
		Methods("GET")
	gMux.Handle("/api/events", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.EventStreamHandler)))).
		Methods("GET")

	gMux.Handle("/api/telemetry/metrics/grafana/config", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GrafanaConfigHandler)))).
		Methods("GET", "POST", "DELETE")
	gMux.Handle("/api/telemetry/metrics/grafana/boards", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GrafanaBoardsHandler)))).
		Methods("GET", "POST")
	gMux.Handle("/api/telemetry/metrics/grafana/query", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GrafanaQueryHandler)))).
		Methods("GET")
	gMux.Handle("/api/grafana/query_range", h.ProviderMiddleware(h.AuthMiddleware(http.HandlerFunc(h.GrafanaQueryRangeHandler)))).
		Methods("GET")
	gMux.Handle("/api/telemetry/metrics/grafana/ping", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GrafanaPingHandler)))).
		Methods("GET")
	gMux.Handle("/api/telemetry/metrics/grafana/scan", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.ScanGrafanaHandler))))

	gMux.Handle("/api/telemetry/metrics/config", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.PrometheusConfigHandler)))).
		Methods("GET", "POST", "DELETE")
	gMux.Handle("/api/telemetry/metrics/board_import", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GrafanaBoardImportForPrometheusHandler)))).
		Methods("POST")
	gMux.Handle("/api/telemetry/metrics/query", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.PrometheusQueryHandler)))).
		Methods("GET")
	gMux.Handle("/api/prometheus/query_range", h.ProviderMiddleware(h.AuthMiddleware(http.HandlerFunc(h.PrometheusQueryRangeHandler)))).
		Methods("GET")
	gMux.Handle("/api/telemetry/metrics/ping", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.PrometheusPingHandler)))).
		Methods("GET")
	gMux.Handle("/api/telemetry/metrics/static-board", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.PrometheusStaticBoardHandler)))).
		Methods("GET")
	gMux.Handle("/api/telemetry/metrics/boards", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.SaveSelectedPrometheusBoardsHandler)))).
		Methods("POST")
	gMux.Handle("/api/prometheus/scan", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.ScanPrometheusHandler))))

	gMux.Handle("/api/promGrafana/scan", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.ScanPromGrafanaHandler))))

	gMux.Handle("/api/pattern/deploy", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.PatternFileHandler)))).
		Methods("POST", "DELETE")
	gMux.Handle("/api/pattern", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.PatternFileRequestHandler)))).
		Methods("POST", "GET")
	gMux.Handle("/api/pattern/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GetMesheryPatternHandler)))).
		Methods("GET")
	gMux.Handle("/api/pattern/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.DeleteMesheryPatternHandler)))).
		Methods("DELETE")
	gMux.HandleFunc("/api/oam/{type}", h.OAMRegisterHandler).Methods("GET", "POST")
	gMux.HandleFunc("/api/experimental/oam/{type}", h.OAMRegisterHandler).Methods("GET", "POST")

	gMux.Handle("/api/experimental/filter", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.FilterFileRequestHandler)))).
		Methods("POST", "GET")
	gMux.Handle("/api/experimental/filter/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GetMesheryFilterHandler)))).
		Methods("GET")
	gMux.Handle("/api/experimental/filter/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.DeleteMesheryFilterHandler)))).
		Methods("DELETE")
	gMux.Handle("/api/experimental/filter/file/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GetMesheryFilterFileHandler)))).
		Methods("GET")

	gMux.Handle("/api/experimental/application/deploy", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.ApplicationFileHandler)))).
		Methods("POST", "DELETE")
	gMux.Handle("/api/experimental/application", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.ApplicationFileRequestHandler)))).
		Methods("POST", "GET")
	gMux.Handle("/api/experimental/application/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GetMesheryApplicationHandler)))).
		Methods("GET")
	gMux.Handle("/api/experimental/application/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.DeleteMesheryApplicationHandler)))).
		Methods("DELETE")

	gMux.Handle("/api/user/performance/profiles", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GetPerformanceProfilesHandler)))).
		Methods("GET")
	gMux.Handle("/api/user/performance/profiles/results", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.FetchAllResultsHandler)))).
		Methods("GET")
	gMux.Handle("/api/user/performance/profiles/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GetPerformanceProfileHandler)))).
		Methods("GET")
	gMux.Handle("/api/user/performance/profiles/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.DeletePerformanceProfileHandler)))).
		Methods("DELETE")
	gMux.Handle("/api/user/performance/profiles", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.SavePerformanceProfileHandler)))).
		Methods("POST")
	gMux.Handle("/api/user/performance/profiles/{id}/run", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.LoadTestHandler)))).
		Methods("GET")
	gMux.Handle("/api/user/performance/profiles/{id}/results", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.FetchResultsHandler)))).
		Methods("GET")

	gMux.Handle("/api/user/schedules", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GetSchedulesHandler)))).
		Methods("GET")
	gMux.Handle("/api/user/schedules/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GetScheduleHandler)))).
		Methods("GET")
	gMux.Handle("/api/user/schedules/{id}", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.DeleteScheduleHandler)))).
		Methods("DELETE")
	gMux.Handle("/api/user/schedules", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.SaveScheduleHandler)))).
		Methods("POST")

	//gMux.PathPrefix("/api/system/graphql").Handler(h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(h.GraphqlSystemHandler)))).Methods("GET", "POST")

	gMux.Handle("/user/logout", h.ProviderMiddleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		providerI := req.Context().Value(models.ProviderCtxKey)
		provider, ok := providerI.(models.Provider)
		if !ok {
			logrus.Debug("Inside not OK")
			http.Redirect(w, req, "/provider", http.StatusFound)
			return
		}
		h.LogoutHandler(w, req, provider)
	})))
	gMux.Handle("/user/login", h.ProviderMiddleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		providerI := req.Context().Value(models.ProviderCtxKey)
		provider, ok := providerI.(models.Provider)
		if !ok {
			http.Redirect(w, req, "/provider", http.StatusFound)
			return
		}
		h.LoginHandler(w, req, provider, false)
	})))
	gMux.Handle("/api/user/token", h.ProviderMiddleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		providerI := req.Context().Value(models.ProviderCtxKey)
		provider, ok := providerI.(models.Provider)
		if !ok {
			http.Redirect(w, req, "/provider", http.StatusFound)
			return
		}
		h.TokenHandler(w, req, provider, false)
	}))).Methods("POST", "GET")
	gMux.Handle("/api/token", h.ProviderMiddleware(h.AuthMiddleware(h.SessionInjectorMiddleware(
		func(w http.ResponseWriter, req *http.Request, _ *models.Preference, _ *models.User, provider models.Provider) {
			provider.ExtractToken(w, req)
		})))).Methods("GET")

	// TODO: have to change this too
	gMux.Handle("/favicon.ico", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hr
		http.ServeFile(w, r, "../ui/out/static/img/meshery-logo/meshery-logo.svg")
	}))

	// Swagger Interactive Playground
	swaggerOpts := middleware.SwaggerUIOpts{SpecURL: "./swagger.yaml"}
	swaggerSh := middleware.SwaggerUI(swaggerOpts, nil)
	gMux.Handle("/swagger.yaml", http.FileServer(http.Dir("../helpers/")))
	gMux.Handle("/docs", swaggerSh)

	gMux.PathPrefix("/").
		Handler(h.ProviderMiddleware(h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.ServeUI(w, r, "", "../ui/out/")
		})))).
		Methods("GET")

	return &Router{
		S:    gMux,
		port: port,
	}
}

// Run starts the http server
func (r *Router) Run() error {
	// s := &http.Server{
	// 	Addr:           fmt.Sprintf(":%d", r.port),
	// 	Handler:        r.s,
	// 	ReadTimeout:    5 * time.Second,
	// 	WriteTimeout:   2 * time.Minute,
	// 	MaxHeaderBytes: 1 << 20,
	// 	IdleTimeout:    0, //time.Second,
	// }
	// return s.ListenAndServe()
	return http.ListenAndServe(fmt.Sprintf(":%d", r.port), r.S)
}
