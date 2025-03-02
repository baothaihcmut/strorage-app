package main

import (
	"log"

	"github.com/baothaihcmut/Bibox/storage-app/internal/config"
	commentControllers "github.com/baothaihcmut/Bibox/storage-app/internal/modules/file_comment/controllers"

	commentInteractors "github.com/baothaihcmut/Bibox/storage-app/internal/modules/file_comment/interactors"
	commentRepo "github.com/baothaihcmut/Bibox/storage-app/internal/modules/file_comment/repositories"
	permControllers "github.com/baothaihcmut/Bibox/storage-app/internal/modules/file_permission/controllers"
	permInteractors "github.com/baothaihcmut/Bibox/storage-app/internal/modules/file_permission/interactors"
	permRepo "github.com/baothaihcmut/Bibox/storage-app/internal/modules/file_permission/repositories"
	"github.com/baothaihcmut/Bibox/storage-app/internal/server"
	"github.com/baothaihcmut/Bibox/storage-app/internal/server/initialize"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// @title Storage App API
// @version 1.0
// @description This is a sample API for file storage
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load configuration
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	// Initialize logger
	logger := initialize.InitializeLogger(&config.Logger)

	// Initialize Gin engine
	g := gin.Default()

	// Initialize MongoDB
	mongoClient, err := initialize.InitializeMongo(&config.Mongo)
	if err != nil {
		logger.Panic(err)
		log.Fatal("Failed to initialize MongoDB:", err)
	}

	// Select the database
	mongoDatabase := mongoClient.Database(config.Mongo.DatabaseName)

	// Initialize OAuth2 (Google & GitHub)
	oauth2Google := initialize.InitializeOauth2(&config.Oauth2.Google, []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	}, google.Endpoint)

	oauth2Github := initialize.InitializeOauth2(&config.Oauth2.Github, []string{
		"read:user", "user:email",
	}, github.Endpoint)

	// Initialize S3 Storage
	s3, err := initialize.InitalizeS3(config.S3)
	if err != nil {
		logger.Panic(err)
		log.Fatal("Failed to initialize S3:", err)
	}

	// Create a new server instance
	s := server.NewServer(g, mongoClient, oauth2Google, oauth2Github, s3, logger, config)

	// Initialize PermissionRepository & Interactor
	permissionRepository := permRepo.NewPermissionRepository(mongoDatabase)
	permissionInteractor := permInteractors.NewPermissionInteractor(permissionRepository)
	permissionController := permControllers.NewPermissionController(permissionInteractor)

	//  Initialize Comment Repository & Interactor
	commentRepository := commentRepo.NewCommentRepository(mongoDatabase)
	commentInteractor := commentInteractors.NewCommentInteractor(commentRepository, permissionRepository)
	commentController := commentControllers.NewCommentController(commentInteractor)

	//  Set up routes for file permission & comment
	server.SetupRoutes(g, permissionController, commentController)

	// Run the server
	go s.Run()

}
