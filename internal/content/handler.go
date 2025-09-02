package content

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type ContentHandler struct {
	service *ContentService
}

type CreateContentRequest struct {
	OriginalFilename string          `json:"original_filename"`
	FileSize         int64           `json:"file_size"`
	MimeType         string          `json:"mime_type"`
	ContentHash      string          `json:"content_hash"`
	ContentCategory  ContentCategory `json:"content_category"`
	AltText          string          `json:"alt_text,omitempty"`
	Description      string          `json:"description,omitempty"`
	Tags             []string        `json:"tags,omitempty"`
	AccessLevel      AccessLevel     `json:"access_level,omitempty"`
}

type UpdateContentMetadataRequest struct {
	AltText     string   `json:"alt_text,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type SetAccessLevelRequest struct {
	AccessLevel AccessLevel `json:"access_level"`
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
	// Public routes - simple GET endpoints for environment validation
	public := router.PathPrefix("/api/v1/content").Subrouter()
	public.HandleFunc("", h.ListAvailableContent).Methods("GET")
	public.HandleFunc("/{id}", h.GetContent).Methods("GET")
	public.HandleFunc("/category/{category}", h.ListContentByCategory).Methods("GET")
	public.HandleFunc("/type/{type}", h.ListContentByType).Methods("GET")
	public.HandleFunc("/tags", h.ListContentByTags).Methods("GET")
	
	// Admin routes - simple GET endpoints for environment validation
	admin := router.PathPrefix("/admin/api/v1/content").Subrouter()
	admin.HandleFunc("", h.ListAllContent).Methods("GET")
	admin.HandleFunc("/{id}", h.GetContent).Methods("GET")
}

func (h *ContentHandler) GetContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID := vars["id"]
	
	content, err := h.service.GetContent(r.Context(), contentID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Content not found", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, content)
}

func (h *ContentHandler) ListAllContent(w http.ResponseWriter, r *http.Request) {
	offset, limit := h.getPagination(r)
	
	content, err := h.service.ListContent(r.Context(), offset, limit)
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

func (h *ContentHandler) ListAvailableContent(w http.ResponseWriter, r *http.Request) {
	offset, limit := h.getPagination(r)
	
	content, err := h.service.ListAvailableContent(r.Context(), offset, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to list available content", err.Error())
		return
	}
	
	response := &ContentListResponse{
		Content: content,
		Total:   len(content),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

func (h *ContentHandler) ListContentByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentCategory := ContentCategory(vars["category"])
	offset, limit := h.getPagination(r)
	
	content, err := h.service.ListContentByCategory(r.Context(), contentCategory, offset, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to list content by category", err.Error())
		return
	}
	
	response := &ContentListResponse{
		Content: content,
		Total:   len(content),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

func (h *ContentHandler) ListContentByType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentCategory := ContentCategory(vars["type"])
	offset, limit := h.getPagination(r)
	
	content, err := h.service.ListContentByType(r.Context(), contentCategory, offset, limit)
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
	tagsParam := r.URL.Query().Get("tags")
	if tagsParam == "" {
		h.sendError(w, http.StatusBadRequest, "Tags parameter is required", "")
		return
	}
	
	tags := []string{tagsParam}
	if strings.Contains(tagsParam, ",") {
		tags = strings.Split(tagsParam, ",")
	}
	offset, limit := h.getPagination(r)
	
	content, err := h.service.ListContentByTags(r.Context(), tags, offset, limit)
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