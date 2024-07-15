package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
)

func TestCreateUser(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	db = mockDB
	defer mockDB.Close()

	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))

	user := User{Ads: "test ads", Username: "testuser"}
	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	CreateUser(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}

func TestGetUsers(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	db = mockDB
	defer mockDB.Close()

	rows := sqlmock.NewRows([]string{"userid", "ads", "username"}).
		AddRow(1, "test ads", "testuser")
	mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/users", nil)
	rr := httptest.NewRecorder()

	GetUsers(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGetUserByID(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	db = mockDB
	defer mockDB.Close()

	rows := sqlmock.NewRows([]string{"userid", "ads", "username"}).
		AddRow(1, "test ads", "testuser")
	mock.ExpectQuery("SELECT (.+) FROM users WHERE userid = ?").WithArgs("1").WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/users/1", nil)
	rr := httptest.NewRecorder()
	vars := map[string]string{"userid": "1"}
	req = mux.SetURLVars(req, vars)

	GetUserByID(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestUpdateUser(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	db = mockDB
	defer mockDB.Close()

	mock.ExpectQuery("SELECT userid FROM users WHERE userid = ?").WithArgs("1").WillReturnRows(sqlmock.NewRows([]string{"userid"}).AddRow(1))
	mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs("updated ads", "updateduser", "1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT (.+) FROM users WHERE userid = ?").WithArgs("1").WillReturnRows(sqlmock.NewRows([]string{"userid", "ads", "username"}).AddRow(1, "updated ads", "updateduser"))

	user := User{Ads: "updated ads", Username: "updateduser"}
	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	vars := map[string]string{"userid": "1"}
	req = mux.SetURLVars(req, vars)

	UpdateUser(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
