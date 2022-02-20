package scrobble

import (
	"net/url"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ScrobblerTestSuite struct {
	suite.Suite
	db     *gorm.DB
	dbmock sqlmock.Sqlmock
}

func (suite *ScrobblerTestSuite) SetupTest() {

	db, mock, err := sqlmock.New()

	if err != nil {
		suite.Errorf(err, "Failed to init sqlmock")
	}

	var dialector gorm.Dialector

	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.18"))
	dialector = mysql.Dialector{Config: &mysql.Config{
		Conn: db,
	}}

	gdb, err := gorm.Open(dialector)

	if err != nil {
		suite.Errorf(err, "Failed to init db")
	}

	suite.db = gdb
	suite.dbmock = mock
}

func (suite *ScrobblerTestSuite) TestValidateScrobbleType() {

	scrobbler := NewScrobbler(suite.db)

	form := url.Values{}
	form.Add("type", "test")

	// test that an invalid scrobble type fails
	err := scrobbler.ValidateType(&form)
	suite.EqualError(err, "unknown/invalid scrobble type test")

	// test that valid scrobble types pass
	for scrobbleType := range ScrobbleTypeNames {
		form := url.Values{}
		form.Add("type", scrobbleType)

		err := scrobbler.ValidateType(&form)
		suite.Nil(err)
	}

}

func TestScrobble(t *testing.T) {

	suite.Run(t, &ScrobblerTestSuite{})

}
