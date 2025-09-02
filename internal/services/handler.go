package services

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ServicesHandler struct {
	service *ServicesService
}

type CreateServiceRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Slug         string `json:"slug"`
	DeliveryMode string `json:"delivery_mode"`
}

type UpdateServiceRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Slug         string `json:"slug"`
	DeliveryMode string `json:"delivery_mode"`
}

type AssignCategoryRequest struct {
	CategoryID string `json:"category_id"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type ServicesListResponse struct {
	Services []*Service `json:"services"`
	Total    int        `json:"total"`
}

func NewServicesHandler(service *ServicesService) *ServicesHandler {
	return &ServicesHandler{service: service}
}

func (h *ServicesHandler) RegisterRoutes(router *mux.Router) {
	// Public routes
	public := router.PathPrefix("/api/v1/services").Subrouter()
	public.HandleFunc("", h.ListPublishedServices).Methods("GET")
	public.HandleFunc("/{id}", h.GetService).Methods("GET")
	public.HandleFunc("/slug/{slug}", h.GetServiceBySlug).Methods("GET")
	
	// Admin routes
	admin := router.PathPrefix("/admin/api/v1/services").Subrouter()
	admin.HandleFunc("", h.CreateService).Methods("POST")
	admin.HandleFunc("", h.ListAllServices).Methods("GET")
	admin.HandleFunc("/{id}", h.UpdateService).Methods("PUT")
	admin.HandleFunc("/{id}", h.DeleteService).Methods("DELETE")
	admin.HandleFunc("/{id}/publish", h.PublishService).Methods("POST")
	admin.HandleFunc("/{id}/archive", h.ArchiveService).Methods("POST")
	admin.HandleFunc("/{id}/assign-category", h.AssignServiceCategory).Methods("POST")
}

func (h *ServicesHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	
	service, err := h.service.CreateService(req.Title, req.Description, req.Slug, req.DeliveryMode)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to create service", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusCreated, service)
}

func (h *ServicesHandler) GetService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["id"]
	
	service, err := h.service.GetService(serviceID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Service not found", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, service)
}

func (h *ServicesHandler) GetServiceBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	
	service, err := h.service.GetServiceBySlug(slug)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Service not found", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, service)
}

func (h *ServicesHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["id"]
	userID := h.getUserID(r)
	
	var req UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	
	service, err := h.service.UpdateService(serviceID, req.Title, req.Description, req.Slug, req.DeliveryMode, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to update service", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, service)
}

func (h *ServicesHandler) PublishService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["id"]
	userID := h.getUserID(r)
	
	err := h.service.PublishService(serviceID, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to publish service", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Service published successfully"})
}

func (h *ServicesHandler) ArchiveService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["id"]
	userID := h.getUserID(r)
	
	err := h.service.ArchiveService(serviceID, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to archive service", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Service archived successfully"})
}

func (h *ServicesHandler) AssignServiceCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["id"]
	userID := h.getUserID(r)
	
	var req AssignCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	
	err := h.service.AssignServiceCategory(serviceID, req.CategoryID, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to assign category", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Category assigned successfully"})
}

func (h *ServicesHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["id"]
	userID := h.getUserID(r)
	
	err := h.service.DeleteService(serviceID, userID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to delete service", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Service deleted successfully"})
}

func (h *ServicesHandler) ListAllServices(w http.ResponseWriter, r *http.Request) {
	offset, limit := h.getPagination(r)
	
	services, err := h.service.ListServices(offset, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to list services", err.Error())
		return
	}
	
	response := &ServicesListResponse{
		Services: services,
		Total:    len(services),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

func (h *ServicesHandler) ListPublishedServices(w http.ResponseWriter, r *http.Request) {
	offset, limit := h.getPagination(r)
	
	services, err := h.service.ListPublishedServices(offset, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to list published services", err.Error())
		return
	}
	
	response := &ServicesListResponse{
		Services: services,
		Total:    len(services),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

func (h *ServicesHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ServicesHandler) sendError(w http.ResponseWriter, status int, message, details string) {
	response := &ErrorResponse{
		Error:   message,
		Message: details,
	}
	h.sendJSON(w, status, response)
}

func (h *ServicesHandler) getUserID(r *http.Request) string {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		return "system"
	}
	return userID
}

func (h *ServicesHandler) getPagination(r *http.Request) (int, int) {
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