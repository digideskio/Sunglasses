package lamp

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestUser(t *testing.T) {
	Convey("Subject: Creating an user and removing it", t, func() {
		Convey("Given a database connection", func() {
			config, err := NewConfig("../config.json")
			if err != nil {
				panic(err)
			}
			conn, err := NewDatabaseConn(config)
			if err != nil {
				panic(err)
			}

			Convey("and an user instance", func() {
				user := new(User)
				user.Username = "John Doe"
				err = user.SetPassword("testing")
				if err != nil {
					panic(err)
				}
				user.Role = RoleUser
				user.Active = true

				Convey("User should be saved correctly", func() {
					success, err := user.Save(conn)
					So(success, ShouldEqual, true)
					So(err, ShouldEqual, nil)

					Convey("And the password must match 'testing'", func() {
						valid := user.CheckPassword("testing")
						So(valid, ShouldEqual, true)
					})
				})

				Convey("But with the same username the user should not be inserted", func() {
					uid := user.ID
					user.ID = ""

					success, err := user.Save(conn)
					So(success, ShouldEqual, false)
					So(err, ShouldEqual, nil)

					user.ID = uid
				})

				Convey("Retrieving user info and settings", func() {
					settingsID := user.Settings.ID
					infoID := user.Info.ID

					err = user.RetrieveUserInfo(conn)
					So(err, ShouldEqual, nil)
					So(settingsID, ShouldEqual, user.Settings.ID)

					err = user.RetrieveUserSettings(conn)
					So(err, ShouldEqual, nil)
					So(infoID, ShouldEqual, user.Info.ID)
				})

				Convey("Deleting the user should not produce any errors", func() {
					success, err := user.Remove(conn)
					So(err, ShouldEqual, nil)
					So(success, ShouldEqual, true)
				})

				Convey("Deleting the info and settings should not produce any errors", func() {
					success, err := user.Settings.Remove(conn)
					So(err, ShouldEqual, nil)
					So(success, ShouldEqual, true)

					success, err = user.Info.Remove(conn)
					So(err, ShouldEqual, nil)
					So(success, ShouldEqual, true)
				})
			})
		})
	})
}
