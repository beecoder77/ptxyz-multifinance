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

func TestCustomerRepository_Create(t *testing.T) {
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

	repo := repository.NewCustomerRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		customer := &domain.Customer{
			NIK:          "1234567890123456",
			FullName:     "John Doe",
			LegalName:    "John Doe",
			PlaceOfBirth: "Jakarta",
			DateOfBirth:  time.Now().AddDate(-30, 0, 0),
			Salary:       5000000,
			KTPPhoto:     "ktp.jpg",
			SelfiePhoto:  "selfie.jpg",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "customers"`).
			WithArgs(
				customer.NIK,
				customer.FullName,
				customer.LegalName,
				customer.PlaceOfBirth,
				customer.DateOfBirth,
				customer.Salary,
				customer.KTPPhoto,
				customer.SelfiePhoto,
				sqlmock.AnyArg(), // version
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				sqlmock.AnyArg(), // deleted_at
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Create(customer)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error", func(t *testing.T) {
		customer := &domain.Customer{
			NIK:          "1234567890123456",
			FullName:     "John Doe",
			LegalName:    "John Doe",
			PlaceOfBirth: "Jakarta",
			DateOfBirth:  time.Now().AddDate(-30, 0, 0),
			Salary:       5000000,
			KTPPhoto:     "ktp.jpg",
			SelfiePhoto:  "selfie.jpg",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "customers"`).
			WithArgs(
				customer.NIK,
				customer.FullName,
				customer.LegalName,
				customer.PlaceOfBirth,
				customer.DateOfBirth,
				customer.Salary,
				customer.KTPPhoto,
				customer.SelfiePhoto,
				sqlmock.AnyArg(), // version
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				sqlmock.AnyArg(), // deleted_at
			).
			WillReturnError(gorm.ErrInvalidTransaction)
		mock.ExpectRollback()

		err := repo.Create(customer)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCustomerRepository_GetByID(t *testing.T) {
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

	repo := repository.NewCustomerRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		id := uint(1)
		dob := time.Now().AddDate(-30, 0, 0)

		// Main customer query
		customerRows := sqlmock.NewRows([]string{
			"id", "nik", "full_name", "legal_name", "place_of_birth",
			"date_of_birth", "salary", "ktp_photo", "selfie_photo",
			"created_at", "updated_at", "deleted_at", "version",
		}).AddRow(
			id, "1234567890123456", "John Doe", "John Doe", "Jakarta",
			dob, 5000000, "ktp.jpg", "selfie.jpg",
			time.Now(), time.Now(), nil, 1,
		)

		// Credit limits query
		creditLimitRows := sqlmock.NewRows([]string{
			"id", "customer_id", "tenor", "limit_amount",
			"created_at", "updated_at", "deleted_at", "version",
		})

		mock.ExpectQuery(`SELECT \* FROM "customers" WHERE "customers"."id" = \$1 ORDER BY "customers"."id" LIMIT \$2`).
			WithArgs(id, 1).
			WillReturnRows(customerRows)

		mock.ExpectQuery(`SELECT \* FROM "credit_limits" WHERE "credit_limits"."customer_id" = \$1`).
			WithArgs(id).
			WillReturnRows(creditLimitRows)

		customer, err := repo.GetByID(id)

		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, id, customer.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		id := uint(1)

		mock.ExpectQuery(`SELECT \* FROM "customers" WHERE "customers"."id" = \$1 ORDER BY "customers"."id" LIMIT \$2`).
			WithArgs(id, 1).
			WillReturnRows(sqlmock.NewRows([]string{}))

		customer, err := repo.GetByID(id)

		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCustomerRepository_GetByNIK(t *testing.T) {
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

	repo := repository.NewCustomerRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		nik := "1234567890123456"
		dob := time.Now().AddDate(-30, 0, 0)

		// Main customer query
		customerRows := sqlmock.NewRows([]string{
			"id", "nik", "full_name", "legal_name", "place_of_birth",
			"date_of_birth", "salary", "ktp_photo", "selfie_photo",
			"created_at", "updated_at", "deleted_at", "version",
		}).AddRow(
			1, nik, "John Doe", "John Doe", "Jakarta",
			dob, 5000000, "ktp.jpg", "selfie.jpg",
			time.Now(), time.Now(), nil, 1,
		)

		// Credit limits query
		creditLimitRows := sqlmock.NewRows([]string{
			"id", "customer_id", "tenor", "limit_amount",
			"created_at", "updated_at", "deleted_at", "version",
		})

		mock.ExpectQuery(`SELECT \* FROM "customers" WHERE nik = \$1 ORDER BY "customers"."id" LIMIT \$2`).
			WithArgs(nik, 1).
			WillReturnRows(customerRows)

		mock.ExpectQuery(`SELECT \* FROM "credit_limits" WHERE "credit_limits"."customer_id" = \$1`).
			WithArgs(1).
			WillReturnRows(creditLimitRows)

		customer, err := repo.GetByNIK(nik)

		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, nik, customer.NIK)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		nik := "1234567890123456"

		mock.ExpectQuery(`SELECT \* FROM "customers" WHERE nik = \$1 ORDER BY "customers"."id" LIMIT \$2`).
			WithArgs(nik, 1).
			WillReturnRows(sqlmock.NewRows([]string{}))

		customer, err := repo.GetByNIK(nik)

		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCustomerRepository_Update(t *testing.T) {
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

	repo := repository.NewCustomerRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		customer := &domain.Customer{
			ID:           1,
			NIK:          "1234567890123456",
			FullName:     "John Doe Updated",
			LegalName:    "John Doe",
			PlaceOfBirth: "Jakarta",
			DateOfBirth:  time.Now().AddDate(-30, 0, 0),
			Salary:       6000000,
			KTPPhoto:     "ktp.jpg",
			SelfiePhoto:  "selfie.jpg",
			Version:      1,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT version FROM "customers" WHERE "customers"."id" = \$1 ORDER BY "customers"."id" LIMIT \$2`).
			WithArgs(customer.ID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(1))

		mock.ExpectExec(`UPDATE "customers" SET "nik"=\$1,"full_name"=\$2,"legal_name"=\$3,"place_of_birth"=\$4,"date_of_birth"=\$5,"salary"=\$6,"ktp_photo"=\$7,"selfie_photo"=\$8,"version"=\$9,"created_at"=\$10,"updated_at"=\$11,"deleted_at"=\$12 WHERE "id" = \$13`).
			WithArgs(
				customer.NIK,
				customer.FullName,
				customer.LegalName,
				customer.PlaceOfBirth,
				customer.DateOfBirth,
				customer.Salary,
				customer.KTPPhoto,
				customer.SelfiePhoto,
				customer.Version+1,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				nil,              // deleted_at
				customer.ID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(customer)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Optimistic_Lock_Error", func(t *testing.T) {
		customer := &domain.Customer{
			ID:           1,
			NIK:          "1234567890123456",
			FullName:     "John Doe Updated",
			LegalName:    "John Doe",
			PlaceOfBirth: "Jakarta",
			DateOfBirth:  time.Now().AddDate(-30, 0, 0),
			Salary:       6000000,
			KTPPhoto:     "ktp.jpg",
			SelfiePhoto:  "selfie.jpg",
			Version:      1,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT version FROM "customers" WHERE "customers"."id" = \$1 ORDER BY "customers"."id" LIMIT \$2`).
			WithArgs(customer.ID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(2))
		mock.ExpectRollback()

		err := repo.Update(customer)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "concurrent modification detected")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCustomerRepository_GetCreditLimits(t *testing.T) {
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

	repo := repository.NewCustomerRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		customerID := uint(1)
		rows := sqlmock.NewRows([]string{
			"id", "customer_id", "tenor", "amount", "used_amount",
			"created_at", "updated_at", "deleted_at", "version",
		}).AddRow(
			1, customerID, 12, 10000000, 0,
			time.Now(), time.Now(), nil, 1,
		).AddRow(
			2, customerID, 24, 20000000, 5000000,
			time.Now(), time.Now(), nil, 1,
		)

		mock.ExpectQuery("^SELECT (.+) FROM \"credit_limits\"").
			WithArgs(customerID).
			WillReturnRows(rows)

		limits, err := repo.GetCreditLimits(customerID)

		assert.NoError(t, err)
		assert.Len(t, limits, 2)
		assert.Equal(t, customerID, limits[0].CustomerID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		customerID := uint(1)

		mock.ExpectQuery("^SELECT (.+) FROM \"credit_limits\"").
			WithArgs(customerID).
			WillReturnRows(sqlmock.NewRows([]string{}))

		limits, err := repo.GetCreditLimits(customerID)

		assert.NoError(t, err)
		assert.Empty(t, limits)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCustomerRepository_UpdateCreditLimit(t *testing.T) {
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

	repo := repository.NewCustomerRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		limit := &domain.CreditLimit{
			ID:         1,
			CustomerID: 1,
			Tenor:      12,
			Amount:     10000000,
			UsedAmount: 5000000,
			Version:    1,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT version FROM "credit_limits" WHERE "credit_limits"."id" = \$1 ORDER BY "credit_limits"."id" LIMIT \$2`).
			WithArgs(limit.ID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(1))

		mock.ExpectExec(`UPDATE "credit_limits" SET "customer_id"=\$1,"tenor"=\$2,"amount"=\$3,"used_amount"=\$4,"version"=\$5,"created_at"=\$6,"updated_at"=\$7,"deleted_at"=\$8 WHERE "id" = \$9`).
			WithArgs(
				limit.CustomerID,
				limit.Tenor,
				limit.Amount,
				limit.UsedAmount,
				limit.Version+1,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				nil,              // deleted_at
				limit.ID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateCreditLimit(limit)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Optimistic_Lock_Error", func(t *testing.T) {
		limit := &domain.CreditLimit{
			ID:         1,
			CustomerID: 1,
			Tenor:      12,
			Amount:     10000000,
			UsedAmount: 5000000,
			Version:    1,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT version FROM "credit_limits" WHERE "credit_limits"."id" = \$1 ORDER BY "credit_limits"."id" LIMIT \$2`).
			WithArgs(limit.ID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(2))
		mock.ExpectRollback()

		err := repo.UpdateCreditLimit(limit)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "concurrent modification detected")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
