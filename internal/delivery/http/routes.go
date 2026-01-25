package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"my-chi-app/internal/database/repository"
	"my-chi-app/internal/storage"
)

// RouterDeps holds all dependencies required to set up the router
type RouterDeps struct {
	UserRepo            *repository.UserRepository
	TokenRepo           *repository.TokenRepository
	CategoryRepo        *repository.CategoryRepository
	MembershipRepo      *repository.MembershipRepository
	PostRepo            *repository.PostRepository
	ReactionRepo        *repository.ReactionRepository
	ReactionTypeRepo    *repository.ReactionTypeRepository
	CommentRepo         *repository.CommentRepository
	CommentReactionRepo *repository.CommentReactionRepository
	NotificationRepo    *repository.NotificationRepository
	S3Client            *storage.S3Client
	JWTSecret           string
}

// Routes constructs and returns the application router including all routes and middleware
func Routes(deps RouterDeps) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(CORS)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Public auth endpoints
	r.Post("/auth/register", HandleRegister(deps.UserRepo, deps.TokenRepo, deps.JWTSecret))
	r.Post("/auth/login", HandleLogin(deps.UserRepo, deps.TokenRepo, deps.JWTSecret))

	// Protected routes
	r.Group(func(pr chi.Router) {
		pr.Use(AuthMiddleware(deps.TokenRepo, deps.JWTSecret))

		pr.Get("/auth/verify", HandleVerifyAuth(deps.UserRepo))
		pr.Post("/auth/logout", HandleLogOut(deps.TokenRepo))

		// Uploads
		pr.Post("/uploads/presign", HandleGetPresignedUploadURL(deps.S3Client))

		// Categories
		pr.Route("/categories", func(cr chi.Router) {
			cr.Get("/", HandleGetAllCategories(deps.CategoryRepo))
			cr.Post("/", HandleCreateCategory(deps.CategoryRepo))
			cr.Get("/{category_id}", HandleGetCategoryByID(deps.CategoryRepo))
			cr.Get("/{category_id}/posts", HandleGetPostsByCategory(deps.PostRepo, deps.ReactionRepo, deps.ReactionTypeRepo))
			cr.Post("/{category_id}/posts", HandleCreatePost(deps.PostRepo))
			cr.Get("/{category_id}/posts/user", HandleGetUserPostsByCategory(deps.PostRepo, deps.ReactionRepo, deps.ReactionTypeRepo))
			cr.Get("/{category_id}/comments/user", HandleGetUserCommentsByCategory(deps.CommentRepo, deps.UserRepo, deps.CommentReactionRepo, deps.ReactionTypeRepo))
		})

		// Posts
		pr.Route("/posts", func(pr chi.Router) {
			pr.Get("/{post_id}", HandleGetPost(deps.PostRepo, deps.ReactionRepo, deps.ReactionTypeRepo))
			pr.Put("/{post_id}", HandleUpdatePost(deps.PostRepo))
			pr.Delete("/{post_id}", HandleDeletePost(deps.PostRepo))
			pr.Post("/{post_id}/react", HandleReactToPost(deps.ReactionRepo))
			pr.Get("/{post_id}/comments", HandleGetCommentsByPost(deps.CommentRepo, deps.UserRepo, deps.CommentReactionRepo, deps.ReactionTypeRepo, deps.PostRepo))
			pr.Post("/{post_id}/comments", HandleCreateCommentOnPost(deps.CommentRepo, deps.PostRepo))
		})

		// Comments
		pr.Route("/comments", func(cr chi.Router) {
			cr.Get("/{comment_id}", HandleGetComment(deps.CommentRepo, deps.UserRepo, deps.CommentReactionRepo, deps.ReactionTypeRepo))
			cr.Put("/{comment_id}", HandleUpdateComment(deps.CommentRepo))
			cr.Delete("/{comment_id}", HandleDeleteComment(deps.CommentRepo))
			cr.Get("/{comment_id}/replies", HandleGetRepliesByComment(deps.CommentRepo, deps.UserRepo, deps.CommentReactionRepo, deps.ReactionTypeRepo))
			cr.Post("/{comment_id}/replies", HandleCreateReplyToComment(deps.CommentRepo))
			cr.Post("/{comment_id}/react", HandleReactToComment(deps.CommentRepo, deps.CommentReactionRepo, deps.ReactionTypeRepo))
		})

		// User-scoped resources
		pr.Get("/user/posts", HandleGetUserPosts(deps.PostRepo, deps.ReactionRepo, deps.ReactionTypeRepo))
		pr.Get("/user/comments", HandleGetUserComments(deps.CommentRepo, deps.UserRepo, deps.CommentReactionRepo, deps.ReactionTypeRepo))
		pr.Get("/user/comments/category/{category_id}", HandleGetUserCommentsByCategory(deps.CommentRepo, deps.UserRepo, deps.CommentReactionRepo, deps.ReactionTypeRepo))
		pr.Get("/user/categories", HandleGetUserCategories(deps.MembershipRepo, deps.CategoryRepo))
		pr.Post("/user/subscribe", HandleSubscribeCategory(deps.UserRepo, deps.CategoryRepo, deps.MembershipRepo))
		pr.Post("/user/unsubscribe", HandleUnsubscribeCategory(deps.MembershipRepo))
		pr.Put("/user/profile-picture", HandleUploadProfilePicture(deps.UserRepo))
		pr.Delete("/user/profile-picture", HandleDeleteProfilePicture(deps.UserRepo))
		pr.Put("/user/username", HandleUpdateUsername(deps.UserRepo))
		pr.Delete("/user", HandleDeleteAccount(deps.UserRepo))

		// Users
		pr.Get("/users/{user_id}", HandleGetAccount(deps.UserRepo))

		// Notifications
		pr.Route("/notifications", func(nr chi.Router) {
			nr.Get("/", HandleGetAllUserNotifications(deps.NotificationRepo))
			nr.Get("/read", HandleGetAllReadNotifications(deps.NotificationRepo))
			nr.Get("/unread", HandleGetAllUnreadNotifications(deps.NotificationRepo))
			nr.Put("/{notification_id}/read", HandleMarkNotificationAsRead(deps.NotificationRepo))
			nr.Put("/{notification_id}/unread", HandleMarkNotificationAsUnread(deps.NotificationRepo))
		})
	})

	return r
}
