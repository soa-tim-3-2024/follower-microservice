package main

import (
	"context"
	"fmt"
	follower "followersModule/proto"
	repository "followersModule/repository"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	logger := log.New(os.Stdout, "[followers-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[followers-store] ", log.LstdFlags)

	store, err := repository.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseDriverConnection(timeoutContext)
	store.CheckConnection()
	// FollowersHandler := handlers.NewFollowersHandler(logger, store)

	// //Initialize the router and add a middleware for all the requests
	// router := mux.NewRouter()
	// router.Use(FollowersHandler.MiddlewareContentTypeSet)

	// postUserRouter := router.Methods(http.MethodPost).Subrouter()
	// postUserRouter.HandleFunc("/user", FollowersHandler.CreateUser)
	// postUserRouter.Use(FollowersHandler.MiddlewarePersonDeserialization)

	// getUserRouter := router.Methods(http.MethodGet).Subrouter()
	// getUserRouter.HandleFunc("/user/{userId}", FollowersHandler.GetUser)

	// postFollowingRouter := router.Methods(http.MethodPost).Subrouter()
	// postFollowingRouter.HandleFunc("/following", FollowersHandler.CreateFollowing)
	// postFollowingRouter.Use(FollowersHandler.MiddlewareNewFollowingDeserialization)

	// deleteFollowingRouter := router.Methods(http.MethodPut).Subrouter()
	// deleteFollowingRouter.HandleFunc("/unfollow", FollowersHandler.UnfollowUser)
	// deleteFollowingRouter.Use(FollowersHandler.MiddlewareUnfollowUserDeserialization)

	// //koga trenutni korisnik (userId) prati
	// getFollowingsRouter := router.Methods(http.MethodGet).Subrouter()
	// getFollowingsRouter.HandleFunc("/user-followings/{userId}", FollowersHandler.GetFollowingsForUser)

	// //ko sve trenutnog korisnika (userId) prati
	// getFollowersRouter := router.Methods(http.MethodGet).Subrouter()
	// getFollowersRouter.HandleFunc("/user-followers/{userId}", FollowersHandler.GetFollowersForUser)

	// //korisnici koje prate pratioci korisnika (userId) 😵 - preporuke za nove pratioce
	// getRecommendationsRouter := router.Methods(http.MethodGet).Subrouter()
	// getRecommendationsRouter.HandleFunc("/user-recommendations/{userId}", FollowersHandler.GetRecommendationsForUser)

	// cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	// server := http.Server{
	// 	Addr:         ":8089",
	// 	Handler:      cors(router),
	// 	IdleTimeout:  120 * time.Second,
	// 	ReadTimeout:  5 * time.Second,
	// 	WriteTimeout: 5 * time.Second,
	// }

	// logger.Println("Server listening on port 8089")
	// //Distribute all the connections to goroutines
	// go func() {
	// 	err := server.ListenAndServe()
	// 	if err != nil {
	// 		logger.Fatal(err)
	// 	}
	// }()

	// sigCh := make(chan os.Signal)
	// signal.Notify(sigCh, os.Interrupt)
	// signal.Notify(sigCh, os.Kill)

	// sig := <-sigCh
	// logger.Println("Received terminate, graceful shutdown", sig)

	// //Try to shutdown gracefully
	// if server.Shutdown(timeoutContext) != nil {
	// 	logger.Fatal("Cannot gracefully shutdown...")
	// }
	// logger.Println("Server stopped")

	listener, err := net.Listen("tcp", "localhost:8089")
	if err != nil {
		log.Fatalln(err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	followerServer := follower.FollowerServer{Repo: *store}
	follower.RegisterFollowersServer(grpcServer, followerServer)

	fmt.Print("server started")
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("server error: ", err)
		}
	}()

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, syscall.SIGTERM)

	<-stopCh

	grpcServer.Stop()
}
