package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"time"
)

var db *sql.DB

type Actor struct {
	Id           uint64 `json:"id,omitempty"`
	Name         string `json:"name"`
	Gender       string `json:"gender" valid:"in(Male|Female)"`
	BirthdayStr  string `json:"birthday"`
	BirthdayTime time.Time
}

type Film struct {
	Id              uint64 `json:"id,omitempty"`
	Name            string `json:"name"`
	ReleaseDateStr  string `json:"release_date"`
	ReleaseDateTime time.Time
	Rating          int   `json:"rating"`
	Actors          []int `json:"actors"`
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email" valid:"email"`
	Password string `json:"password"`
	Role     string `json:"-"`
}

type Session struct {
	UID        int       `json:"uid"`
	Cookie     string    `json:"cookie"`
	ExpireDate time.Time `json:"expire_date"`
}

func handleCreateActor(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var actor Actor
	err = json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// create actor
	_, err = db.Exec("INSERT INTO actors (name, gender, birthday) VALUES (?, ?, ?)",
		actor.Name, actor.Gender, actor.BirthdayStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	actor.BirthdayTime, err = time.Parse("2006-01-02", actor.BirthdayStr)
	w.WriteHeader(http.StatusCreated)
}

func handleDeleteActor(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var actor Actor
	err = json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		fmt.Println(err)
	}

	// delete Actor from Actor table
	_, err = db.Exec("DELETE FROM actors WHERE actor_id = ?",
		actor.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// delete Actor from FilmMembership table
	_, err = db.Exec("DELETE FROM FilmMembership WHERE actor = ?",
		actor.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleGetActors(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT a.name AS actor_name, f.name AS film_name FROM actors as a
		JOIN FilmMembership as fm ON a.actor_id = fm.actor
		JOIN films as f ON fm.film = f.film_id;`
	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []map[string]interface{}
	columns, err := rows.Columns()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}
		err = rows.Scan(values...)
		if err != nil {
			log.Println(err)
			continue
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			val := *(values[i].(*interface{}))

			// Обработка различных типов данных
			switch v := val.(type) {
			case []byte:
				entry[col] = string(v) // Преобразование []byte в строку
			default:
				entry[col] = val // Если тип неизвестен, просто добавляем в результат как есть
			}
		}
		result = append(result, entry)
	}

	json.NewEncoder(w).Encode(result)
}

func handleUpdateActor(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var actor Actor
	err = json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		fmt.Println(err)
	}

	_, err = db.Exec("UPDATE actors SET name=?, gender=?, birthday=? WHERE actor_id=?",
		actor.Name, actor.Gender, actor.BirthdayStr, actor.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	actor.BirthdayTime, err = time.Parse("2006-01-02", actor.BirthdayStr)

	w.WriteHeader(http.StatusCreated)
}

func handleCreateFilm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var film Film
	err = json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		fmt.Println(err)
	}

	// create film
	result, err := db.Exec("INSERT INTO films (name, release_date, rating) VALUES (?, ?, ?)",
		film.Name, film.ReleaseDateStr, film.Rating)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	film.ReleaseDateTime, err = time.Parse("2006-01-02", film.ReleaseDateStr)

	// Получение ID новой записи
	filmId, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	for _, actorID := range film.Actors {
		_, err = db.Exec("INSERT INTO FilmMembership (actor, film) VALUES (?, ?)",
			actorID, filmId)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func handleDeleteFilm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var film Film
	err = json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		fmt.Println(err)
	}

	// delete Film from Actor table
	_, err = db.Exec("DELETE FROM films WHERE film_id = ?",
		film.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// delete Film from FilmMembership table
	_, err = db.Exec("DELETE FROM FilmMembership WHERE actor = ?",
		film.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleUpdateFilm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var film Film
	err = json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		fmt.Println(err)
	}

	_, err = db.Exec("UPDATE films SET name=?, release_date=?, rating=? WHERE film_id=?",
		film.Name, film.ReleaseDateStr, film.Rating, film.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	film.ReleaseDateTime, err = time.Parse("2006-01-02", film.ReleaseDateStr)

	for _, actorID := range film.Actors {
		_, err = db.Exec("UPDATE FilmMembership SET actor=? WHERE film=?",
			actorID, film.Id)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func handleGetAllFilms(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sort_by")

	query := `
		SELECT * FROM films ORDER BY
		CASE
			WHEN ? = 'name' THEN name
			WHEN ? = 'release_date' THEN release_date
			ELSE rating
    	END DESC;`
	rows, err := db.Query(query, sortBy, sortBy)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []map[string]interface{}
	columns, err := rows.Columns()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}
		err = rows.Scan(values...)
		if err != nil {
			log.Println(err)
			continue
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			val := *(values[i].(*interface{}))

			// Обработка различных типов данных
			switch v := val.(type) {
			case []byte:
				entry[col] = string(v) // Преобразование []byte в строку
			default:
				entry[col] = val // Если тип неизвестен, просто добавляем в результат как есть
			}
		}
		result = append(result, entry)
	}

	json.NewEncoder(w).Encode(result)
}

func handleGetFilm(w http.ResponseWriter, r *http.Request) {
	fragment := r.URL.Query().Get("fragment")

	query := `
		SELECT DISTINCT f.name FROM films as f
		JOIN FilmMembership as fm ON f.film_id = fm.film
		JOIN actors as a ON fm.actor = a.actor_id
		WHERE f.name LIKE CONCAT('%', ?, '%')
		OR a.name LIKE CONCAT('%', ?, '%');`
	rows, err := db.Query(query, fragment, fragment)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []map[string]interface{}
	columns, err := rows.Columns()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}
		err = rows.Scan(values...)
		if err != nil {
			log.Println(err)
			continue
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			val := *(values[i].(*interface{}))

			// Обработка различных типов данных
			switch v := val.(type) {
			case []byte:
				entry[col] = string(v) // Преобразование []byte в строку
			default:
				entry[col] = val // Если тип неизвестен, просто добавляем в результат как есть
			}
		}
		result = append(result, entry)
	}

	json.NewEncoder(w).Encode(result)
}

func signIn(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Проверка имени пользователя и пароля в базе данных
	err = db.QueryRow("SELECT id FROM users WHERE email = ? AND password = ?", user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Генерация случайного значения для Cookie
	randBytes := make([]byte, 16)
	_, err = rand.Read(randBytes)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cookieValue := base64.URLEncoding.EncodeToString(randBytes)

	// Время жизни Cookie (7 дней)
	expireDate := time.Now().Add(7 * 24 * time.Hour)

	// Сохранение информации о сессии в базе данных
	_, err = db.Exec("INSERT INTO sessions (uid, cookie, expire_date) VALUES (?, ?, ?)", user.ID, cookieValue, expireDate)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Установка Cookie в заголовок ответа
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    cookieValue,
		Expires:  expireDate,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

func signUp(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Проверка наличия пользователя с таким же email в базе данных
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", user.Email).Scan(&count)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Email already exists", http.StatusBadRequest)
		return
	}

	// Вставка нового пользователя в базу данных
	_, err = db.Exec("INSERT INTO users (email, password, role) VALUES (?, ?, ?)", user.Email, user.Password, "user")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		log.Printf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/sign-in" || path == "/sign-up" || path == "/actors/get" || path == "/films/get" || path == "/film/get" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("session_id")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Получение информации о сессии из базы данных
		var session Session
		var expireDateStr string
		err = db.QueryRow("SELECT uid, expire_date FROM sessions WHERE cookie = ?", cookie.Value).Scan(&session.UID, &expireDateStr)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		session.ExpireDate, err = time.Parse("2006-01-02 15:04:05", expireDateStr)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var userRole string
		err = db.QueryRow("SELECT role FROM users WHERE id = ?", session.UID).Scan(&userRole)
		if err != nil {
			log.Println(err)
			http.Error(w, "cant found user", http.StatusNotFound)
			return
		}

		if userRole != "admin" {
			log.Println("")
			http.Error(w, "user has wrong role", http.StatusUnauthorized)
			return
		}

		// Проверка времени жизни сессии
		if time.Now().After(session.ExpireDate) {
			// Если сессия истекла, удаляем запись из базы данных и возвращаем ошибку Unauthorized
			_, err = db.Exec("DELETE FROM sessions WHERE cookie = ?", cookie.Value)
			if err != nil {
				log.Println(err)
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Обновление Cookie в заголовке ответа
		http.SetCookie(w, cookie)

		// Продолжаем обработку запроса
		next.ServeHTTP(w, r)
	})
}

func main() {
	// db
	var err error
	db, err = sql.Open("mysql",
		"user:mypassword@tcp(localhost:8765)/testdb?&charset=utf8&interpolateParams=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// routs
	mux.HandleFunc("/sign-in", signIn)
	mux.HandleFunc("/sign-up", signUp)
	mux.HandleFunc("/actors/add", handleCreateActor)
	mux.HandleFunc("/actors/delete", handleDeleteActor)
	mux.HandleFunc("/actors/get", handleGetActors)
	mux.HandleFunc("/actors/update", handleUpdateActor)
	mux.HandleFunc("/films/add", handleCreateFilm)
	mux.HandleFunc("/films/delete", handleDeleteFilm)
	mux.HandleFunc("/films/update", handleUpdateFilm)
	mux.HandleFunc("/films/get", handleGetAllFilms)
	mux.HandleFunc("/film/get", handleGetFilm)

	// middleware
	handler := authMiddleware(mux)
	handler = Logging(handler)

	fmt.Println("Server up!")
	log.Fatal(http.ListenAndServe("localhost:8080", handler))
}
