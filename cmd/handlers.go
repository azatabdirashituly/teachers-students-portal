package cmd

import (
	"SantaWeb/db"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte("your-secret-key"))

type errorss struct {
	ErrorCode int
	ErrorMsg  string
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		ErrorHandler(w, r, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	err := renderTemplate(w, "home.html", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
}

func teachLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		phone := r.FormValue("phone")
		password := r.FormValue("password")
		incMsg := "Wrong password or phone"

		collection := db.Client.Database("EduPortal").Collection("teachers")
		var teacher Teacher
		err := collection.FindOne(context.Background(), bson.M{"phone": phone}).Decode(&teacher)
		if err != nil {
			renderTemplate(w, "vollogin.html", incMsg)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(teacher.Password), []byte(password))
		if err != nil {
			renderTemplate(w, "vollogin.html", incMsg)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/vol/%s", teacher.ID.Hex()), http.StatusSeeOther)
	} else if r.Method == "GET" {
		renderTemplate(w, "vollogin.html", nil)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}
}

func teachRegHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		firstName := r.FormValue("firstName")
		lastName := r.FormValue("lastName")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		password := r.FormValue("password")

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		teacher := Teacher{
			Name:     firstName,
			Surname:  lastName,
			Email:    email,
			Phone:    phone,
			Password: string(hashedPassword),
			Student:  nil,
		}

		collection := db.Client.Database("EduPortal").Collection("teachers")

		result, err := collection.InsertOne(context.Background(), teacher)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
			return
		}

		insertedID := result.InsertedID.(primitive.ObjectID)
		http.Redirect(w, r, fmt.Sprintf("/vol/%s", insertedID.Hex()), http.StatusSeeOther)
	} else if r.Method == "GET" {
		renderTemplate(w, "volreg.html", nil)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}
}

func studLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		phone := r.FormValue("phone")
		password := r.FormValue("password")
		incMsg := "Wrong password or phone"

		collection := db.Client.Database("EduPortal").Collection("students")
		var student Student
		err := collection.FindOne(context.Background(), bson.M{"phone": phone}).Decode(&student)
		if err != nil {
			renderTemplate(w, "chilog.html", incMsg)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(student.Password), []byte(password))
		if err != nil {
			renderTemplate(w, "chilog.html", incMsg)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/chil/%s", student.ID.Hex()), http.StatusSeeOther)
	} else if r.Method == "GET" {
		renderTemplate(w, "chilog.html", nil)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}
}

func studRegHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		firstName := r.FormValue("firstName")
		lastName := r.FormValue("lastName")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		password := r.FormValue("password")

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
			return
		}

		student := Student{
			Name:     firstName,
			Surname:  lastName,
			Email:    email,
			Phone:    phone,
			Password: string(hashedPassword),
			Teacher:  nil,
		}

		collection := db.Client.Database("EduPortal").Collection("students")

		result, err := collection.InsertOne(context.Background(), student)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
			return
		}

		insertedID := result.InsertedID.(primitive.ObjectID)

		http.Redirect(w, r, fmt.Sprintf("/chil/%s", insertedID.Hex()), http.StatusSeeOther)
	} else if r.Method == "GET" {
		renderTemplate(w, "chireg.html", nil)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	t, err := template.ParseFiles("frontend/templates/" + tmpl)
	if err != nil {
		return err
	}
	err = t.Execute(w, data)
	if err != nil {
		return err
	}
	return nil
}

func teachPersonalPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teacherID := vars["id"]

	var teacher Teacher
	collection := db.Client.Database("EduPortal").Collection("teachers")
	objID, _ := primitive.ObjectIDFromHex(teacherID)

	err := collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&teacher)
	if err != nil {
		http.Error(w, "teacher not found", http.StatusNotFound)
		return
	}

	renderTemplate(w, "vol.html", teacher)
}

func studPersonalPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	studentID := vars["id"]

	var student Student
	collection := db.Client.Database("EduPortal").Collection("students")
	objID, _ := primitive.ObjectIDFromHex(studentID)

	err := collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&student)
	if err != nil {
		http.Error(w, "student not found", http.StatusNotFound)
		return
	}

	renderTemplate(w, "chil.html", student)
}

func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func getChildIDFromSession(r *http.Request) (primitive.ObjectID, error) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		return primitive.NilObjectID, err
	}

	childID, ok := session.Values["childID"].(string)
	if !ok {
		return primitive.NilObjectID, errors.New("Child ID not found in session")
	}

	objID, err := primitive.ObjectIDFromHex(childID)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return objID, nil
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, errCode int, msg string) {
	t, err := template.ParseFiles("frontend/templates/Error.html")
	if err != nil {
		// w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	Errors := errorss{
		ErrorCode: errCode,
		ErrorMsg:  msg,
	}
	// w.WriteHeader(Errors.ErrorCode)
	t.Execute(w, Errors)
}
