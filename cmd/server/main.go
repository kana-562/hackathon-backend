package main

import (
	"log"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"hobby-relay-backend/internal/ai"
	"hobby-relay-backend/internal/config"
	"hobby-relay-backend/internal/handler"
	mw "hobby-relay-backend/internal/middleware"
	"hobby-relay-backend/internal/repository/mysql"
	"hobby-relay-backend/internal/usecase"
)

func main() {
	cfg := config.Load()

	db, err := mysql.NewDB(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect DB: %v", err)
	}
	defer db.Close()

	// AI client
	var aiClient ai.Client
	if cfg.AIAPIKey != "" {
		log.Printf("Using Gemini client (model: %s)", cfg.AIModel)
		aiClient = ai.NewGeminiClient(cfg.AIAPIKey, cfg.AIModel)
	} else {
		aiClient = ai.NewMockClient()
	}

	// Repositories
	userRepo := mysql.NewUserRepository(db)
	categoryRepo := mysql.NewCategoryRepository(db)
	setRepo := mysql.NewSetRepository(db)
	favoriteRepo := mysql.NewFavoriteRepository(db)
	txRepo := mysql.NewTransactionRepository(db)
	startPlanRepo := mysql.NewStartPlanRepository(db)
	sessionRepo := mysql.NewAIChatSessionRepository(db)
	messageRepo := mysql.NewAIMessageRepository(db)
	dmRepo := mysql.NewDMRepository(db)

	// Usecases
	authUC := usecase.NewAuthUsecase(userRepo, cfg.JWTSecret)
	homeUC := usecase.NewHomeUsecase(categoryRepo, setRepo, favoriteRepo)
	setsUC := usecase.NewSetsUsecase(setRepo, favoriteRepo, aiClient)
	sellUC := usecase.NewSellUsecase(setRepo, sessionRepo, messageRepo, aiClient)
	txUC := usecase.NewTransactionUsecase(txRepo, setRepo, startPlanRepo, aiClient)
	mypageUC := usecase.NewMypageUsecase(userRepo, setRepo, txRepo, favoriteRepo)
	dmUC := usecase.NewDMUsecase(dmRepo, userRepo, setRepo)

	// Handlers
	authH := handler.NewAuthHandler(authUC)
	homeH := handler.NewHomeHandler(homeUC)
	setsH := handler.NewSetsHandler(setsUC)
	sellH := handler.NewSellHandler(sellUC)
	txH := handler.NewTransactionHandler(txUC)
	mypageH := handler.NewMypageHandler(mypageUC)
	dmH := handler.NewDMHandler(dmUC)

	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(mw.CORS(cfg.FrontendOrigin))

	jwtRequired := mw.JWTAuth(cfg.JWTSecret)
	jwtOptional := mw.OptionalJWTAuth(cfg.JWTSecret)

	api := e.Group("/api")

	// Auth
	api.POST("/auth/signup", authH.Signup)
	api.POST("/auth/login", authH.Login)
	api.GET("/me", mypageH.GetMe, jwtRequired)

	// Home & Category
	api.GET("/home", homeH.GetHome, jwtOptional)
	api.GET("/categories", homeH.GetCategories)
	api.GET("/categories/:id/sets", setsH.GetCategorySets, jwtOptional)

	// Sets
	api.GET("/sets", setsH.GetSets, jwtOptional)
	api.GET("/sets/:id", setsH.GetSetDetail, jwtOptional)
	api.POST("/sets/:id/favorite", setsH.AddFavorite, jwtRequired)
	api.DELETE("/sets/:id/favorite", setsH.RemoveFavorite, jwtRequired)
	api.POST("/sets/:id/questions", setsH.AskQuestion, jwtOptional)

	// Sell
	api.POST("/sell/drafts", sellH.CreateDraft, jwtRequired)
	api.POST("/sell/drafts/:id/messages", sellH.SendMessage, jwtRequired)
	api.GET("/sell/drafts/:id", sellH.GetDraft, jwtRequired)
	api.PUT("/sell/drafts/:id", sellH.UpdateDraft, jwtRequired)
	api.POST("/sell/drafts/:id/publish", sellH.PublishDraft, jwtRequired)

	// Transactions
	api.POST("/transactions", txH.CreateTransaction, jwtRequired)
	api.GET("/transactions/:id", txH.GetTransaction, jwtRequired)
	api.PATCH("/transactions/:id", txH.UpdateTransaction, jwtRequired)
	api.POST("/transactions/:id/start-plan", txH.GetStartPlan, jwtRequired)

	// MyPage
	api.GET("/mypage", mypageH.GetMyPage, jwtRequired)
	api.GET("/mypage/selling", mypageH.GetSelling, jwtRequired)
	api.GET("/mypage/purchases", mypageH.GetPurchases, jwtRequired)
	api.GET("/mypage/favorites", mypageH.GetFavorites, jwtRequired)

	// DM
	api.POST("/dm/rooms", dmH.GetOrCreateRoom, jwtRequired)
	api.GET("/dm/rooms", dmH.ListRooms, jwtRequired)
	api.GET("/dm/rooms/:id/messages", dmH.GetMessages, jwtRequired)
	api.POST("/dm/rooms/:id/messages", dmH.SendMessage, jwtRequired)
	api.PATCH("/dm/rooms/:id/read", dmH.MarkRead, jwtRequired)

	log.Printf("Starting server on port %s", cfg.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
