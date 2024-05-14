package follower

import (
	context "context"
	"fmt"
	"followersModule/model"
	"followersModule/repository"
)

type FollowerServer struct {
	Repo repository.FollowersRepo
	UnimplementedFollowersServer
}

func (fs FollowerServer) GetUserFollowings(context context.Context, id *Identificator) (*ListFollowingResponse, error) {
	users, err := fs.Repo.GetFollowingsForUser(fmt.Sprint(id.Id))
	if err != nil {
		return nil, err
	}
	var ret ListFollowingResponse
	for _, user := range users {
		ret.List = append(ret.List, &FollowingResponse{UserId: user.UserId, Username: user.Username, ProfileImage: user.ProfileImage})
	}
	return &ret, nil
}
func (fs FollowerServer) GetUserFollowers(context context.Context, id *Identificator) (*ListFollowingResponse, error) {
	users, err := fs.Repo.GetFollowersForUser(fmt.Sprint(id.Id))
	if err != nil {
		return nil, err
	}
	var ret ListFollowingResponse
	for _, user := range users {
		ret.List = append(ret.List, &FollowingResponse{UserId: user.UserId, Username: user.Username, ProfileImage: user.ProfileImage})
	}
	return &ret, nil
}
func (fs FollowerServer) GetUserRecommendations(context context.Context, id *Identificator) (*ListFollowingResponse, error) {
	users, err := fs.Repo.GetRecommendationsForUser(fmt.Sprint(id.Id))
	if err != nil {
		return nil, err
	}
	var ret ListFollowingResponse
	for _, user := range users {
		ret.List = append(ret.List, &FollowingResponse{UserId: user.UserId, Username: user.Username, ProfileImage: user.ProfileImage})
	}
	return &ret, nil
}
func (fs FollowerServer) CreateNewFollowing(context context.Context, followingCreate *FollowingCreateRequest) (*FollowerResponse, error) {
	err := fs.Repo.SaveFollowing(&model.User{UserId: followingCreate.UserId, Username: followingCreate.Username, ProfileImage: followingCreate.ProfileImage},
		&model.User{UserId: followingCreate.FollowingUserId, Username: followingCreate.FollowingUsername, ProfileImage: followingCreate.FollowingProfileImage})
	if err != nil {
		return nil, err
	}
	return &FollowerResponse{}, nil
}
func (fs FollowerServer) UnfollowUser(context context.Context, userUnfollow *UserUnfollowRequest) (*FollowerResponse, error) {
	err := fs.Repo.DeleteFollowing(userUnfollow.UserId, userUnfollow.UserToUnfollowId)
	if err != nil {
		return nil, err
	}
	return &FollowerResponse{}, nil
}
