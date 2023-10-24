package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"webooktrial/internal/domain"
	"webooktrial/internal/integration/startup"
	"webooktrial/internal/repository/dao/article"
	ijwt "webooktrial/internal/web/jwt"
)

type ArticleMongoHandlerTestSuite struct {
	suite.Suite
	server  *gin.Engine
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (s *ArticleMongoHandlerTestSuite) SetupSuite() {
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Uid: 123,
		})
		ctx.Next()
	})
	s.mdb = startup.InitMongoDB()
	node, err := snowflake.NewNode(1)
	assert.NoError(s.T(), err)
	if err != nil {
		panic(err)
	}
	s.col = s.mdb.Collection("articles")
	s.liveCol = s.mdb.Collection("published_articles")
	hdl := startup.InitArticleHandler(article.NewMongoDBDAO(s.mdb, node))
	hdl.RegisterRoutes(s.server)
}

func (s *ArticleMongoHandlerTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := s.mdb.Collection("articles").
		DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.mdb.Collection("published_articles").
		DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func (s *ArticleMongoHandlerTestSuite) TestCleanMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := s.mdb.Collection("articles").
		DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.mdb.Collection("published_articles").
		DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func (s *ArticleMongoHandlerTestSuite) TestArticleHandler_Edit() {
	t := s.T()
	testCases := []struct {
		name string
		// 提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T) // 构造请求，直接使用 req
		// 也就是说，我们放弃测试 Bind 的异常分支
		req Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子-保存成功",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				// ID为雪花算法生成，没法直接拿到。每个作者只有一篇文章，直接用作者ID查询
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				art.Id = 0
				assert.Equal(t, article.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			req: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			name: "修改已有帖子，并保存",
			before: func(t *testing.T) {
				// 提前准备数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 模拟已经存在的帖子，并且是已经发布的帖子
				_, err := s.col.InsertOne(ctx, &article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    234,
					Utime:    456,
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据库
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
				assert.NoError(t, err)
				// 是为了确保我更新了 Utime
				assert.True(t, art.Utime > 456)
				art.Utime = 0
				assert.Equal(t, article.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					Ctime:    234,
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			req: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "修改别人的帖子",
			before: func(t *testing.T) {
				// 提前准备数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				_, err := s.col.InsertOne(ctx, article.Article{
					Id:      3,
					Title:   "我的标题",
					Content: "我的内容",
					// 测试模拟的用户 ID 是123，这里是 789
					// 意味着你在修改别人的数据
					AuthorId: 789,
					// 跟时间有关的测试，不是逼不得已，不要用 time.Now()
					// 因为 time.Now() 每次运行都不同，你很难断言
					Ctime: 234,
					Utime: 456,
					// 为了验证状态没有变
					Status: domain.ArticleStatusPublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据库
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    234,
					Utime:    456,
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
			},
			req: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 构造请求
			// 执行
			// 验证结果
			tc.before(t)
			reqBody, err := json.Marshal(&tc.req)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			// 这里你就可以继续使用 req
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			code := resp.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			var result Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult.Code, result.Code)
			// 只能判定有 ID，因为雪花算法你无法确定具体的值
			if tc.wantResult.Data > 0 {
				assert.True(t, result.Data > 0)
			}
			tc.after(t)
		})
	}
}
func TestMongoArticle(t *testing.T) {
	suite.Run(t, new(ArticleMongoHandlerTestSuite))
}

func (s *ArticleMongoHandlerTestSuite) TestArticle_Publish() {
	t := s.T()
	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)
		req   Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {
				// 什么也不需要做
			},
			after: func(t *testing.T) {
				// 验证一下数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.Equal(t, "hello，你好", art.Title)
				assert.Equal(t, "随便试试", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				var publishedArt article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.True(t, publishedArt.Id > 0)
				assert.Equal(t, "hello，你好", publishedArt.Title)
				assert.Equal(t, "随便试试", publishedArt.Content)
				assert.Equal(t, int64(123), publishedArt.AuthorId)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
			},
			req: Article{
				Title:   "hello，你好",
				Content: "随便试试",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			// 制作库有，但是线上库没有
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 模拟已经存在的帖子，并且是已经发布的帖子
				_, err := s.col.InsertOne(ctx, &article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 123,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, int64(2), art.Id)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 234)
				var publishedArt article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, int64(2), art.Id)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
			},
			req: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "更新帖子，并且重新发表",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				art := article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 123,
				}
				// 模拟已经存在的帖子，并且是已经发布的帖子
				_, err := s.col.InsertOne(ctx, &art)
				assert.NoError(t, err)
				part := article.PublishedArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &part)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, int64(3), art.Id)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 234)

				var part article.PublishedArticle
				err = s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&part)
				assert.NoError(t, err)
				assert.Equal(t, int64(3), part.Id)
				assert.Equal(t, "新的标题", part.Title)
				assert.Equal(t, "新的内容", part.Content)
				assert.Equal(t, int64(123), part.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), part.Ctime)
				// 更新时间变了
				assert.True(t, part.Utime > 234)
			},
			req: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 3,
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				art := article.Article{
					Id:      4,
					Title:   "我的标题",
					Content: "我的内容",
					Ctime:   456,
					Utime:   234,
					// 注意。这个 AuthorID 我们设置为另外一个人的ID
					AuthorId: 789,
				}
				// 模拟已经存在的帖子，并且是已经发布的帖子
				_, err := s.col.InsertOne(ctx, &art)
				assert.NoError(t, err)
				part := article.PublishedArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &part)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 更新应该是失败了，数据没有发生变化
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, int64(4), art.Id)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, int64(234), art.Utime)
				assert.Equal(t, int64(789), art.AuthorId)

				var part article.PublishedArticle
				// 数据没有变化
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).Decode(&part)
				assert.NoError(t, err)
				assert.Equal(t, int64(4), part.Id)
				assert.Equal(t, "我的标题", part.Title)
				assert.Equal(t, "我的内容", part.Content)
				assert.Equal(t, int64(789), part.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), part.Ctime)
				// 更新时间变了
				assert.Equal(t, int64(234), part.Utime)
			},
			req: Article{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			// 不能有 error
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			// 反序列化为结果
			// 利用泛型来限定结果必须是 int64
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult.Code, result.Code)
			// 只能判定有 ID，因为雪花算法你无法确定具体的值
			if tc.wantResult.Data > 0 {
				assert.True(t, result.Data > 0)
			}
			tc.after(t)
		})
	}
}
