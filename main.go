package main

import (
	"database/sql"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine"
)

var db *sql.DB

var err error

func signupPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		t, _ := template.ParseFiles("templates/signup.html")
		t.Execute(res, nil)
		return
	}

	username := html.EscapeString(req.FormValue("username"))
	password := html.EscapeString(req.FormValue("password"))

	var user string

	err := db.QueryRow("SELECT username FROM users WHERE username=?", username).Scan(&user)
	log.Println(err)
	switch {
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(res, "Hash error, unable to create your account.", 500)
			return
		}

		_, err = db.Exec("INSERT INTO users(username, password) VALUES(?, ?)", username, hashedPassword)
		if err != nil {
			http.Error(res, "Insert error, unable to create your account.", 500)
			return
		}

		res.Write([]byte("User created!"))
		return
	case err != nil:
		http.Error(res, "Existing user error, unable to create your account.", 500)
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
		t, _ := template.ParseFiles("templates/login.html")
		t.Execute(res, varmap)
		return
	}
	//logging
	req.ParseForm()
	username := html.EscapeString(req.FormValue("username"))
	password := html.EscapeString(req.FormValue("password"))
	log.Println(time.Now().Format(time.RFC850), "User Login Attempt by: ", username)
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

	res.Write([]byte("Hello " + databaseUsername))

}

func homePage(res http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(res, nil)
}

func main() {

	debugmode := "true"
	if debugmode == "true" {
		db, err = sql.Open("mysql", "jharvard:crimson@tcp(localhost:3306)/GoJudge")
	} else {
		connectionName := mustGetenv("CLOUDSQL_CONNECTION_NAME")
		user := mustGetenv("CLOUDSQL_USER")
		password := os.Getenv("CLOUDSQL_PASSWORD") // NOTE: password may be empty
		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@cloudsql(%s)/GoJudge", user, password, connectionName))
	}

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
	err := http.ListenAndServe(":9000", nil) // setup listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	appengine.Main()
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Panicf("%s environment variable not set.", k)
	}
	return v
}
