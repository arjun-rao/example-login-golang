package main

import (
	"database/sql"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var err error

func signupPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		t, _ := template.ParseFiles("signup.gtpl")
		t.Execute(res, nil)
		return
	}

	username := html.EscapeString(req.FormValue("username"))
	password := html.EscapeString(req.FormValue("password"))

	var user string

	err := db.QueryRow("SELECT username FROM users WHERE username=?", username).Scan(&user)

	switch {
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		_, err = db.Exec("INSERT INTO users(username, password) VALUES(?, ?)", username, hashedPassword)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		res.Write([]byte("User created!"))
		return
	case err != nil:
		http.Error(res, "Server error, unable to create your account.", 500)
		return
	default:
		http.Redirect(res, req, "/", 301)
	}
}

func loginPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {

		message := "Enter username and password to login!"
		retry := req.URL.Query().Get("retry")
		checkRetry, _ := strconv.ParseBool(retry)
		varmap := map[string]interface{}{
			"Message": message,
			"Status":  "",
		}
		if checkRetry == true {
			message = "Invalid Username or Password!"
			varmap["Message"] = message
			varmap["Status"] = "error"
		}

		//http.ServeFile(res, req, "login.html")
		t, _ := template.ParseFiles("login.gtpl")
		t.Execute(res, varmap)
		return
	}
	//logging
	req.ParseForm()
	username := html.EscapeString(req.FormValue("username"))
	password := html.EscapeString(req.FormValue("password"))
	fmt.Println(time.Now().Format(time.RFC850), "User Login Attempt by: ", username)

	var databaseUsername string
	var databasePassword string

	err := db.QueryRow("SELECT username, password FROM users WHERE username=?", username).Scan(&databaseUsername, &databasePassword)

	if err != nil {
		http.Redirect(res, req, "/login?retry=1", 301)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	if err != nil {
		http.Redirect(res, req, "/login?retry=1", 301)
		return
	}

	res.Write([]byte("Hello" + databaseUsername))

}

func homePage(res http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("index.gtpl")
	t.Execute(res, nil)
}

func main() {
	db, err = sql.Open("mysql", "jharvard:crimson@tcp(127.0.0.1:3306)/GoJudge")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	http.HandleFunc("/signup", signupPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/", homePage)
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	fmt.Println("Listening on 127.0.0.1:8080")
	err := http.ListenAndServe(":8080", nil) // setup listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
