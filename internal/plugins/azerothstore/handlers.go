package azerothstore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/delivery"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/itemview"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothstore/domain"
)

var (
	errStoreUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"store_unavailable", "store is unavailable")
	errProductNotFound = httpapi.NewAPIError(http.StatusNotFound,
		"product_not_found", "product not found")
	errMissingSKU = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"missing_sku", "sku is required")
	errAccountNotLinked = httpapi.NewAPIError(http.StatusConflict,
		"account_not_linked", "link an AzerothCore account before purchasing")
	errInsufficientFunds = httpapi.NewAPIError(http.StatusPaymentRequired,
		"insufficient_funds", "not enough points")
	errNotOwner = httpapi.NewAPIError(http.StatusForbidden,
		"not_owner", "character does not belong to your account")
	errCharacterNotFound = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"character_not_found", "character does not exist")
	errInvalidDelivery = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_delivery", "delivery request is invalid")
	errDeliveryUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"delivery_unavailable", "delivery service is unavailable")
	errDeliveryFailed = httpapi.NewAPIError(http.StatusBadGateway,
		"delivery_failed", "delivery failed")
	errMissingUser = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"missing_user", "user_id or discord_id is required")
	errInvalidUserID = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_user_id", "user_id is not a valid UUID")
	errUserNotFound = httpapi.NewAPIError(http.StatusNotFound,
		"user_not_found", "community user not found")
	errInvalidPoints = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_points", "points must be greater than zero")
	errDirectoryUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"identity_storage_unavailable", "community user directory is unavailable")
	errUnknownItem = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"unknown_item", "one or more items do not exist")
	errProductExists = httpapi.NewAPIError(http.StatusConflict,
		"product_exists", "a product with this sku already exists")
	errInvalidItems = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_items", "each item needs a positive id and count")
	errEmptyProduct = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"empty_product", "provide items or money")
)

// PurchaseRequest is the body of POST /api/v1/store/orders.
type PurchaseRequest struct {
	SKU       string `json:"sku"`
	Character string `json:"character"`
} // @name StorePurchaseRequest

// GrantRequest is the body of POST /api/v1/store/wallets/grant.
type GrantRequest struct {
	UserID    string `json:"user_id"`
	DiscordID string `json:"discord_id"`
	Points    int64  `json:"points"`
	Reason    string `json:"reason"`
} // @name StoreGrantRequest

// ProductItem is a catalog product item with optional rendering.
type ProductItem struct {
	ItemID int            `json:"item_id"`
	Count  int            `json:"count"`
	Name   *string        `json:"name,omitempty"`
	Render *itemview.View `json:"render,omitempty"`
} // @name StoreProductItem

// Product is the JSON representation of a catalog product.
type Product struct {
	SKU         string        `json:"sku"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	PricePoints int64         `json:"price_points"`
	Money       int64         `json:"money"`
	Active      bool          `json:"active"`
	Items       []ProductItem `json:"items"`
} // @name StoreProduct

// ProductsResponse is the body of GET /api/v1/store/products.
type ProductsResponse struct {
	Products []Product `json:"products"`
} // @name StoreProductsResponse

// WalletResponse is the body of GET /api/v1/store/wallet and the grant endpoint.
type WalletResponse struct {
	UserID  string `json:"user_id"`
	Balance int64  `json:"balance"`
} // @name StoreWalletResponse

// Order is the JSON representation of a store order.
type Order struct {
	OrderID     string `json:"order_id"`
	UserID      string `json:"user_id"`
	SKU         string `json:"sku"`
	PricePoints int64  `json:"price_points"`
	Character   string `json:"character"`
	Status      string `json:"status"`
	Output      string `json:"output"`
	CreatedAt   string `json:"created_at"`
} // @name StoreOrder

// OrdersResponse is the body of GET /api/v1/store/orders.
type OrdersResponse struct {
	Orders []Order `json:"orders"`
} // @name StoreOrdersResponse

// handleProducts handles GET /api/v1/store/products.
//
//	@Summary		List products
//	@Description	Lists the store catalog. Requires the store.catalog.read permission.
//	@Tags			store
//	@ID				store.products.list
//	@Produce		json
//	@Success		200	{object}	ProductsResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/store/products [get]
func (p *Plugin) handleProducts(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	products, err := p.store.Products(r.Context())
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	items := make([]Product, 0, len(products))
	for _, product := range products {
		items = append(items, p.productDTO(r.Context(), product))
	}
	httpapi.WriteJSON(w, http.StatusOK, ProductsResponse{Products: items})
}

// handleProduct handles GET /api/v1/store/products/{sku}.
//
//	@Summary		Get a product
//	@Description	Returns one catalog product. Requires the store.catalog.read permission.
//	@Tags			store
//	@ID				store.products.get
//	@Produce		json
//	@Param			sku	path	string	true	"product SKU"
//	@Success		200	{object}	Product
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/store/products/{sku} [get]
func (p *Plugin) handleProduct(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	product, err := p.store.ProductBySKU(r.Context(), r.PathValue("sku"))
	if errors.Is(err, domain.ErrProductNotFound) {
		httpapi.WriteError(w, r, errProductNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, p.productDTO(r.Context(), product))
}

// handleWallet handles GET /api/v1/store/wallet.
//
//	@Summary		Get the wallet
//	@Description	Returns the authenticated user's point balance. Requires the store.wallet.read permission.
//	@Tags			store
//	@ID				store.wallet.get
//	@Produce		json
//	@Success		200	{object}	WalletResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/store/wallet [get]
func (p *Plugin) handleWallet(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	balance, err := p.store.Wallet(r.Context(), principal.UserID)
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, WalletResponse{UserID: principal.UserID.String(), Balance: balance})
}

// handleOrders handles GET /api/v1/store/orders.
//
//	@Summary		List orders
//	@Description	Lists the authenticated user's orders. Requires the store.orders.read permission.
//	@Tags			store
//	@ID				store.orders.list
//	@Produce		json
//	@Param			limit	query	int	false	"page size"	default(50)
//	@Param			offset	query	int	false	"page offset"	default(0)
//	@Success		200	{object}	OrdersResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/store/orders [get]
func (p *Plugin) handleOrders(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	limit := intParam(r, "limit", 50)
	offset := intParam(r, "offset", 0)
	orders, err := p.store.Orders(r.Context(), principal.UserID, limit, offset)
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	items := make([]Order, 0, len(orders))
	for _, order := range orders {
		items = append(items, orderDTO(order))
	}
	httpapi.WriteJSON(w, http.StatusOK, OrdersResponse{Orders: items})
}

// handleAdminOrders handles GET /api/v1/admin/store/orders.
//
//	@Summary		List all orders
//	@Description	Lists every store order, newest first, optionally filtered by status. Requires the store.admin.orders.read permission.
//	@Tags			store
//	@ID				store.admin.orders.list
//	@Produce		json
//	@Param			status	query	string	false	"order status (pending, delivered, failed)"
//	@Param			limit	query	int		false	"page size"	default(50)
//	@Param			offset	query	int		false	"page offset"	default(0)
//	@Success		200	{object}	OrdersResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/store/orders [get]
func (p *Plugin) handleAdminOrders(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	limit := intParam(r, "limit", 50)
	offset := intParam(r, "offset", 0)
	orders, err := p.store.AdminOrders(r.Context(), status, limit, offset)
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	items := make([]Order, 0, len(orders))
	for _, order := range orders {
		items = append(items, orderDTO(order))
	}
	httpapi.WriteJSON(w, http.StatusOK, OrdersResponse{Orders: items})
}

// handlePurchase handles POST /api/v1/store/orders.
//
//	@Summary		Purchase a product
//	@Description	Buys a product and delivers it to a character. Requires the store.purchase permission.
//	@Tags			store
//	@ID				store.orders.create
//	@Accept			json
//	@Produce		json
//	@Param			request	body	PurchaseRequest	true	"purchase request"
//	@Success		201	{object}	Order
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		402	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		409	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/store/orders [post]
func (p *Plugin) handlePurchase(w http.ResponseWriter, r *http.Request) {
	if p.store == nil || p.delivery == nil || p.accounts == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	var req PurchaseRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	sku := strings.TrimSpace(req.SKU)
	character := strings.TrimSpace(req.Character)
	if sku == "" {
		httpapi.WriteError(w, r, errMissingSKU)
		return
	}

	principal, _ := auth.PrincipalFromContext(r.Context())
	product, err := p.store.ProductBySKU(r.Context(), sku)
	if errors.Is(err, domain.ErrProductNotFound) {
		httpapi.WriteError(w, r, errProductNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	if !product.Active {
		httpapi.WriteError(w, r, errProductNotFound)
		return
	}

	_, accountID, err := p.accounts.LinkedAccount(r.Context(), principal.UserID.String())
	if err != nil || accountID == nil {
		httpapi.WriteError(w, r, errAccountNotLinked)
		return
	}

	order, err := p.store.CreateOrder(r.Context(), principal.UserID, product, character, *accountID)
	if errors.Is(err, domain.ErrInsufficientFunds) {
		httpapi.WriteError(w, r, errInsufficientFunds)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}

	request := delivery.Request{
		AccountID: *accountID,
		Character: character,
		Subject:   product.Name,
		Body:      product.Description,
		Money:     product.Money,
	}
	for _, item := range product.Items {
		request.Items = append(request.Items, delivery.Item{ID: item.ItemID, Count: item.Count})
	}

	output, deliveryErr := p.delivery.Deliver(r.Context(), request)
	if deliveryErr != nil {
		if _, failErr := p.store.FailOrder(r.Context(), order.ID, deliveryErr.Error()); failErr != nil {
			_ = p.store.CompleteOrder(r.Context(), order.ID, "refund failed: "+failErr.Error())
		}
		p.record(r, "store.purchase", order, audit.ResultFailure)
		writeDeliveryError(w, r, deliveryErr)
		return
	}
	if err := p.completeOrderWithRetry(r.Context(), order.ID, output); err != nil {
		_ = p.store.SetOrderOutput(r.Context(), order.ID, output)
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	order.Status = domain.OrderDelivered
	order.CommandOutput = output
	p.record(r, "store.purchase", order, audit.ResultSuccess)
	httpapi.WriteJSON(w, http.StatusCreated, orderDTO(order))
}

// completeOrderWithRetry retries a transient completion failure a few times. A
// terminal-state error is returned immediately (the order was already handled).
func (p *Plugin) completeOrderWithRetry(ctx context.Context, orderID uuid.UUID, output string) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if err = p.store.CompleteOrder(ctx, orderID, output); err == nil {
			return nil
		}
		if errors.Is(err, domain.ErrOrderNotPending) || errors.Is(err, domain.ErrOrderNotFound) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 100 * time.Millisecond):
		}
	}
	return err
}

// handleGrant handles POST /api/v1/store/wallets/grant.
//
//	@Summary		Grant wallet points
//	@Description	Adds points to a community user's wallet. Requires the store.admin.wallets permission.
//	@Tags			store
//	@ID				store.wallets.grant
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GrantRequest	true	"grant request"
//	@Success		200	{object}	WalletResponse
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/store/wallets/grant [post]
func (p *Plugin) handleGrant(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	var req GrantRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	if req.Points <= 0 {
		httpapi.WriteError(w, r, errInvalidPoints)
		return
	}
	userID, err := p.resolveUser(r, req)
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "admin grant"
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	balance, err := p.store.Grant(r.Context(), userID, req.Points, reason, principal.UserID)
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	_ = p.audit.Record(r.Context(), audit.Entry{
		Timestamp:      time.Now(),
		ActorID:        principal.UserID,
		ActorDiscordID: principal.DiscordID,
		Action:         "store.wallet.grant",
		Permission:     string(PermissionAdminWallets),
		TargetType:     "community_user",
		TargetID:       userID.String(),
		Result:         audit.ResultSuccess,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})
	httpapi.WriteJSON(w, http.StatusOK, WalletResponse{UserID: userID.String(), Balance: balance})
}

func (p *Plugin) resolveUser(r *http.Request, req GrantRequest) (uuid.UUID, error) {
	if value := strings.TrimSpace(req.UserID); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			return uuid.Nil, errInvalidUserID
		}
		return id, nil
	}
	if value := strings.TrimSpace(req.DiscordID); value != "" {
		if p.users == nil {
			return uuid.Nil, errDirectoryUnavailable
		}
		id, found, err := p.users.ResolveUser(r.Context(), value)
		if err != nil {
			return uuid.Nil, errDirectoryUnavailable
		}
		if !found {
			return uuid.Nil, errUserNotFound
		}
		return id, nil
	}
	return uuid.Nil, errMissingUser
}

func (p *Plugin) record(r *http.Request, action string, order domain.Order, result audit.Result) {
	principal, _ := auth.PrincipalFromContext(r.Context())
	_ = p.audit.Record(r.Context(), audit.Entry{
		Timestamp:      time.Now(),
		ActorID:        principal.UserID,
		ActorDiscordID: principal.DiscordID,
		Action:         action,
		Permission:     string(PermissionPurchase),
		TargetType:     "store_order",
		TargetID:       order.ID.String(),
		Result:         result,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})
}

func writeDeliveryError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, delivery.ErrNotOwner):
		httpapi.WriteError(w, r, errNotOwner)
	case errors.Is(err, delivery.ErrCharacterNotFound):
		httpapi.WriteError(w, r, errCharacterNotFound)
	case errors.Is(err, delivery.ErrInvalidRequest):
		httpapi.WriteError(w, r, errInvalidDelivery)
	case errors.Is(err, azerothdb.ErrUnavailable):
		httpapi.WriteError(w, r, errDeliveryUnavailable)
	default:
		httpapi.WriteError(w, r, errDeliveryFailed)
	}
}

func writeGrantError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *httpapi.APIError
	if errors.As(err, &apiErr) {
		httpapi.WriteError(w, r, apiErr)
		return
	}
	httpapi.WriteError(w, r, httpapi.ErrInternal)
}

func (p *Plugin) productDTO(ctx context.Context, product domain.Product) Product {
	items := make([]ProductItem, 0, len(product.Items))
	for _, item := range product.Items {
		entry := ProductItem{ItemID: item.ItemID, Count: item.Count}
		if p.catalog != nil {
			if resolved, found, err := p.catalog.LookupItem(ctx, int64(item.ItemID)); err == nil && found {
				name := resolved.Name
				render := itemview.Build(resolved)
				entry.Name = &name
				entry.Render = &render
			}
		}
		items = append(items, entry)
	}
	return Product{
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		PricePoints: product.PricePoints,
		Money:       product.Money,
		Active:      product.Active,
		Items:       items,
	}
}

// ProductItemRequest is one item stack in a product write request.
type ProductItemRequest struct {
	ItemID int `json:"item_id"`
	Count  int `json:"count"`
} // @name StoreProductItemRequest

// ProductRequest is the body of the product write endpoints.
type ProductRequest struct {
	SKU         string               `json:"sku"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	PricePoints int64                `json:"price_points"`
	Money       int64                `json:"money"`
	Active      *bool                `json:"active"`
	Items       []ProductItemRequest `json:"items"`
} // @name StoreProductRequest

// handleCreateProduct handles POST /api/v1/store/products.
//
//	@Summary		Create a product
//	@Description	Creates a catalog product. Requires the store.admin.products permission.
//	@Tags			store
//	@ID				store.products.create
//	@Accept			json
//	@Produce		json
//	@Param			request	body	ProductRequest	true	"product to create"
//	@Success		201	{object}	Product
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		409	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/store/products [post]
func (p *Plugin) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	var req ProductRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	product, err := p.buildProduct(r.Context(), req, "")
	if err != nil {
		writeProductError(w, r, err)
		return
	}
	created, err := p.store.CreateProduct(r.Context(), product)
	if errors.Is(err, domain.ErrProductExists) {
		httpapi.WriteError(w, r, errProductExists)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, p.productDTO(r.Context(), created))
}

// handleUpdateProduct handles PUT /api/v1/store/products/{sku}.
//
//	@Summary		Update a product
//	@Description	Updates a catalog product. Requires the store.admin.products permission.
//	@Tags			store
//	@ID				store.products.update
//	@Accept			json
//	@Produce		json
//	@Param			sku		path	string			true	"product SKU"
//	@Param			request	body	ProductRequest	true	"product to update"
//	@Success		200	{object}	Product
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/store/products/{sku} [put]
func (p *Plugin) handleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	var req ProductRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	product, err := p.buildProduct(r.Context(), req, r.PathValue("sku"))
	if err != nil {
		writeProductError(w, r, err)
		return
	}
	updated, err := p.store.UpdateProduct(r.Context(), product)
	if errors.Is(err, domain.ErrProductNotFound) {
		httpapi.WriteError(w, r, errProductNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, p.productDTO(r.Context(), updated))
}

// handleDeleteProduct handles DELETE /api/v1/store/products/{sku}.
//
//	@Summary		Deactivate a product
//	@Description	Deactivates a catalog product. Requires the store.admin.products permission.
//	@Tags			store
//	@ID				store.products.delete
//	@Produce		json
//	@Param			sku	path	string	true	"product SKU"
//	@Success		200	{object}	Product
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/store/products/{sku} [delete]
func (p *Plugin) handleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	product, err := p.store.SetProductActive(r.Context(), r.PathValue("sku"), false)
	if errors.Is(err, domain.ErrProductNotFound) {
		httpapi.WriteError(w, r, errProductNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errStoreUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, p.productDTO(r.Context(), product))
}

// buildProduct validates a product request, resolves item entries through the
// catalog and applies item-derived defaults for the name and description.
func (p *Plugin) buildProduct(ctx context.Context, req ProductRequest, sku string) (domain.Product, error) {
	if sku == "" {
		sku = strings.TrimSpace(req.SKU)
	}
	product := domain.Product{
		SKU:         sku,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		PricePoints: req.PricePoints,
		Money:       req.Money,
		Active:      true,
	}
	if req.Active != nil {
		product.Active = *req.Active
	}
	if product.PricePoints < 0 || product.Money < 0 {
		return domain.Product{}, errInvalidItems
	}
	if len(req.Items) == 0 && product.Money == 0 {
		return domain.Product{}, errEmptyProduct
	}

	for _, item := range req.Items {
		if item.ItemID <= 0 || item.Count <= 0 {
			return domain.Product{}, errInvalidItems
		}
		product.Items = append(product.Items, domain.ProductItem{ItemID: item.ItemID, Count: item.Count})
	}

	if p.catalog != nil {
		for _, item := range product.Items {
			resolved, found, err := p.catalog.LookupItem(ctx, int64(item.ItemID))
			if err != nil {
				return domain.Product{}, errStoreUnavailable
			}
			if !found {
				return domain.Product{}, errUnknownItem
			}
			if len(product.Items) == 1 {
				if product.Name == "" {
					product.Name = resolved.Name
				}
				if product.Description == "" {
					product.Description = resolved.Description
				}
			}
		}
	}

	if product.SKU == "" {
		if len(product.Items) == 1 {
			product.SKU = fmt.Sprintf("item-%d", product.Items[0].ItemID)
		} else {
			return domain.Product{}, errMissingSKU
		}
	}
	if product.Name == "" {
		product.Name = product.SKU
	}
	return product, nil
}

func writeProductError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *httpapi.APIError
	if errors.As(err, &apiErr) {
		httpapi.WriteError(w, r, apiErr)
		return
	}
	httpapi.WriteError(w, r, httpapi.ErrInternal)
}

func orderDTO(order domain.Order) Order {
	return Order{
		OrderID:     order.ID.String(),
		UserID:      order.UserID.String(),
		SKU:         order.SKU,
		PricePoints: order.PricePoints,
		Character:   order.CharacterName,
		Status:      string(order.Status),
		Output:      order.CommandOutput,
		CreatedAt:   order.CreatedAt.UTC().Format(time.RFC3339),
	}
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
