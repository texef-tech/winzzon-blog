package public

import (
	"net/http"
	"strconv"

	"github.com/texef-tech/winzzon-blog/internal/handler"
	"github.com/texef-tech/winzzon-blog/internal/service"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(ss *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: ss}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		handler.ErrorJSON(w, http.StatusBadRequest, "MISSING_QUERY", "Search query parameter 'q' is required")
		return
	}

	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	result, err := h.searchService.Search(r.Context(), query, limit, offset)
	if err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "SEARCH_ERROR", "Search failed")
		return
	}

	handler.JSON(w, http.StatusOK, result)
}
