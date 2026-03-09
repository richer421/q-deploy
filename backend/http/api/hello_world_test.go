package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/richer421/q-deploy/http/common"
	"github.com/richer421/q-deploy/infra/mysql/dao"
	"github.com/richer421/q-deploy/pkg/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ---------- helpers ----------

func setupMockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	gormDB, mock, err := testutil.NewMockDB()
	require.NoError(t, err)
	dao.SetDefault(gormDB)
	return mock
}

func newTestContext(t *testing.T, method, url string, params ...gin.Param) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, url, nil)
	c.Params = params
	return w, c
}

func helloWorldRow() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).
		AddRow(1, "test", "desc", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
}

func parseResp(t *testing.T, w *httptest.ResponseRecorder) common.Response {
	t.Helper()
	var resp common.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	return resp
}

// ---------- tests ----------

func TestHelloWorldAPI_List_Success(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectQuery("SELECT \\*").WillReturnRows(helloWorldRow())

	w, c := newTestContext(t, "GET", "/api/v1/hello-world?page=1&page_size=10")
	NewHelloWorldAPI().List(c)

	resp := parseResp(t, w)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, resp.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHelloWorldAPI_List_MissingParams(t *testing.T) {
	w, c := newTestContext(t, "GET", "/api/v1/hello-world")
	NewHelloWorldAPI().List(c)

	resp := parseResp(t, w)
	assert.Equal(t, -1, resp.Code)
}

func TestHelloWorldAPI_Get_Success(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectQuery("SELECT \\*").WillReturnRows(helloWorldRow())

	w, c := newTestContext(t, "GET", "/api/v1/hello-world/1", gin.Param{Key: "id", Value: "1"})
	NewHelloWorldAPI().Get(c)

	resp := parseResp(t, w)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, resp.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHelloWorldAPI_Get_InvalidID(t *testing.T) {
	w, c := newTestContext(t, "GET", "/api/v1/hello-world/abc", gin.Param{Key: "id", Value: "abc"})
	NewHelloWorldAPI().Get(c)

	resp := parseResp(t, w)
	assert.Equal(t, -1, resp.Code)
}
