package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"my-chi-app/internal/database"
	"my-chi-app/internal/database/repository"
	httpdelivery "my-chi-app/internal/delivery/http"
	"my-chi-app/internal/storage"

	_ "my-chi-app/docs"

	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Swagger Info
// @title WebForum API
// @version 1.0
// @description A forum application with posts, comments, and reactions
// @host localhost:3000
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description JWT token
// Main function to start the server
func main() {
	ctx := context.Background()

	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	dsn := os.Getenv("DB_DSN")

	if dsn == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	db, err := database.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	s3Bucket := os.Getenv("AWS_S3_BUCKET")
	s3Region := os.Getenv("AWS_REGION")
	s3Client, err := storage.NewS3Client(ctx, s3Bucket, s3Region)

	if err != nil {
		log.Fatalf("failed to create S3 client: %v", err)
	}

	deps := httpdelivery.RouterDeps{
		UserRepo:            repository.NewUserRepository(db),
		TokenRepo:           repository.NewTokenRepository(db),
		CategoryRepo:        repository.NewCategoryRepository(db),
		MembershipRepo:      repository.NewMembershipRepository(db),
		PostRepo:            repository.NewPostRepository(db),
		ReactionRepo:        repository.NewReactionRepository(db),
		ReactionTypeRepo:    repository.NewReactionTypeRepository(db),
		CommentRepo:         repository.NewCommentRepository(db),
		CommentReactionRepo: repository.NewCommentReactionRepository(db),
		NotificationRepo:    repository.NewNotificationRepository(db),
		S3Client:            s3Client,
		JWTSecret:           jwtSecret,
	}

	r := httpdelivery.Routes(deps)

	// Swagger Ui endpoint for API documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":" + port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
