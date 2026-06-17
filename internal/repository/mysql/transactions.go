package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) repository.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(tx *domain.Transaction) (int64, error) {
	dbTx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = dbTx.Rollback()
		}
	}()

	// Lock the set row
	var status string
	var sellerID int64
	var price int
	err = dbTx.QueryRow(`SELECT status, seller_id, price FROM starter_sets WHERE id = ? FOR UPDATE`, tx.StarterSetID).
		Scan(&status, &sellerID, &price)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("set not found")
	}
	if err != nil {
		return 0, err
	}
	if status != domain.SetStatusOnSale {
		return 0, fmt.Errorf("set is not available for purchase (status: %s)", status)
	}
	if sellerID == tx.BuyerID {
		return 0, fmt.Errorf("cannot purchase your own set")
	}

	// Update set status
	_, err = dbTx.Exec(`UPDATE starter_sets SET status=?, updated_at=? WHERE id=?`,
		domain.SetStatusReserved, time.Now(), tx.StarterSetID)
	if err != nil {
		return 0, err
	}

	// Create transaction
	now := time.Now()
	result, err := dbTx.Exec(`
		INSERT INTO transactions (starter_set_id, buyer_id, seller_id, price, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		tx.StarterSetID, tx.BuyerID, sellerID, price, domain.TxStatusReserved, now, now,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = dbTx.Commit(); err != nil {
		return 0, err
	}

	tx.SellerID = sellerID
	tx.Price = price
	tx.Status = domain.TxStatusReserved
	return id, nil
}

func (r *transactionRepository) FindByID(id int64) (*domain.Transaction, error) {
	tx := &domain.Transaction{}
	err := r.db.QueryRow(`
		SELECT id, starter_set_id, buyer_id, seller_id, price, status, created_at, updated_at
		FROM transactions WHERE id = ?`, id,
	).Scan(&tx.ID, &tx.StarterSetID, &tx.BuyerID, &tx.SellerID, &tx.Price, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *transactionRepository) FindByBuyer(buyerID int64) ([]domain.Transaction, error) {
	rows, err := r.db.Query(`
		SELECT id, starter_set_id, buyer_id, seller_id, price, status, created_at, updated_at
		FROM transactions WHERE buyer_id = ? ORDER BY created_at DESC`, buyerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		if err := rows.Scan(&tx.ID, &tx.StarterSetID, &tx.BuyerID, &tx.SellerID, &tx.Price, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}

func (r *transactionRepository) FindBySeller(sellerID int64) ([]domain.Transaction, error) {
	rows, err := r.db.Query(`
		SELECT id, starter_set_id, buyer_id, seller_id, price, status, created_at, updated_at
		FROM transactions WHERE seller_id = ? ORDER BY created_at DESC`, sellerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		if err := rows.Scan(&tx.ID, &tx.StarterSetID, &tx.BuyerID, &tx.SellerID, &tx.Price, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}

func (r *transactionRepository) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE transactions SET status=?, updated_at=? WHERE id=?`, status, time.Now(), id)
	return err
}
