package azerothstore

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/delivery"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothstore/domain"
)

var (
	errOrderNotResolvable = httpapi.NewAPIError(http.StatusConflict,
		"order_not_pending", "the order is not pending")
	errOrderMissing = httpapi.NewAPIError(http.StatusNotFound,
		"order_not_found", "order not found")
)

// ResolveOrderRequest is the body of the order resolve endpoints.
type ResolveOrderRequest struct {
	Reason string `json:"reason"`
} // @name StoreResolveOrderRequest

// handleRefundOrder handles POST /api/v1/admin/store/orders/{id}/refund.
//
//	@Summary		Refund a stuck order
//	@Description	Marks a pending order failed and refunds its points. Requires the store.admin.orders.resolve permission.
//	@Tags			store
//	@ID				store.admin.orders.refund
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string				true	"order id"
//	@Param			request	body	ResolveOrderRequest	false	"reason"
//	@Success		200	{object}	Order
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		409	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/store/orders/{id}/refund [post]
func (p *Plugin) handleRefundOrder(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	var req ResolveOrderRequest
	_ = httpapi.DecodeJSON(r, &req)
	order, err := p.store.OrderByID(r.Context(), id)
	if errors.Is(err, domain.ErrOrderNotFound) {
		httpapi.WriteError(w, r, errOrderMissing)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	if order.Status != domain.OrderPending {
		httpapi.WriteError(w, r, errOrderNotResolvable)
		return
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "manual refund"
	}
	refunded, err := p.store.FailOrder(r.Context(), id, "manual refund: "+reason)
	if errors.Is(err, domain.ErrOrderNotPending) {
		httpapi.WriteError(w, r, errOrderNotResolvable)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	p.record(r, "store.order.refund", refunded, audit.ResultSuccess)
	httpapi.WriteJSON(w, http.StatusOK, orderDTO(refunded))
}

// handleRetryOrder handles POST /api/v1/admin/store/orders/{id}/retry.
//
//	@Summary		Retry a stuck order
//	@Description	Re-delivers a pending order's reward and completes it. Requires the store.admin.orders.resolve permission.
//	@Tags			store
//	@ID				store.admin.orders.retry
//	@Produce		json
//	@Param			id	path	string	true	"order id"
//	@Success		200	{object}	Order
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		409	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/store/orders/{id}/retry [post]
func (p *Plugin) handleRetryOrder(w http.ResponseWriter, r *http.Request) {
	if p.store == nil || p.delivery == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	order, err := p.store.OrderByID(r.Context(), id)
	if errors.Is(err, domain.ErrOrderNotFound) {
		httpapi.WriteError(w, r, errOrderMissing)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	if order.Status != domain.OrderPending {
		httpapi.WriteError(w, r, errOrderNotResolvable)
		return
	}
	product, err := p.store.ProductBySKU(r.Context(), order.SKU)
	if errors.Is(err, domain.ErrProductNotFound) {
		httpapi.WriteError(w, r, errProductNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	request := delivery.Request{
		AccountID: order.AccountID,
		Character: order.CharacterName,
		Subject:   product.Name,
		Body:      product.Description,
		Money:     product.Money,
	}
	for _, item := range product.Items {
		request.Items = append(request.Items, delivery.Item{ID: item.ItemID, Count: item.Count})
	}
	output, deliveryErr := p.delivery.Deliver(r.Context(), request)
	if deliveryErr != nil {
		p.record(r, "store.order.retry", order, audit.ResultFailure)
		writeDeliveryError(w, r, deliveryErr)
		return
	}
	if err := p.store.CompleteOrder(r.Context(), id, output); err != nil {
		_ = p.store.SetOrderOutput(r.Context(), id, output)
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	order.Status = domain.OrderDelivered
	order.CommandOutput = output
	p.record(r, "store.order.retry", order, audit.ResultSuccess)
	httpapi.WriteJSON(w, http.StatusOK, orderDTO(order))
}
