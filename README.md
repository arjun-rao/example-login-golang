
### How To Run 

####TODO
    -Logout
    -Sessions
    -Integration with Google Sign-in

Create a new database with a users table 

```sql
CREATE TABLE users(
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50),
    password VARCHAR(120)
);
```

Go get both required packages listed below 

```bash
go get golang.org/x/crypto/bcrypt

go get github.com/go-sql-driver/mysql
```

Inside of **main.go** line **111** replace with your own credentials:

```go
db, err = sql.Open("mysql", "myUsername:myPassword@tcp(127.0.0.1:3306)/myDatabase")
```

Finally run as:
```go
go run main.go
```









