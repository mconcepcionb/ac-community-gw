// Package repository implements the store persistence.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/store/domain"
	storerepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/store/repository/generated"
)

// Store implements the store persistence operations.
type Store struct {
	db *sql.DB
	q  *storerepo.Queries
}

// New creates a store over an open database pool.
func New(db *sql.DB) *Store {
	return &Store{db: db, q: storerepo.New(db)}
}

// Products returns the active catalog.
func (s *Store) Products(ctx context.Context) ([]domain.Product, error) {
	rows, err := s.q.ListActiveProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: list products: %w", err)
	}
	products := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		product, err := s.productWithItems(ctx, row)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

// ProductBySKU returns a product or domain.ErrProductNotFound.
func (s *Store) ProductBySKU(ctx context.Context, sku string) (domain.Product, error) {
	row, err := s.q.GetProductBySKU(ctx, sku)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Product{}, domain.ErrProductNotFound
	}
	if err != nil {
		return domain.Product{}, fmt.Errorf("repository: get product: %w", err)
	}
	return s.productWithItems(ctx, row)
}

// CreateProduct inserts a product and its items in one transaction. A product
// created without an explicit id is assigned a fresh UUID.
func (s *Store) CreateProduct(ctx context.Context, product domain.Product) (domain.Product, error) {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Product{}, fmt.Errorf("repository: begin product: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	q := s.q.WithTx(tx)

	row, err := q.InsertProduct(ctx, storerepo.InsertProductParams{
		ID:          product.ID,
		Sku:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		PricePoints: product.PricePoints,
		Money:       product.Money,
		Active:      product.Active,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Product{}, domain.ErrProductExists
		}
		return domain.Product{}, fmt.Errorf("repository: insert product: %w", err)
	}
	if err := insertProductItems(ctx, q, row.ID, product.Items); err != nil {
		return domain.Product{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Product{}, fmt.Errorf("repository: commit product: %w", err)
	}
	return s.ProductBySKU(ctx, row.Sku)
}

// UpdateProduct replaces a product's fields and items in one transaction.
func (s *Store) UpdateProduct(ctx context.Context, product domain.Product) (domain.Product, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Product{}, fmt.Errorf("repository: begin product update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	q := s.q.WithTx(tx)

	row, err := q.UpdateProduct(ctx, storerepo.UpdateProductParams{
		Sku:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		PricePoints: product.PricePoints,
		Money:       product.Money,
		Active:      product.Active,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Product{}, domain.ErrProductNotFound
	}
	if err != nil {
		return domain.Product{}, fmt.Errorf("repository: update product: %w", err)
	}
	if err := q.DeleteProductItems(ctx, row.ID); err != nil {
		return domain.Product{}, fmt.Errorf("repository: clear product items: %w", err)
	}
	if err := insertProductItems(ctx, q, row.ID, product.Items); err != nil {
		return domain.Product{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Product{}, fmt.Errorf("repository: commit product update: %w", err)
	}
	return s.ProductBySKU(ctx, row.Sku)
}

// SetProductActive enables or disables a product.
func (s *Store) SetProductActive(ctx context.Context, sku string, active bool) (domain.Product, error) {
	row, err := s.q.SetProductActive(ctx, storerepo.SetProductActiveParams{Sku: sku, Active: active})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Product{}, domain.ErrProductNotFound
	}
	if err != nil {
		return domain.Product{}, fmt.Errorf("repository: set product active: %w", err)
	}
	return s.productWithItems(ctx, row)
}

func insertProductItems(ctx context.Context, q *storerepo.Queries, productID uuid.UUID, items []domain.ProductItem) error {
	for _, item := range items {
		if err := q.InsertProductItem(ctx, storerepo.InsertProductItemParams{
			ProductID: productID,
			ItemID:    int32(item.ItemID),
			Count:     int32(item.Count),
		}); err != nil {
			return fmt.Errorf("repository: insert product item: %w", err)
		}
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Wallet returns the user's balance (0 when no wallet exists yet).
func (s *Store) Wallet(ctx context.Context, userID uuid.UUID) (int64, error) {
	balance, err := s.q.GetWallet(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("repository: get wallet: %w", err)
	}
	return balance, nil
}

// Grant credits points to the user and records a ledger entry.
func (s *Store) Grant(ctx context.Context, userID uuid.UUID, points int64, reason string, actorID uuid.UUID) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("repository: begin grant: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	q := s.q.WithTx(tx)

	if err := q.EnsureWallet(ctx, userID); err != nil {
		return 0, fmt.Errorf("repository: ensure wallet: %w", err)
	}
	balance, err := q.CreditWallet(ctx, storerepo.CreditWalletParams{UserID: userID, Balance: points})
	if err != nil {
		return 0, fmt.Errorf("repository: credit wallet: %w", err)
	}
	if err := q.InsertWalletEntry(ctx, storerepo.InsertWalletEntryParams{
		UserID:       userID,
		Delta:        points,
		BalanceAfter: balance,
		Reason:       reason,
		ActorID:      nullUUID(actorID),
	}); err != nil {
		return 0, fmt.Errorf("repository: insert wallet entry: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("repository: commit grant: %w", err)
	}
	return balance, nil
}

// CreateOrder debits the price and records a pending order in one transaction.
func (s *Store) CreateOrder(ctx context.Context, userID uuid.UUID, product domain.Product, character string, accountID int64) (domain.Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Order{}, fmt.Errorf("repository: begin order: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	q := s.q.WithTx(tx)

	if err := q.EnsureWallet(ctx, userID); err != nil {
		return domain.Order{}, fmt.Errorf("repository: ensure wallet: %w", err)
	}
	balance, err := q.DebitWallet(ctx, storerepo.DebitWalletParams{UserID: userID, Balance: product.PricePoints})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Order{}, domain.ErrInsufficientFunds
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("repository: debit wallet: %w", err)
	}

	orderID := uuid.New()
	row, err := q.InsertOrder(ctx, storerepo.InsertOrderParams{
		ID:            orderID,
		UserID:        userID,
		ProductID:     product.ID,
		Sku:           product.SKU,
		PricePoints:   product.PricePoints,
		CharacterName: character,
		AccountID:     accountID,
		Status:        string(domain.OrderPending),
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf("repository: insert order: %w", err)
	}
	if err := q.InsertWalletEntry(ctx, storerepo.InsertWalletEntryParams{
		UserID:       userID,
		Delta:        -product.PricePoints,
		BalanceAfter: balance,
		Reason:       "purchase:" + product.SKU,
		OrderID:      nullUUID(orderID),
	}); err != nil {
		return domain.Order{}, fmt.Errorf("repository: insert wallet entry: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.Order{}, fmt.Errorf("repository: commit order: %w", err)
	}
	return toOrder(row), nil
}

// CompleteOrder marks a pending order delivered. Completing an already
// delivered order is an idempotent no-op; any other terminal state is rejected.
func (s *Store) CompleteOrder(ctx context.Context, orderID uuid.UUID, output string) error {
	_, err := s.q.UpdateOrderStatus(ctx, storerepo.UpdateOrderStatusParams{
		ID:            orderID,
		Status:        string(domain.OrderDelivered),
		CommandOutput: output,
	})
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("repository: complete order: %w", err)
	}
	order, getErr := s.q.GetOrder(ctx, orderID)
	if errors.Is(getErr, sql.ErrNoRows) {
		return domain.ErrOrderNotFound
	}
	if getErr != nil {
		return fmt.Errorf("repository: get order: %w", getErr)
	}
	if domain.OrderStatus(order.Status) == domain.OrderDelivered {
		return nil
	}
	return domain.ErrOrderNotPending
}

// FailOrder marks a pending order failed and refunds the points in one
// transaction. It is idempotent: a second call does not refund again and
// returns ErrOrderNotPending.
func (s *Store) FailOrder(ctx context.Context, orderID uuid.UUID, output string) (domain.Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Order{}, fmt.Errorf("repository: begin fail: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	q := s.q.WithTx(tx)

	order, err := q.GetOrder(ctx, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("repository: get order: %w", err)
	}
	if domain.OrderStatus(order.Status) != domain.OrderPending {
		return domain.Order{}, domain.ErrOrderNotPending
	}
	if err := q.EnsureWallet(ctx, order.UserID); err != nil {
		return domain.Order{}, fmt.Errorf("repository: ensure wallet: %w", err)
	}

	row, err := q.UpdateOrderStatus(ctx, storerepo.UpdateOrderStatusParams{
		ID:            orderID,
		Status:        string(domain.OrderFailed),
		CommandOutput: output,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotPending
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("repository: fail order: %w", err)
	}
	balance, err := q.CreditWallet(ctx, storerepo.CreditWalletParams{UserID: row.UserID, Balance: row.PricePoints})
	if err != nil {
		return domain.Order{}, fmt.Errorf("repository: refund wallet: %w", err)
	}
	if err := q.InsertWalletEntry(ctx, storerepo.InsertWalletEntryParams{
		UserID:       row.UserID,
		Delta:        row.PricePoints,
		BalanceAfter: balance,
		Reason:       "refund:" + row.Sku,
		OrderID:      nullUUID(row.ID),
	}); err != nil {
		return domain.Order{}, fmt.Errorf("repository: insert refund entry: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.Order{}, fmt.Errorf("repository: commit fail: %w", err)
	}
	return toOrder(row), nil
}

// SetOrderOutput persists the delivery output of a still-pending order without
// changing its status. It is used when CompleteOrder fails so reconciliation can
// finish the order later.
func (s *Store) SetOrderOutput(ctx context.Context, orderID uuid.UUID, output string) error {
	if err := s.q.SetOrderOutput(ctx, storerepo.SetOrderOutputParams{
		ID:            orderID,
		CommandOutput: output,
	}); err != nil {
		return fmt.Errorf("repository: set order output: %w", err)
	}
	return nil
}

// ReconcilePendingOrders completes pending orders whose delivery output is
// already known. It never refunds and never touches terminal orders.
func (s *Store) ReconcilePendingOrders(ctx context.Context, limit int) (int, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.q.ListPendingOrdersWithOutput(ctx, int32(limit))
	if err != nil {
		return 0, fmt.Errorf("repository: list pending orders: %w", err)
	}
	completed := 0
	for _, row := range rows {
		if _, err := s.q.UpdateOrderStatus(ctx, storerepo.UpdateOrderStatusParams{
			ID:            row.ID,
			Status:        string(domain.OrderDelivered),
			CommandOutput: row.CommandOutput,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return completed, fmt.Errorf("repository: reconcile order %s: %w", row.ID, err)
		}
		completed++
	}
	return completed, nil
}

// Orders returns the user's orders, newest first.
func (s *Store) Orders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Order, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.q.ListOrdersByUser(ctx, storerepo.ListOrdersByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("repository: list orders: %w", err)
	}
	orders := make([]domain.Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, toOrder(row))
	}
	return orders, nil
}

// AdminOrders lists all orders, newest first, optionally filtered by status.
func (s *Store) AdminOrders(ctx context.Context, status string, limit, offset int) ([]domain.Order, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.q.ListOrders(ctx, storerepo.ListOrdersParams{
		Limit:   int32(limit),
		Offset:  int32(offset),
		Column3: status,
	})
	if err != nil {
		return nil, fmt.Errorf("repository: list orders: %w", err)
	}
	orders := make([]domain.Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, toOrder(row))
	}
	return orders, nil
}

// OrderByID returns an order or domain.ErrOrderNotFound.
func (s *Store) OrderByID(ctx context.Context, id uuid.UUID) (domain.Order, error) {
	row, err := s.q.GetOrder(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("repository: get order: %w", err)
	}
	return toOrder(row), nil
}

func (s *Store) productWithItems(ctx context.Context, row storerepo.StoreProduct) (domain.Product, error) {
	items, err := s.q.ListProductItems(ctx, row.ID)
	if err != nil {
		return domain.Product{}, fmt.Errorf("repository: list product items: %w", err)
	}
	product := domain.Product{
		ID:          row.ID,
		SKU:         row.Sku,
		Name:        row.Name,
		Description: row.Description,
		PricePoints: row.PricePoints,
		Money:       row.Money,
		Active:      row.Active,
		Items:       make([]domain.ProductItem, 0, len(items)),
	}
	for _, item := range items {
		product.Items = append(product.Items, domain.ProductItem{ItemID: int(item.ItemID), Count: int(item.Count)})
	}
	return product, nil
}

func toOrder(row storerepo.StoreOrder) domain.Order {
	return domain.Order{
		ID:            row.ID,
		UserID:        row.UserID,
		ProductID:     row.ProductID,
		SKU:           row.Sku,
		PricePoints:   row.PricePoints,
		CharacterName: row.CharacterName,
		AccountID:     row.AccountID,
		Status:        domain.OrderStatus(row.Status),
		CommandOutput: row.CommandOutput,
		CreatedAt:     row.CreatedAt,
	}
}

func nullUUID(value uuid.UUID) uuid.NullUUID {
	if value == uuid.Nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: value, Valid: true}
}
