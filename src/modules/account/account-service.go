package accountModule

import (
	errorHelpers "go-gin-test-job/src/common/error-helpers"
	"go-gin-test-job/src/database"
	"go-gin-test-job/src/database/entities"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getAccounts(
	c *gin.Context,
	status entities.AccountStatus,
	offset, count int,
	search string,
	orderParams map[string]string,
) ([]*entities.Account, int64) {
	params := database.AccountQueryParams{
		Status:  status,
		Offset:  offset,
		Count:   count,
		Search:  search,
		OrderBy: orderParams,
	}
	return database.GetAccountsAndTotal(c, params)
}

func createAccount(
	c *gin.Context,
	address string,
	status entities.AccountStatus,
	name string,
	rank uint8,
	memo *string,
) (*entities.Account, error) {
	var account *entities.Account
	transactionError := database.DbConn.Transaction(func(tx *gorm.DB) error {
		if database.IsAddressExists(tx, address) {
			return errorHelpers.RespondConflictError(c, "Address already exists")
		}
		newAccount, err := database.CreateAccount(
			tx,
			entities.CreateAccount(
				address,
				status,
				name,
				rank,
				memo,
			),
		)
		if err != nil {
			return err
		}
		account = newAccount
		return nil
	}, database.DefaultTxOptions)
	if transactionError != nil {
		return nil, transactionError
	}
	return account, nil
}
