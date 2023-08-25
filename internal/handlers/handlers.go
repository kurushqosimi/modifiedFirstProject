package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"log"
	"main/pkg/models"
	"net/http"
	"strconv"
	"time"
)

// todo we need to close connection with db after closing the app
func Create(response http.ResponseWriter, request *http.Request) {
	dbUri := "host=localhost port = 5432 user = app password = pass dbname = db"
	db, err := gorm.Open(postgres.Open(dbUri), &gorm.Config{}) //installs connection with database
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	var note models.Notes
	bytes, err := io.ReadAll(request.Body)
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(bytes, &note)
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	note.Active = true
	fmt.Println(note)
	sqlQuery := `insert into notes (content) values (?)`
	err = db.Exec(sqlQuery, note.Content).Error
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	response.Write([]byte("Success"))
}
func Read(response http.ResponseWriter, request *http.Request) { //todo make a list of all the notes, declare global variable with db connection
	dbUri := "host=localhost port = 5432 user = app password = pass dbname = db"
	db, err := gorm.Open(postgres.Open(dbUri), &gorm.Config{}) //installs connection with database
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(request)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	sqlQuery := `select * from notes where id = ? and active = true`
	var note models.Notes
	err = db.Raw(sqlQuery, id).Scan(&note).Error
	if note.Content == "" {
		response.Write([]byte("There is not such note!"))
		return
	}
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(struct {
		Content string    `json:"content"`
		Date    time.Time `json:"date"`
	}{
		Content: note.Content,
		Date:    note.Date,
	})
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	response.Write(jsonData)
}
func Update(response http.ResponseWriter, request *http.Request) {
	dbUri := "host=localhost port = 5432 user = app password = pass dbname = db"
	db, err := gorm.Open(postgres.Open(dbUri), &gorm.Config{}) //installs connection with database
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(request)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	bytes, err := io.ReadAll(request.Body)
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	var updatedNote models.Notes
	err = json.Unmarshal(bytes, &updatedNote)
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	sqlQuery := `update notes set content = ?, updated_at = ? where id = ? and active = true`
	var check models.Notes
	err = db.Exec(sqlQuery, updatedNote.Content, time.Now(), id).Error
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	sqlQuery = `select * from notes where id = ? and active = true`
	err = db.Raw(sqlQuery, id).Scan(&check).Error
	log.Println(check)
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	if check.Content == "" {
		response.Write([]byte("There is not such note!"))
		return
	}
	response.Write([]byte("Success"))
}
func Delete(response http.ResponseWriter, request *http.Request) {
	dbUri := "host=localhost port = 5432 user = app password = pass dbname = db"
	db, err := gorm.Open(postgres.Open(dbUri), &gorm.Config{}) //installs connection with database
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(request)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	var check models.Notes
	sqlQuery := `select * from notes where id = ? and active = true`
	err = db.Raw(sqlQuery, id).Scan(&check).Error
	log.Println(check)
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	if check.Content == "" {
		response.Write([]byte("There is not such note!"))
		return
	}
	response.Write([]byte("Success"))
	sqlQuery = `update notes set active = false, deleted_at = ? where id = ? and active = true`
	err = db.Exec(sqlQuery, time.Now(), id).Error
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
}
