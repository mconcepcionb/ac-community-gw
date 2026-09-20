package azerothitem

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/itemview"
)

var (
	errItemDBNotConfigured = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"item_db_not_configured", "AzerothCore world database is not configured")
	errItemDBUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"item_db_unavailable", "AzerothCore world database is unavailable")
	errItemNotFound = httpapi.NewAPIError(http.StatusNotFound,
		"item_not_found", "item not found")
	errInvalidEntry = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_entry", "entry must be a positive integer")
)

// ItemsResponse is the body of GET /api/v1/azeroth/items.
type ItemsResponse struct {
	Items []itemview.View `json:"items"`
} // @name AzerothItemsResponse

// handleListItems handles GET /api/v1/azeroth/items.
//
//	@Summary		Search items
//	@Description	Searches the AzerothCore item catalog. Requires the azeroth.item.list permission.
//	@Tags			azeroth-item
//	@ID				azeroth.items.list
//	@Produce		json
//	@Param			filter	query	string	false	"item name filter"
//	@Param			class	query	int		false	"item class id"
//	@Param			limit	query	int		false	"page size"	default(50)
//	@Param			offset	query	int		false	"page offset"	default(0)
//	@Success		200	{object}	ItemsResponse
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/items [get]
func (p *Plugin) handleListItems(w http.ResponseWriter, r *http.Request) {
	if p.items == nil {
		httpapi.WriteError(w, r, errItemDBNotConfigured)
		return
	}
	query := azerothdb.ItemQuery{
		Filter: r.URL.Query().Get("filter"),
		Limit:  intParam(r, "limit", 50),
		Offset: intParam(r, "offset", 0),
	}
	if raw := r.URL.Query().Get("class"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpapi.WriteError(w, r, httpapi.ErrBadRequest)
			return
		}
		query.Class = parsed
	}
	items, err := p.items.ListItems(r.Context(), query)
	if err != nil {
		httpapi.WriteError(w, r, errItemDBUnavailable)
		return
	}
	out := make([]itemview.View, 0, len(items))
	for _, item := range items {
		out = append(out, itemview.Build(item))
	}
	httpapi.WriteJSON(w, http.StatusOK, ItemsResponse{Items: out})
}

// handleGetItem handles GET /api/v1/azeroth/items/{entry}.
//
//	@Summary		Get an item
//	@Description	Returns one item template by entry id. Requires the azeroth.item.list permission.
//	@Tags			azeroth-item
//	@ID				azeroth.items.get
//	@Produce		json
//	@Param			entry	path	int	true	"item entry id"
//	@Success		200	{object}	itemview.View
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/items/{entry} [get]
func (p *Plugin) handleGetItem(w http.ResponseWriter, r *http.Request) {
	if p.items == nil {
		httpapi.WriteError(w, r, errItemDBNotConfigured)
		return
	}
	entry, err := strconv.ParseInt(r.PathValue("entry"), 10, 64)
	if err != nil || entry <= 0 {
		httpapi.WriteError(w, r, errInvalidEntry)
		return
	}
	item, err := p.items.FindItem(r.Context(), entry)
	if errors.Is(err, azerothdb.ErrItemNotFound) {
		httpapi.WriteError(w, r, errItemNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errItemDBUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, itemview.Build(item))
}

func intParam(r *http.Request, name string, fallback int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return parsed
}
