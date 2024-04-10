package main

import (
	"context"
	"followersModule/handlers"
	repository "followersModule/repository"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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
	FollowersHandler := handlers.NewFollowersHandler(logger, store)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(FollowersHandler.MiddlewareContentTypeSet)

	postUserRouter := router.Methods(http.MethodPost).Subrouter()
	postUserRouter.HandleFunc("/user", FollowersHandler.CreateUser)
	postUserRouter.Use(FollowersHandler.MiddlewarePersonDeserialization)

	getUserRouter := router.Methods(http.MethodGet).Subrouter()
	getUserRouter.HandleFunc("/user/{userId}", FollowersHandler.GetUser)

	postFollowingRouter := router.Methods(http.MethodPost).Subrouter()
	postFollowingRouter.HandleFunc("/following", FollowersHandler.CreateFollowing)
	postFollowingRouter.Use(FollowersHandler.MiddlewareNewFollowingDeserialization)

	deleteFollowingRouter := router.Methods(http.MethodPut).Subrouter()
	deleteFollowingRouter.HandleFunc("/unfollow", FollowersHandler.UnfollowUser)
	deleteFollowingRouter.Use(FollowersHandler.MiddlewareUnfollowUserDeserialization)

	//koga trenutni korisnik (userId) prati
	getFollowingsRouter := router.Methods(http.MethodGet).Subrouter()
	getFollowingsRouter.HandleFunc("/user-followings/{userId}", FollowersHandler.GetFollowingsForUser)

	//ko sve trenutnog korisnika (userId) prati
	getFollowersRouter := router.Methods(http.MethodGet).Subrouter()
	getFollowersRouter.HandleFunc("/user-followers/{userId}", FollowersHandler.GetFollowersForUser)

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	server := http.Server{
		Addr:         ":8089",
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logger.Println("Server listening on port 8089")
	//Distribute all the connections to goroutines
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
	logger.Println("Server stopped")
}
