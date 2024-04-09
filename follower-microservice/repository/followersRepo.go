package repository

import (
	"context"
	"followersModule/model"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// NoSQL: MovieRepo struct encapsulating Neo4J api client
type FollowersRepo struct {
	// Thread-safe instance which maintains a database connection pool
	driver neo4j.DriverWithContext
	logger *log.Logger
}

// NoSQL: Constructor which reads db configuration from environment and creates a keyspace
func New(logger *log.Logger) (*FollowersRepo, error) {
	// Local instance
	uri := "bolt://localhost:7687"
	user := "neo4j"
	pass := "Dejann03"
	auth := neo4j.BasicAuth(user, pass, "")

	driver, err := neo4j.NewDriverWithContext(uri, auth)
	if err != nil {
		logger.Panic(err)
		return nil, err
	}

	// Return repository with logger and DB session
	return &FollowersRepo{
		driver: driver,
		logger: logger,
	}, nil
}

// Check if connection is established
func (mr *FollowersRepo) CheckConnection() {
	ctx := context.Background()
	err := mr.driver.VerifyConnectivity(ctx)
	if err != nil {
		mr.logger.Panic(err)
		return
	}
	// Print Neo4J server address
	mr.logger.Printf(`Neo4J server address: %s`, mr.driver.Target().Host)
}

// Disconnect from database
func (mr *FollowersRepo) CloseDriverConnection(ctx context.Context) {
	mr.driver.Close(ctx)
}

func (mr *FollowersRepo) SaveFollowing(user *model.User, userToFollow *model.User) error {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "baza"})
	defer session.Close(ctx)
	mr.SaveUser(user)
	mr.SaveUser(userToFollow)
	_, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (a:User), (b:User) WHERE a.username = $userUsername AND b.username = $followUsername CREATE (a)-[r: IS_FOLLOWING]->(b) RETURN type(r)",
				map[string]any{"userUsername": user.Username, "followUsername": userToFollow.Username})
			if err != nil {
				return nil, err
			}
			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}
			return nil, result.Err()
		})
	if err != nil {
		mr.logger.Println("Error inserting following:", err)
		return err
	}
	return nil
}

// pokusava da sacuva korisnika ako ne postoji u bazi, ako postoji nece ga cuvati
func (mr *FollowersRepo) SaveUser(user *model.User) (bool, error) {
	userInDatabase, err := mr.ReadUser(user.UserId)
	if (userInDatabase == model.User{}) {
		err = mr.WriteUserToDatabase(user)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func (mr *FollowersRepo) WriteUserToDatabase(user *model.User) error {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "baza"}) //baza podataka na koju se povezujem
	defer session.Close(ctx)
	newUser, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"CREATE (u:User) SET u.userId = $userId, u.username = $username, u.profileImage = $profileImage RETURN u.username + ', from node ' + id(u)",
				map[string]any{"userId": user.UserId, "username": user.Username, "profileImage": user.ProfileImage})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}

			return nil, result.Err()
		})
	if err != nil {
		mr.logger.Println("Error inserting Person:", err)
		return err
	}
	mr.logger.Println(newUser.(string))
	return nil
}

func (mr *FollowersRepo) ReadUser(userId string) (model.User, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "baza"})
	defer session.Close(ctx)
	user, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (u {userId: $userId}) RETURN u.userId, u.username, u.profileImage",
				map[string]any{"userId": userId})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values, nil
			}

			return nil, result.Err()
		})
	if err != nil {
		mr.logger.Println("Error reading user:", err)
		return model.User{}, err
	}
	if user == nil {
		return model.User{}, nil
	}
	var id, username, profileImage string
	for _, value := range user.([]interface{}) {
		if val, ok := value.(string); ok {
			if id == "" {
				id = val
			} else if username == "" {
				username = val
			} else if profileImage == "" {
				profileImage = val
			}
		}
	}
	userFromDatabase := model.User{
		UserId:       id,
		Username:     username,
		ProfileImage: profileImage,
	}

	return userFromDatabase, nil
}
