package repositories

import (
	"context"

	"github.com/baothaihcmut/Bibox/storage-app/internal/common/enums"
	"github.com/baothaihcmut/Bibox/storage-app/internal/common/logger"
	"github.com/baothaihcmut/Bibox/storage-app/internal/modules/file_permission/models"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FilterPermssionOption int

const (
	PermssionInList FilterPermssionOption = iota
	PermssionLessThan
	PermssionGreaterThan
	PermssionEqual
	PermssionNotEqual
	PermssionNotInList
)

type FilterPermssionType struct {
	Option FilterPermssionOption
	Value  []enums.FilePermissionType
}

type FilePermissionRepository interface {
	UpdatePermission(context.Context, primitive.ObjectID, primitive.ObjectID, enums.FilePermissionType, bool, bool) error
	GetFileByID(ctx context.Context, fileID primitive.ObjectID) (*models.FilePermission, error)
	CreateFilePermission(ctx context.Context, filePermission *models.FilePermission) error
	GetFilePermission(ctx context.Context, fileID primitive.ObjectID, userID primitive.ObjectID) (*models.FilePermission, error)
	GetPermissionOfFileWithUserInfo(ctx context.Context, fileId primitive.ObjectID) ([]*models.FilePermissionWithUser, error)
	GetPermissionOfFile(ctx context.Context, fileID primitive.ObjectID) ([]*models.FilePermission, error)
	BulkCreatePermission(ctx context.Context, filePermissions []*models.FilePermission) error
}

type FilePermissionRepositoryImpl struct {
	collection *mongo.Collection
	logger     logger.Logger
}

// BulkCreatePermission implements FilePermissionRepository.
func (pr *FilePermissionRepositoryImpl) BulkCreatePermission(ctx context.Context, filePermissions []*models.FilePermission) error {
	_, err := pr.collection.InsertMany(ctx, lo.Map(filePermissions, func(item *models.FilePermission, _ int) interface{} {
		return item
	}))
	if err != nil {
		return err
	}
	return nil

}

// GetPermissionOfFile implements FilePermissionRepository.
func (pr *FilePermissionRepositoryImpl) GetPermissionOfFile(ctx context.Context, fileID primitive.ObjectID) ([]*models.FilePermission, error) {
	cursor, err := pr.collection.Find(ctx, bson.D{
		{Key: "file_id", Value: fileID},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var res []*models.FilePermission
	if err := cursor.All(ctx, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func NewPermissionRepository(collection *mongo.Collection, logger logger.Logger) FilePermissionRepository {
	return &FilePermissionRepositoryImpl{
		collection: collection,
		logger:     logger,
	}
}

// update permissino
func (pr *FilePermissionRepositoryImpl) UpdatePermission(
	ctx context.Context,
	fileID primitive.ObjectID,
	userID primitive.ObjectID,
	permissionType enums.FilePermissionType,
	accessSecure bool,
	canShare bool,
) error {
	filter := bson.M{"file_id": fileID, "user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"permission_type":    permissionType,
			"access_secure_file": accessSecure,
			"can_share":          canShare,
		},
	}

	_, err := pr.collection.UpdateOne(ctx, filter, update)
	return err
}

// get file by ID to check ownership
func (pr *FilePermissionRepositoryImpl) GetFileByID(ctx context.Context, fileID primitive.ObjectID) (*models.FilePermission, error) {
	var file models.FilePermission
	err := pr.collection.FindOne(ctx, bson.M{"file_id": fileID}).Decode(&file)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // File not found
		}
		return nil, err
	}
	return &file, nil
}

// insert file permission into DB
func (pr *FilePermissionRepositoryImpl) CreateFilePermission(ctx context.Context, filePermission *models.FilePermission) error {
	_, err := pr.collection.InsertOne(ctx, filePermission)
	if err != nil {
		pr.logger.Errorf(ctx, map[string]any{
			"file_id": filePermission.FileID,
			"user_id": filePermission.UserID,
		}, "Error when insert file permission: ", err)
		return err
	}
	return nil
}

// get file permission by fileID and userID and permssion type
func (pr *FilePermissionRepositoryImpl) GetFilePermission(ctx context.Context, fileID primitive.ObjectID, userID primitive.ObjectID) (*models.FilePermission, error) {
	// build filter
	var file models.FilePermission
	err := pr.collection.FindOne(ctx, bson.M{
		"file_id": fileID,
		"user_id": userID,
	}).Decode(&file)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // File permission not found
		}
		return nil, err
	}
	return &file, nil
}
func (pr *FilePermissionRepositoryImpl) GetPermissionOfFileWithUserInfo(ctx context.Context, fileId primitive.ObjectID) ([]*models.FilePermissionWithUser, error) {
	pipeline := mongo.Pipeline{
		//match stage
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "file_id", Value: fileId},
			}},
		},
		//lookup stage
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "user_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "user"},
			}},
		},
		bson.D{{Key: "$unwind", Value: "$user"}},
		//projection stage
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "file_id", Value: 1},
				{Key: "user_id", Value: 1},
				{Key: "permission_type", Value: 1},
				{Key: "can_share", Value: 1},
				{Key: "access_secure_file", Value: 1},
				{Key: "user.first_name", Value: 1},
				{Key: "user.last_name", Value: 1},
				{Key: "user.email", Value: 1},
				{Key: "user.image", Value: 1},
			}},
		},
	}

	cursor, err := pr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var result []*models.FilePermissionWithUser
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil

}
