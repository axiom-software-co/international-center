package content

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type ContentHandler struct {
	service *ContentService
}

type CreateContentRequest struct {
	Title            string   `json:"title"`
	Body             string   `json:"body"`
	Slug             string   `json:"slug"`
	ContentType      string   `json:"content_type"`
	Tags             []string `json:"tags,omitempty"`
	MetaDescription  string   `json:"meta_description,omitempty"`
	FeaturedImageURL string   `json:"featured_image_url,omitempty"`
}

type UpdateContentRequest struct {
	Title            string   `json:"title"`
	Body             string   `json:"body"`
	Slug             string   `json:"slug"`
	ContentType      string   `json:"content_type"`
	Tags             []string `json:"tags,omitempty"`
	MetaDescription  string   `json:"meta_description,omitempty"`
	FeaturedImageURL string   `json:"featured_image_url,omitempty"`
}

type AssignCategoryRequest struct {
	CategoryID string `json:"category_id"`
}

type AssignTagsRequest struct {
	Tags []string `json:"tags"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type ContentListResponse struct {
	Content []*Content `json:"content"`
	Total   int        `json:"total"`
}

func NewContentHandler(service *ContentService) *ContentHandler {
	return &ContentHandler{service: service}
}

func (h *ContentHandler) RegisterRoutes(router *mux.Router) {
	// Public routes
	public := router.PathPrefix("/api/v1/content").Subrouter()
	public.HandleFunc("", h.ListPublishedContent).Methods("GET")
	public.HandleFunc("/{id}", h.GetContent).Methods("GET")
	public.HandleFunc("/slug/{slug}", h.GetContentBySlug).Methods("GET")
	public.HandleFunc("/type/{type}", h.ListContentByType).Methods("GET")
	public.HandleFunc("/tags", h.ListContentByTags).Methods("GET")
	
	// Admin routes
	admin := router.PathPrefix("/admin/api/v1/content").Subrouter()
	admin.HandleFunc("", h.CreateContent).Methods("POST")
	admin.HandleFunc("", h.ListAllContent).Methods("GET")
	admin.HandleFunc("/{id}", h.UpdateContent).Methods("PUT")
	admin.HandleFunc("/{id}", h.DeleteContent).Methods("DELETE")
	admin.HandleFunc("/{id}/publish", h.PublishContent).Methods("POST")
	admin.HandleFunc("/{id}/archive", h.ArchiveContent).Methods("POST")
	admin.HandleFunc("/{id}/assign-category", h.AssignContentCategory).Methods("POST")
	admin.HandleFunc("/{id}/assign-tags", h.AssignContentTags).Methods("POST")
}

func (h *ContentHandler) CreateContent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	var req CreateContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	
	content, err := h.service.CreateContent(ctx, req.Title, req.Body, req.Slug, req.ContentType)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to create content", err.Error())
		return
	}
	
	// Assign optional fields if provided
	userID := h.getUserID(r)
	if len(req.Tags) > 0 {
		if err := h.service.AssignContentTags(ctx, content.ContentID, req.Tags, userID); err != nil {
			h.sendError(w, http.StatusBadRequest, "Failed to assign tags", err.Error())
			return
		}
	}
	
	if req.MetaDescription != "" {
		content.MetaDescription = req.MetaDescription
	}
	
	if req.FeaturedImageURL != "" {
		content.FeaturedImageURL = req.FeaturedImageURL
	}
	
	h.sendJSON(w, http.StatusCreated, content)
}

func (h *ContentHandler) GetContent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	vars := mux.Vars(r)
	contentID := vars["id"]
	
	content, err := h.service.GetContent(ctx, contentID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Content not found", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, content)
}

func (h *ContentHandler) GetContentBySlug(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	vars := mux.Vars(r)
	slug := vars["slug"]
	
	content, err := h.service.GetContentBySlug(ctx, slug)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Content not found", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, content)
}

func (h *ContentHandler) UpdateContent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	vars := mux.Vars(r)
	contentID := vars["id"]
	userID := h.getUserID(r)
	
	var req UpdateContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	
	content, err := h.service.UpdateContent(ctx, contentID, req.Title, req.Body, req.Slug, req.ContentType, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to update content", err.Error())
		return
	}
	
	// Update optional fields
	if len(req.Tags) > 0 {
		if err := h.service.AssignContentTags(ctx, contentID, req.Tags, userID); err != nil {
			h.sendError(w, http.StatusBadRequest, "Failed to update tags", err.Error())
			return
		}
	}
	
	if req.MetaDescription != "" {
		content.MetaDescription = req.MetaDescription
	}
	
	if req.FeaturedImageURL != "" {
		content.FeaturedImageURL = req.FeaturedImageURL
	}
	
	h.sendJSON(w, http.StatusOK, content)
}

func (h *ContentHandler) PublishContent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	vars := mux.Vars(r)
	contentID := vars["id"]
	userID := h.getUserID(r)
	
	err := h.service.PublishContent(ctx, contentID, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to publish content", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Content published successfully"})
}

func (h *ContentHandler) ArchiveContent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	vars := mux.Vars(r)
	contentID := vars["id"]
	userID := h.getUserID(r)
	
	err := h.service.ArchiveContent(ctx, contentID, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to archive content", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Content archived successfully"})
}

func (h *ContentHandler) AssignContentCategory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	vars := mux.Vars(r)
	contentID := vars["id"]
	userID := h.getUserID(r)
	
	var req AssignCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	
	err := h.service.AssignContentCategory(ctx, contentID, req.CategoryID, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to assign category", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Category assigned successfully"})
}

func (h *ContentHandler) AssignContentTags(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	vars := mux.Vars(r)
	contentID := vars["id"]
	userID := h.getUserID(r)
	
	var req AssignTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	
	err := h.service.AssignContentTags(ctx, contentID, req.Tags, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to assign tags", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Tags assigned successfully"})
}

func (h *ContentHandler) DeleteContent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	vars := mux.Vars(r)
	contentID := vars["id"]
	userID := h.getUserID(r)
	
	err := h.service.DeleteContent(ctx, contentID, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to delete content", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Content deleted successfully"})
}

func (h *ContentHandler) ListAllContent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	offset, limit := h.getPagination(r)
	
	content, err := h.service.ListContent(ctx, offset, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to list content", err.Error())
		return
	}
	
	response := &ContentListResponse{
		Content: content,
		Total:   len(content),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

func (h *ContentHandler) ListPublishedContent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	offset, limit := h.getPagination(r)
	
	content, err := h.service.ListPublishedContent(ctx, offset, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to list published content", err.Error())
		return
	}
	
	response := &ContentListResponse{
		Content: content,
		Total:   len(content),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

func (h *ContentHandler) ListContentByType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	vars := mux.Vars(r)
	contentType := ContentType(vars["type"])
	offset, limit := h.getPagination(r)
	
	content, err := h.service.ListContentByType(ctx, contentType, offset, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to list content by type", err.Error())
		return
	}
	
	response := &ContentListResponse{
		Content: content,
		Total:   len(content),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

func (h *ContentHandler) ListContentByTags(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	
	tagsParam := r.URL.Query().Get("tags")
	if tagsParam == "" {
		h.sendError(w, http.StatusBadRequest, "Tags parameter is required", "")
		return
	}
	
	tags := strings.Split(tagsParam, ",")
	offset, limit := h.getPagination(r)
	
	content, err := h.service.ListContentByTags(ctx, tags, offset, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to list content by tags", err.Error())
		return
	}
	
	response := &ContentListResponse{
		Content: content,
		Total:   len(content),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

func (h *ContentHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ContentHandler) sendError(w http.ResponseWriter, status int, message, details string) {
	response := &ErrorResponse{
		Error:   message,
		Message: details,
	}
	h.sendJSON(w, status, response)
}

func (h *ContentHandler) getUserID(r *http.Request) string {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		return "system"
	}
	return userID
}

func (h *ContentHandler) getPagination(r *http.Request) (int, int) {
	offset := 0
	limit := 20
	
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	
	return offset, limit
}