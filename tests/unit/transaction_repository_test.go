package tests

import (
	"testing"
	"time"
	"xyz-multifinance/internal/domain"
	"xyz-multifinance/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTransactionRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}

	repo := repository.NewTransactionRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		tx := &domain.Transaction{
			CustomerID:        1,
			ContractNumber:    "XYZ-1-123",
			Source:            domain.SourceECommerce,
			Status:            domain.StatusPending,
			AssetName:         "Laptop",
			OTRAmount:         10000000,
			AdminFee:          100000,
			InstallmentAmount: 916667,
			InterestAmount:    1000000,
			Tenor:             12,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "transactions" \("customer_id","contract_number","source","status","asset_name","otr_amount","admin_fee","installment_amount","interest_amount","tenor","version","created_at","updated_at","deleted_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10,\$11,\$12,\$13,\$14\) RETURNING "id"`).
			WithArgs(
				tx.CustomerID,
				tx.ContractNumber,
				tx.Source,
				tx.Status,
				tx.AssetName,
				tx.OTRAmount,
				tx.AdminFee,
				tx.InstallmentAmount,
				tx.InterestAmount,
				tx.Tenor,
				1,                // version
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				nil,              // deleted_at
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Expect installment creation
		for i := 1; i <= tx.Tenor; i++ {
			mock.ExpectQuery(`INSERT INTO "installments" \("transaction_id","installment_number","amount","status","due_date","version","created_at","updated_at","deleted_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9\) RETURNING "id"`).
				WithArgs(
					1, // transaction_id
					i, // installment_number
					tx.InstallmentAmount,
					"unpaid",
					sqlmock.AnyArg(), // due_date
					1,                // version
					sqlmock.AnyArg(), // created_at
					sqlmock.AnyArg(), // updated_at
					nil,              // deleted_at
				).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i))
		}

		mock.ExpectCommit()

		err := repo.Create(tx)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error", func(t *testing.T) {
		tx := &domain.Transaction{
			CustomerID:        1,
			ContractNumber:    "XYZ-1-123",
			Source:            domain.SourceECommerce,
			Status:            domain.StatusPending,
			AssetName:         "Laptop",
			OTRAmount:         10000000,
			AdminFee:          100000,
			InstallmentAmount: 916667,
			InterestAmount:    1000000,
			Tenor:             12,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "transactions" \("customer_id","contract_number","source","status","asset_name","otr_amount","admin_fee","installment_amount","interest_amount","tenor","version","created_at","updated_at","deleted_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10,\$11,\$12,\$13,\$14\) RETURNING "id"`).
			WithArgs(
				tx.CustomerID,
				tx.ContractNumber,
				tx.Source,
				tx.Status,
				tx.AssetName,
				tx.OTRAmount,
				tx.AdminFee,
				tx.InstallmentAmount,
				tx.InterestAmount,
				tx.Tenor,
				1,                // version
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				nil,              // deleted_at
			).
			WillReturnError(gorm.ErrInvalidTransaction)
		mock.ExpectRollback()

		err := repo.Create(tx)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTransactionRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}

	repo := repository.NewTransactionRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		id := uint(1)

		// Main transaction query
		transactionRows := sqlmock.NewRows([]string{
			"id", "customer_id", "contract_number", "source", "status",
			"asset_name", "otr_amount", "admin_fee", "installment_amount",
			"interest_amount", "tenor", "created_at", "updated_at",
			"deleted_at", "version",
		}).AddRow(
			id, 1, "XYZ-1-123", domain.SourceECommerce, domain.StatusPending,
			"Laptop", 10000000, 100000, 916667,
			1000000, 12, time.Now(), time.Now(),
			nil, 1,
		)

		// Customer query
		customerRows := sqlmock.NewRows([]string{
			"id", "nik", "full_name", "legal_name", "place_of_birth",
			"date_of_birth", "salary", "ktp_photo", "selfie_photo",
			"created_at", "updated_at", "deleted_at", "version",
		}).AddRow(
			1, "1234567890123456", "John Doe", "John Doe", "Jakarta",
			time.Now().AddDate(-30, 0, 0), 5000000, "ktp.jpg", "selfie.jpg",
			time.Now(), time.Now(), nil, 1,
		)

		// Installments query
		installmentRows := sqlmock.NewRows([]string{
			"id", "transaction_id", "installment_number", "amount",
			"status", "due_date", "paid_at", "created_at", "updated_at",
			"version",
		})
		for i := 1; i <= 12; i++ {
			installmentRows.AddRow(
				i, id, i, 916667,
				"unpaid", time.Now().AddDate(0, i, 0), nil,
				time.Now(), time.Now(), 1,
			)
		}

		mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE "transactions"."id" = \$1 ORDER BY "transactions"."id" LIMIT \$2`).
			WithArgs(id, 1).
			WillReturnRows(transactionRows)

		mock.ExpectQuery(`SELECT \* FROM "customers" WHERE "customers"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(customerRows)

		mock.ExpectQuery(`SELECT \* FROM "installments" WHERE "installments"."transaction_id" = \$1`).
			WithArgs(id).
			WillReturnRows(installmentRows)

		tx, err := repo.GetByID(id)

		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, id, tx.ID)
		assert.NotNil(t, tx.Customer)
		assert.Equal(t, 12, len(tx.Installments))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		id := uint(1)

		mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE "transactions"."id" = \$1 ORDER BY "transactions"."id" LIMIT \$2`).
			WithArgs(id, 1).
			WillReturnRows(sqlmock.NewRows([]string{}))

		tx, err := repo.GetByID(id)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTransactionRepository_GetByContractNumber(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}

	repo := repository.NewTransactionRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		contractNumber := "XYZ-1-123"

		// Transaction query
		transactionRows := sqlmock.NewRows([]string{
			"id", "customer_id", "contract_number", "source", "status",
			"asset_name", "otr_amount", "admin_fee", "installment_amount",
			"interest_amount", "tenor", "created_at", "updated_at",
			"deleted_at", "version",
		}).AddRow(
			1, 1, contractNumber, domain.SourceECommerce, domain.StatusPending,
			"Laptop", 10000000, 100000, 916667,
			1000000, 12, time.Now(), time.Now(),
			nil, 1,
		)

		// Customer query
		customerRows := sqlmock.NewRows([]string{
			"id", "nik", "full_name", "legal_name", "place_of_birth",
			"date_of_birth", "salary", "ktp_photo", "selfie_photo",
			"created_at", "updated_at", "deleted_at", "version",
		}).AddRow(
			1, "1234567890123456", "John Doe", "John Doe", "Jakarta",
			time.Now().AddDate(-30, 0, 0), 5000000, "ktp.jpg", "selfie.jpg",
			time.Now(), time.Now(), nil, 1,
		)

		// Installments query
		installmentRows := sqlmock.NewRows([]string{
			"id", "transaction_id", "installment_number", "amount",
			"status", "due_date", "paid_at", "created_at", "updated_at",
			"version",
		})
		for i := 1; i <= 12; i++ {
			installmentRows.AddRow(
				i, 1, i, 916667,
				"unpaid", time.Now().AddDate(0, i, 0), nil,
				time.Now(), time.Now(), 1,
			)
		}

		mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE "contract_number" = \$1 AND "deleted_at" IS NULL`).
			WithArgs(contractNumber).
			WillReturnRows(transactionRows)

		mock.ExpectQuery(`SELECT \* FROM "customers" WHERE "id" = \$1 AND "deleted_at" IS NULL`).
			WithArgs(1).
			WillReturnRows(customerRows)

		mock.ExpectQuery(`SELECT \* FROM "installments" WHERE "transaction_id" = \$1 AND "deleted_at" IS NULL ORDER BY "installment_number" ASC`).
			WithArgs(1).
			WillReturnRows(installmentRows)

		tx, err := repo.GetByContractNumber(contractNumber)

		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, contractNumber, tx.ContractNumber)
		assert.NotNil(t, tx.Customer)
		assert.Equal(t, 12, len(tx.Installments))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		contractNumber := "XYZ-1-123"

		mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE "contract_number" = \$1 AND "deleted_at" IS NULL`).
			WithArgs(contractNumber).
			WillReturnRows(sqlmock.NewRows([]string{}))

		tx, err := repo.GetByContractNumber(contractNumber)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTransactionRepository_GetInstallments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}

	repo := repository.NewTransactionRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		transactionID := uint(1)
		rows := sqlmock.NewRows([]string{
			"id", "transaction_id", "installment_number", "amount",
			"status", "due_date", "created_at", "updated_at",
			"deleted_at", "version",
		}).AddRow(
			1, transactionID, 1, 916667,
			"unpaid", time.Now().AddDate(0, 1, 0), time.Now(), time.Now(),
			nil, 1,
		).AddRow(
			2, transactionID, 2, 916667,
			"unpaid", time.Now().AddDate(0, 2, 0), time.Now(), time.Now(),
			nil, 1,
		)

		mock.ExpectQuery("^SELECT (.+) FROM \"installments\"").
			WithArgs(transactionID).
			WillReturnRows(rows)

		installments, err := repo.GetInstallments(transactionID)

		assert.NoError(t, err)
		assert.Len(t, installments, 2)
		assert.Equal(t, transactionID, installments[0].TransactionID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		transactionID := uint(1)

		mock.ExpectQuery("^SELECT (.+) FROM \"installments\"").
			WithArgs(transactionID).
			WillReturnRows(sqlmock.NewRows([]string{}))

		installments, err := repo.GetInstallments(transactionID)

		assert.NoError(t, err)
		assert.Empty(t, installments)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTransactionRepository_UpdateInstallment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}

	repo := repository.NewTransactionRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		installment := &domain.Installment{
			ID:                1,
			TransactionID:     1,
			InstallmentNumber: 1,
			Amount:            916667,
			Status:            "paid",
			DueDate:           time.Now().AddDate(0, 1, 0),
			Version:           1,
		}

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE \"installments\"").
			WithArgs(
				installment.TransactionID,
				installment.InstallmentNumber,
				installment.Amount,
				installment.Status,
				installment.DueDate,
				sqlmock.AnyArg(), // updated_at
				installment.Version+1,
				installment.ID,
				installment.Version,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateInstallment(installment)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Optimistic Lock Error", func(t *testing.T) {
		installment := &domain.Installment{
			ID:                1,
			TransactionID:     1,
			InstallmentNumber: 1,
			Amount:            916667,
			Status:            "paid",
			DueDate:           time.Now().AddDate(0, 1, 0),
			Version:           1,
		}

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE \"installments\"").
			WithArgs(
				installment.TransactionID,
				installment.InstallmentNumber,
				installment.Amount,
				installment.Status,
				installment.DueDate,
				sqlmock.AnyArg(), // updated_at
				installment.Version+1,
				installment.ID,
				installment.Version,
			).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectRollback()

		err := repo.UpdateInstallment(installment)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "optimistic lock error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
