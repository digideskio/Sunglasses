package mask

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestCreateComment(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	userTmp := NewUser()
	userTmp.Username = "testing_very_hard"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	tokenTmp := new(Token)
	tokenTmp.Type = UserToken
	tokenTmp.Expires = float64(time.Now().Unix() + int64(3600*time.Second))
	tokenTmp.UserID = userTmp.ID
	if err := tokenTmp.Save(conn); err != nil {
		panic(err)
	}

	post := NewPost(PostStatus, user)
	post.Text = "A fancy post"
	post.Privacy = PrivacySettings{Type: PrivacyFollowingOnly}
	if err := post.Save(conn); err != nil {
		panic(err)
	}

	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		conn.Db.C("comments").RemoveAll(nil)
		conn.Db.C("users").RemoveAll(nil)
		conn.Db.C("tokens").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Creating new comments", t, func() {
		Convey("When no user is passed", func() {
			testPostHandler(CreateComment, nil, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidData)
				So(errResp.Message, ShouldEqual, MsgInvalidData)
			})
		})

		Convey("When an invalid post id is passed", func() {
			testPostHandler(CreateComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("post_id", "")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidData)
				So(errResp.Message, ShouldEqual, MsgInvalidData)
			})
		})

		Convey("When a post id that doesn't exist is passed", func() {
			testPostHandler(CreateComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("post_id", bson.NewObjectId().Hex())
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 404)
				So(errResp.Code, ShouldEqual, CodeNotFound)
				So(errResp.Message, ShouldEqual, MsgNotFound)
			})
		})

		Convey("When the post can't be accessed by the user", func() {
			testPostHandler(CreateComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("post_id", post.ID.Hex())
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 403)
				So(errResp.Code, ShouldEqual, CodeUnauthorized)
				So(errResp.Message, ShouldEqual, MsgUnauthorized)
			})
		})

		Convey("When the comment text is not valid", func() {
			FollowUser(user.ID, userTmp.ID, conn)

			testPostHandler(CreateComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("post_id", post.ID.Hex())
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidCommentText)
				So(errResp.Message, ShouldEqual, MsgInvalidCommentText)
			})
		})

		Convey("When everything is OK", func() {
			testPostHandler(CreateComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", post.ID.Hex())
				r.PostForm.Add("comment_text", "My fancy comment")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				So(res.Code, ShouldEqual, 201)
			})
		})
	})
}
