package svc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureReviewIndexes 确保 review 依赖的 MongoDB 集合索引存在。
// 索引为库级对象，任意进程建一次即可全局生效；重复执行是幂等的，
// 已存在的索引会被识别并跳过，仅记录日志。
func EnsureReviewIndexes(url, db string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	// FindByGroup 组合过滤：createAt 恒必选且放最前，
	// 保证所有调用至少能用 createAt 前缀做范围扫描；
	// group/school/grade 为可选等值条件，在索引扫描后过滤。
	if err := ensureIndex(ctx, client, db, "entry_form", "entry_form_createAt_group_school_grade", false,
		bson.D{{Key: "createAt", Value: 1}, {Key: "group", Value: 1},
			{Key: "school", Value: 1}, {Key: "grade", Value: 1}}); err != nil {
		return err
	}

	// FindByUserIds / FindOneByUserId 的 $in 与单条查询；
	// user_id 唯一索引用于兜底 schedule 并发双写（配合 UpsertByUserId）。
	// 注意：若数据库已存在同名的非唯一 schedule_user_id，需先 drop 旧索引，
	// 否则 unique 索引不会建上（indexExists 按名命中后跳过）。
	return ensureIndex(ctx, client, db, "schedule", "schedule_user_id", true,
		bson.D{{Key: "user_id", Value: 1}})
}

func ensureIndex(ctx context.Context, client *mongo.Client, db, collection, name string, unique bool, keys bson.D) error {
	coll := client.Database(db).Collection(collection)

	exists, err := indexExists(ctx, coll, name)
	if err != nil {
		return fmt.Errorf("list index %s: %w", name, err)
	}
	if exists {
		logx.Infof("index %s already exists, skip", name)
		return nil
	}

	opts := options.Index().SetName(name)
	if unique {
		opts = opts.SetUnique(true)
	}

	_, err = coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    keys,
		Options: opts,
	})
	if err != nil {
		if isIndexExists(err) {
			logx.Infof("index %s already exists, skip", name)
			return nil
		}
		return fmt.Errorf("create index %s: %w", name, err)
	}
	logx.Infof("index %s created", name)
	return nil
}

// indexExists 遍历集合索引，判断指定名称的索引是否已存在。
func indexExists(ctx context.Context, coll *mongo.Collection, name string) (bool, error) {
	cur, err := coll.Indexes().List(ctx)
	if err != nil {
		return false, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var idx bson.M
		if err := cur.Decode(&idx); err != nil {
			return false, err
		}
		if idx["name"] == name {
			return true, nil
		}
	}
	return false, cur.Err()
}

func isIndexExists(err error) bool {
	var se mongo.ServerError
	if errors.As(err, &se) {
		// 48 NamespaceExists, 85 IndexOptionsConflict, 86 IndexKeySpecsConflict
		return se.HasErrorCode(48) || se.HasErrorCode(85) || se.HasErrorCode(86)
	}
	return false
}
