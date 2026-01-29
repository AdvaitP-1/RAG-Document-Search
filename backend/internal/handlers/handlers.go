package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"rag-backend/internal/auth"
	"rag-backend/internal/store"
)

type API struct {
	store *store.Store
}

func New(store *store.Store) *API {
	return &API{store: store}
}

func (a *API) RegisterPublic(r chi.Router) {
	r.Get("/health", a.handleHealth)
}

func (a *API) RegisterProtected(r chi.Router, authMiddleware *auth.Middleware) {
	r.With(authMiddleware.RequireAuth).Route("/", func(r chi.Router) {
		r.Post("/collections", a.handleCreateCollection)
		r.Get("/collections", a.handleListCollections)
		r.Post("/collections/{id}/docs", a.handleCreateDocument)
		r.Get("/docs/{id}", a.handleGetDocument)
		r.Get("/ingestions/{jobId}", a.handleGetIngestionJob)
	})
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createCollectionRequest struct {
	Name string `json:"name"`
}

func (a *API) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	userID := requireUserID(w, r)
	if userID == "" {
		return
	}
	var req createCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	collection, err := a.store.CreateCollection(r.Context(), userID, req.Name)
	if err != nil {
		http.Error(w, "failed to create collection", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, collection)
}

func (a *API) handleListCollections(w http.ResponseWriter, r *http.Request) {
	userID := requireUserID(w, r)
	if userID == "" {
		return
	}
	collections, err := a.store.ListCollections(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to load collections", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, collections)
}

type createDocumentRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type createDocumentResponse struct {
	Document store.Document     `json:"document"`
	Job      store.IngestionJob `json:"job"`
}

func (a *API) handleCreateDocument(w http.ResponseWriter, r *http.Request) {
	userID := requireUserID(w, r)
	if userID == "" {
		return
	}
	collectionID := chi.URLParam(r, "id")
	if collectionID == "" {
		http.Error(w, "missing collection id", http.StatusBadRequest)
		return
	}
	var req createDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_ = req.Content

	doc, job, err := a.store.CreateDocumentAndJob(r.Context(), userID, collectionID, req.Title)
	if err != nil {
		if err == store.ErrForbidden {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "failed to create document", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, createDocumentResponse{Document: doc, Job: job})
}

func (a *API) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	userID := requireUserID(w, r)
	if userID == "" {
		return
	}
	documentID := chi.URLParam(r, "id")
	if documentID == "" {
		http.Error(w, "missing document id", http.StatusBadRequest)
		return
	}

	doc, err := a.store.GetDocument(r.Context(), userID, documentID)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load document", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (a *API) handleGetIngestionJob(w http.ResponseWriter, r *http.Request) {
	userID := requireUserID(w, r)
	if userID == "" {
		return
	}
	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		http.Error(w, "missing job id", http.StatusBadRequest)
		return
	}

	job, err := a.store.GetIngestionJob(r.Context(), userID, jobID)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load job", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func requireUserID(w http.ResponseWriter, r *http.Request) string {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return ""
	}
	return userID
}
