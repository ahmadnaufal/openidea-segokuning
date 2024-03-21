package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ahmadnaufal/openidea-segokuning/internal/config"
	"github.com/ahmadnaufal/openidea-segokuning/internal/friend"
	"github.com/ahmadnaufal/openidea-segokuning/internal/image"
	"github.com/ahmadnaufal/openidea-segokuning/internal/post"
	"github.com/ahmadnaufal/openidea-segokuning/internal/user"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/jwt"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/middleware"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/s3"

	"github.com/ansrivas/fiberprometheus/v2"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.InitializeConfig()

	app := fiber.New(fiber.Config{
		ErrorHandler: config.DefaultErrorHandler(),
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(compress.New())
	// custom middleware to set all method not allowed response to not found
	app.Use(middleware.CustomMiddleware404())

	jwtProvider := jwt.NewJWTProvider(cfg.JWTSecret)

	db := connectToDB(cfg.Database, cfg.Env)

	userRepo := user.NewUserRepo(db)
	friendRepo := friend.NewFriendRepo(db)
	postRepo := post.NewPostRepo(db)

	trxProvider := config.NewTransactionProvider(db)

	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	s3Provider := s3.NewS3Provider(awsCfg, cfg.S3.Bucket, cfg.S3.Region, cfg.S3.ID, cfg.S3.SecretKey)

	imageHandler := image.NewImageHandler(&s3Provider)
	userHandler := user.NewUserHandler(user.UserHandlerConfig{
		UserRepo:    &userRepo,
		JwtProvider: &jwtProvider,
		SaltCost:    cfg.BcryptSalt,
	})
	friendHandler := friend.NewFriendHandler(friend.FriendHandlerConfig{
		UserRepo:   &userRepo,
		FriendRepo: &friendRepo,
	})
	postHandler := post.NewPostHandler(post.PostHandlerConfig{
		PostRepo:   &postRepo,
		TxProvider: &trxProvider,
		FriendRepo: &friendRepo,
	})

	imageHandler.RegisterRoute(app, jwtProvider)
	userHandler.RegisterRoute(app, jwtProvider)
	friendHandler.RegisterRoute(app, jwtProvider)
	postHandler.RegisterRoute(app, jwtProvider)

	// setup instrumentation
	prometheus := fiberprometheus.New("segokuning")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	addr := fmt.Sprintf(":%s", cfg.AppPort)

	log.Fatal(app.Listen(addr))
}

func connectToDB(dbCfg config.DatabaseConfig, env string) *sqlx.DB {
	var dsn string
	if env == "production" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=verify-full&sslrootcert=ap-southeast-1-bundle.pem",
			dbCfg.Username, dbCfg.Password, dbCfg.Host,
			dbCfg.Port, dbCfg.Name,
		)
	} else {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbCfg.Username, dbCfg.Password, dbCfg.Host,
			dbCfg.Port, dbCfg.Name,
		)
	}

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(dbCfg.MaxOpenConnection)
	db.SetMaxIdleConns(dbCfg.MaxIdleConnection)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}
