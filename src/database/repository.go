package database

import (
	"context"
	"fmt"
	"go-gin-test-job/src/database/entities"

	"gorm.io/gorm"
)

func accountTableName() string {
	return entities.Account{}.TableName()
}

func getDb(tx *gorm.DB) *gorm.DB {
	var db *gorm.DB
	if tx != nil {
		db = tx
	} else {
		db = DbConn
	}
	return db
}

///// Account queries

type AccountQueryParams struct {
	Status  entities.AccountStatus
	Search  string
	OrderBy map[string]string
	Offset  int
	Count   int
}

func GetAccountsAndTotal(ctx context.Context, params AccountQueryParams) ([]*entities.Account, int64) {
	var total int64
	totalQuery := getBaseAccountsQuery(ctx, params.Status)
	totalQuery.Count(&total)

	query := getBaseAccountsQuery(ctx, params.Status)
	for column, direction := range params.OrderBy {
		query = query.Order(fmt.Sprintf("account.%s %s", column, direction))
	}

	query = getSearchAccountQueryCondition(query, params.Search)

	var accounts []*entities.Account
	query.
		Limit(params.Count).
		Offset(params.Offset).
		Find(&accounts)
	return accounts, total
}

func getSearchAccountQueryCondition(query *gorm.DB, search string) *gorm.DB {
	if search == "" {
		return query
	}

	target := "%" + search + "%"
	return query.Where(
		query.Where("account.address LIKE ?", target).
			Or("account.name LIKE ?", target).
			Or("account.memo IS NOT NULL AND account.memo LIKE ?", target),
	)
}

func getBaseAccountsQuery(ctx context.Context, status entities.AccountStatus) *gorm.DB {
	query := DbConn.WithContext(ctx).Table(accountTableName() + " account")
	if status != "" {
		query = query.Where("account.status = ?", status)
	}
	return query
}

func IsAddressExists(tx *gorm.DB, address string) bool {
	db := getDb(tx)
	var account *entities.Account
	db.Table(accountTableName()+" account").
		Where("account.address = ?", address).
		First(&account)
	if account.Id != 0 {
		return true
	}
	return false
}

func GetAccountByAddress(address string) *entities.Account {
	var account *entities.Account
	DbConn.Table(accountTableName()+" account").
		Where("account.address = ?", address).
		First(&account)
	if account.Id == 0 {
		return nil
	}
	return account
}

func CreateAccount(tx *gorm.DB, newAccount *entities.Account) (*entities.Account, error) {
	err := tx.Create(newAccount).Error
	if err != nil {
		return nil, err
	}
	return newAccount, nil
}

func GetAccountsBatch(limit int) []*entities.Account {
	var accounts []*entities.Account
	DbConn.Table(accountTableName()+" account").
		Where("account.status = ?", entities.AccountStatusOn).
		Order("account.updated_at ASC").
		Limit(limit).
		Find(&accounts)
	return accounts
}

func GetAccountsByIds(accountIds []int64) []*entities.Account {
	var accounts []*entities.Account
	DbConn.Table(accountTableName()+" account").
		Where("account.id IN(?)", accountIds).
		Find(&accounts)
	return accounts
}

func UpdateAccount(tx *gorm.DB, account *entities.Account, updateData map[string]interface{}) error {
	db := getDb(tx)
	return db.Model(entities.Account{}).Where("id = ?", account.Id).Updates(updateData).Error
}
