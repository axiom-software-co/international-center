package media

import (
	"time"
)

// ListInquiriesParams represents parameters for listing inquiries in a contract-compliant way
type ListInquiriesParams struct {
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
	Search      string `json:"search,omitempty"`
	InquiryType string `json:"inquiry_type,omitempty"`
	Status      string `json:"status,omitempty"`
}

// AdminUpdateInquiryStatusRequest represents a request to update inquiry status
type AdminUpdateInquiryStatusRequest struct {
	Status     InquiryStatus `json:"status"`
	Notes      *string       `json:"notes,omitempty"`
	AssignedTo *string       `json:"assigned_to,omitempty"`
}

// PaginationResult represents pagination information
type PaginationResult struct {
	CurrentPage  int  `json:"current_page"`
	TotalPages   int  `json:"total_pages"`
	TotalItems   int  `json:"total_items"`
	ItemsPerPage int  `json:"items_per_page"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
}

// Inquiry represents a contract-compliant inquiry for API responses
type Inquiry struct {
	ID             string    `json:"id"`
	InquiryType    string    `json:"inquiry_type"`
	Status         string    `json:"status"`
	SubmitterName  string    `json:"submitter_name"`
	SubmitterEmail string    `json:"submitter_email"`
	Subject        string    `json:"subject"`
	Message        string    `json:"message"`
	SubmittedOn    time.Time `json:"submitted_on"`
	ModifiedOn     time.Time `json:"modified_on"`
	Notes          *string   `json:"notes,omitempty"`
	AssignedTo     *string   `json:"assigned_to,omitempty"`
}