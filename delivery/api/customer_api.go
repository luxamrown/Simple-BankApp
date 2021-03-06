package api

import (
	"net/http"
	"strconv"

	"enigmacamp.com/bank/delivery/appreq"
	"enigmacamp.com/bank/delivery/tokenauth"
	"enigmacamp.com/bank/usecase"
	"github.com/gin-gonic/gin"
)

type CustomerApi struct {
	BaseApi
	loginUseCase    usecase.LoginUseCase
	transferUseCase usecase.TransferUseCase
	logoutUseCase   usecase.LogoutUseCase
}

func (a *CustomerApi) userLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user tokenauth.Credential
		err := a.ParseRequestBody(c, &user)
		if err != nil {
			c.Error(err)
			return
		}
		token, err := tokenauth.GenerateToken(user.AccountNumber, "user@corp.com")
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		err = a.loginUseCase.Login(user.AccountNumber, user.Password, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(200, gin.H{
			"token": token,
		})
	}
}

func (a *CustomerApi) userLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		var authHeader appreq.AuthHeader
		accountNumber, _ := strconv.Atoi(c.Param("user"))
		err := c.ShouldBindHeader(&authHeader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = a.logoutUseCase.Logout(accountNumber, authHeader.AuthorizationHeader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"LOGOUT": accountNumber})
	}
}

func (a *CustomerApi) Transfer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var authHeader appreq.AuthHeader
		var custReq appreq.CustomerRequestPayment
		senderAccountNumber, _ := strconv.Atoi(c.Param("user"))
		err := a.ParseRequestBody(c, &custReq)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		err = c.ShouldBindHeader(&authHeader)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		err = a.transferUseCase.Transfer(senderAccountNumber, custReq.ReceiverAccountNumber, authHeader.AuthorizationHeader, custReq.AmountTransfer, custReq.IsMerchant)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":       "TRANSFER SUCCES",
			"Transfer From": senderAccountNumber,
			"Transfer to":   custReq.ReceiverAccountNumber,
			"Amount":        custReq.AmountTransfer,
		})
	}
}

func NewCustomerApi(customerRoute *gin.RouterGroup, loginUseCase usecase.LoginUseCase, transferUseCase usecase.TransferUseCase, logoutUseCase usecase.LogoutUseCase) {
	api := CustomerApi{
		loginUseCase:    loginUseCase,
		transferUseCase: transferUseCase,
		logoutUseCase:   logoutUseCase,
	}
	customerRoute.POST("/login", api.userLogin())
	customerRoute.POST("/:user/logout", api.userLogout())
	customerRoute.POST("/:user/transfer", api.Transfer())
}
