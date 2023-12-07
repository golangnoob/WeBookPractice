package article

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBDao struct {
	client *mongo.Client
	// 代表 webook 的
	database *mongo.Database
	// 代表的是制作库
	col *mongo.Collection
	// 代表的是线上库
	liveCol *mongo.Collection
	node    *snowflake.Node

	idGen IDGenerator
}

func (m *MongoDBDao) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

type IDGenerator func() int64

func (m *MongoDBDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	type IDGenerator func() int64

	id := m.node.Generate().Int64()
	art.Id = id
	_, err := m.col.InsertOne(ctx, art)
	// 你没有自增主键
	// GLOBAL UNIFY ID (GUID，全局唯一ID）
	return id, err
}

func (m *MongoDBDao) UpdateById(ctx context.Context, art Article) error {
	// 操作制作库
	filter := bson.M{"id": art.Id,
		"author_id": art.AuthorId}
	update := bson.D{bson.E{"$set", bson.M{
		"title":   art.Title,
		"content": art.Content,
		"utime":   time.Now().UnixMilli(),
		"status":  art.Status,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// 这边就是校验了 author_id 是不是正确的 ID
	if res.ModifiedCount == 0 {
		return errors.New("更新数据失败")
	}
	return nil

}

func (m *MongoDBDao) GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	var arts []Article
	//cur, err := m.col.Aggregate(ctx, bson.D{
	//  bson.E{Key: "$match", Value: author},
	//	bson.E{Key: "$limit", Value: limit},
	//	bson.E{Key: "$skip", Value: offset}
	//	})
	var findOptions *options.FindOptions
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
		findOptions.SetSkip(int64(offset))
	}
	cur, err := m.col.Find(ctx, bson.D{bson.E{Key: "author_id", Value: author}}, findOptions)
	defer cur.Close(ctx)
	if err == nil {
		err = cur.All(ctx, &arts)
	}
	return arts, err
}

func (m *MongoDBDao) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := m.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: id}}).Decode(&art)
	return art, err
}

func (m *MongoDBDao) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pub PublishedArticle
	err := m.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: id}}).Decode(&pub)
	return pub, err
}

func (m *MongoDBDao) Sync(ctx context.Context, art Article) (int64, error) {
	// 没法子引入事务的概念
	// 首先第一步，保存制作库
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	// 操作线上库了, upsert 语义
	now := time.Now().UnixMilli()
	//update := bson.E{"$set", art}
	//upsert := bson.E{"$setOnInsert", bson.D{bson.E{"ctime", now}}}
	art.Utime = now
	updateV1 := bson.M{
		// 更新，如果不存在，就是插入，
		"$set": PublishedArticle(art),
		// 在插入的时候，要插入 ctime
		"$setOnInsert": bson.M{"ctime": now},
	}
	filter := bson.M{"id": art.Id}
	_, err = m.liveCol.UpdateOne(ctx, filter,
		//bson.D{update, upsert},
		updateV1,
		options.Update().SetUpsert(true))
	return id, err
}

func (m *MongoDBDao) SyncStatus(ctx context.Context, author, id int64, status uint8) error {
	sess, err := m.client.StartSession()
	if err != nil {
		return err
	}
	filter := bson.M{"id": id,
		"author_id": author}
	update := bson.D{bson.E{"$set", bson.M{
		"status": status,
	}}}
	sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		res, err := m.col.UpdateOne(sessCtx, filter, update)
		if err != nil {
			return nil, err
		}
		if res.ModifiedCount != 1 {
			return nil, ErrPossibleIncorrectAuthor
		}
		res, err = m.liveCol.UpdateOne(sessCtx, filter, update)
		if err != nil {
			return nil, err
		}
		if res.ModifiedCount != 1 {
			return nil, ErrPossibleIncorrectAuthor
		}
		return nil, err
	})
	return nil
}

func NewMongoDBDAOV1(db *mongo.Database, idGen IDGenerator) ArticleDao {
	return &MongoDBDao{
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		//node:    node,
		idGen: idGen,
	}
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{Key: "author_id", Value: 1},
				bson.E{Key: "ctime", Value: 1},
			},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().
		CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().
		CreateMany(ctx, index)
	return err
}

func NewMongoDBDAO(db *mongo.Database, node *snowflake.Node) ArticleDao {
	return &MongoDBDao{
		client:  db.Client(),
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		node:    node,
	}
}
