package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"

    "golang.org/x/crypto/bcrypt"
)

type Credentials struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

var userPasswords = make(map[string]string)

func main() {
    // Load your users from users.txt
    err := loadUserPasswords("users.txt")
    if err != nil {
        log.Fatalf("Error loading users: %v\n", err)
    }

    // Routes
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/profile", profileHandler)

    // If you are serving a built React app from ./frontend/build:
    //   fs := http.FileServer(http.Dir("./frontend/build"))
    //   http.Handle("/", fs)

    fmt.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadUserPasswords(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, ":")
        if len(parts) == 2 {
            username := parts[0]
            hashedPassword := parts[1]
            userPasswords[username] = hashedPassword
        }
    }
    return scanner.Err()
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var creds Credentials
    err := json.NewDecoder(r.Body).Decode(&creds)
    if err != nil {
        http.Error(w, "Cannot parse request body", http.StatusBadRequest)
        return
    }

    hashedPass, ok := userPasswords[creds.Username]
    if !ok {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(creds.Password)); err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "success",
        "message": "Logged in!",
    })
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
    // In a real app, you'd check for session/JWT here.
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "success",
        "profile": "This is the protected profile data!",
    })
}

