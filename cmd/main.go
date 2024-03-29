package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		Prefork:      false,
		Concurrency:  1024 * 1024,
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(compress.New())
	// custom middleware to set all method not allowed response to not found
	app.Use(middleware.CustomMiddleware404())

	jwtProvider := jwt.NewJWTProvider(cfg.JWTSecret)

	db := connectToDB(cfg.Database)

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
		TxProvider: &trxProvider,
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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// start the server
		if err := app.Listen(addr); err != nil {
			log.Println("failed to start server: ", err)
			os.Exit(1)
		}
	}()

	receivedSignal := <-sig

	log.Printf("received %v. Stopping app...", receivedSignal)
	if err := app.Shutdown(); err != nil {
		log.Println("failed to shutdown server: ", err)
		os.Exit(1)
	}

	log.Println("App successfully stopped.")
}

func connectToDB(dbCfg config.DatabaseConfig) *sqlx.DB {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?%s",
		dbCfg.Username, dbCfg.Password, dbCfg.Host,
		dbCfg.Port, dbCfg.Name, dbCfg.Params,
	)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(dbCfg.MaxOpenConnection)
	db.SetMaxIdleConns(dbCfg.MaxIdleConnection)
	db.SetConnMaxLifetime(time.Duration(dbCfg.MaxConnLifetime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(dbCfg.MaxConnIdleTime) * time.Minute)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}
