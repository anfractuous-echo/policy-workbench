package httpapi

import (
	"net/http"

	"policy-workbench/internal/application"
	"policy-workbench/internal/domain"
)

func (handler *Handler) workspace(w http.ResponseWriter, r *http.Request) {
	view, err := handler.service.Workspace(r.URL.Query().Get("profile_id"))
	writeResult(w, view, err, http.StatusOK)
}

func (handler *Handler) createPolicy(w http.ResponseWriter, r *http.Request) {
	var policy domain.Policy
	if !decodeJSON(w, r, &policy) {
		return
	}
	created, err := handler.service.CreatePolicy(policy)
	writeResult(w, created, err, http.StatusCreated)
}

func (handler *Handler) deletePolicy(w http.ResponseWriter, r *http.Request) {
	err := handler.service.DeletePolicy(r.URL.Query().Get("profile_id"), r.PathValue("id"))
	writeResult(w, map[string]bool{"ok": true}, err, http.StatusOK)
}

func (handler *Handler) reorderPolicies(w http.ResponseWriter, r *http.Request) {
	var command application.ReorderPoliciesCommand
	if !decodeJSON(w, r, &command) {
		return
	}
	err := handler.service.ReorderPolicies(command)
	writeResult(w, map[string]bool{"ok": true}, err, http.StatusOK)
}

func (handler *Handler) createTransfer(w http.ResponseWriter, r *http.Request) {
	var request domain.TransferRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	transfer, err := handler.service.CreateTransfer(request)
	writeResult(w, transfer, err, http.StatusCreated)
}

func (handler *Handler) applyTransferAction(w http.ResponseWriter, r *http.Request) {
	var command application.TransferActionCommand
	if !decodeJSON(w, r, &command) {
		return
	}
	transfer, err := handler.service.ApplyTransferAction(r.PathValue("id"), command)
	writeResult(w, transfer, err, http.StatusOK)
}

func (handler *Handler) advanceClock(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Hours int `json:"hours"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	now, err := handler.service.AdvanceClock(request.Hours)
	writeResult(w, map[string]any{"now": now}, err, http.StatusOK)
}

func (handler *Handler) resetWorkspace(w http.ResponseWriter, r *http.Request) {
	var request struct{}
	if !decodeJSON(w, r, &request) {
		return
	}
	err := handler.service.ResetWorkspace()
	writeResult(w, map[string]bool{"ok": true}, err, http.StatusOK)
}
