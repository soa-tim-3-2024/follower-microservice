package handlers

import (
	"context"
	"followersModule/model"
	"followersModule/repository"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type KeyProduct struct{}

type FollowersHandler struct {
	logger *log.Logger
	repo   *repository.FollowersRepo
}

func NewFollowersHandler(l *log.Logger, r *repository.FollowersRepo) *FollowersHandler {
	return &FollowersHandler{l, r}
}

func (f *FollowersHandler) CreateUser(rw http.ResponseWriter, h *http.Request) {
	user := h.Context().Value(KeyProduct{}).(*model.User)
	userSaved, err := f.repo.SaveUser(user)
	if err != nil {
		f.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userSaved {
		f.logger.Print("New user saved to database")
		rw.WriteHeader(http.StatusCreated)
	} else {
		rw.WriteHeader(http.StatusConflict)
	}
}

func (f *FollowersHandler) CreateFollowing(rw http.ResponseWriter, h *http.Request) {
	newFollowing := h.Context().Value(KeyProduct{}).(*model.NewFollowing)
	user := model.User{}
	userToFollow := model.User{}
	user.UserId = newFollowing.UserId
	user.Username = newFollowing.Username
	user.ProfileImage = newFollowing.ProfileImage
	userToFollow.UserId = newFollowing.FollowingUserId
	userToFollow.Username = newFollowing.FollowingUsername
	userToFollow.ProfileImage = newFollowing.FollowingProfileImage
	err := f.repo.SaveFollowing(&user, &userToFollow)
	if err != nil {
		f.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f *FollowersHandler) GetUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["userId"]
	user, err := f.repo.ReadUser(id)
	if err != nil {
		f.logger.Print("Database exception: ", err)
	}

	err = user.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		f.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (f *FollowersHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		f.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}

func (f *FollowersHandler) MiddlewarePersonDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		user := &model.User{}
		err := user.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			f.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, user)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

func (f *FollowersHandler) MiddlewareNewFollowingDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		newFollowing := &model.NewFollowing{}
		err := newFollowing.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			f.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, newFollowing)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}
