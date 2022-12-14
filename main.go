package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/rs/xid"
)

var templ *template.Template

type Todo struct {
	Id          string `json:"id"`
	Item        string `json:"item"`
	IsComplited bool   `json:"is_complited"`
}

type PageData struct {
	Title string
	Todos []Todo
}

func CreateFileIfNotExist(path string, content string) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(path)
			if err != nil {
				panic(err)
			} else {
				file.WriteString(content)
			}
		}
	}
}

func GetTodoList(w http.ResponseWriter, r *http.Request) {
	var todos []Todo
	path := "./todos.json"

	CreateFileIfNotExist(path, "[]")

	byteValue, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(byteValue, &todos)
	if err != nil {
		panic(err)
	}

	data := PageData{
		Title: "Daily ToDo List",
		Todos: todos,
	}

	templ.Execute(w, data)
}

func CreateNewTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод запрещен!", http.StatusMethodNotAllowed)

		return
	}

	var todos []Todo
	path := "./todos.json"
	todo := r.FormValue("todo")

	defer r.Body.Close()

	guid := xid.New().String()

	newTodo := Todo{
		Id:          guid,
		Item:        todo,
		IsComplited: false,
	}

	byteValue, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(byteValue, &todos)
	if err != nil {
		panic(err)
	}
	todos = append(todos, newTodo)
	fmt.Println(newTodo, todos)

	jsonTodos, err := json.MarshalIndent(&todos, "", "    ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("todos.json", jsonTodos, 0666)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	var todos []Todo
	todoId := r.FormValue("id")
	path := "./todos.json"

	byteValue, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(byteValue, &todos)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(todos); i++ {
		if todos[i].Id == todoId {
			todos = append(todos[:i], todos[i+1:]...)
			break
		}
	}

	jsonTodos, err := json.MarshalIndent(&todos, "", "    ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("todos.json", jsonTodos, 0666)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func ChangeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	type Id struct {
		Id string
	}

	var todoId Id
	var todos []Todo
	path := "./todos.json"

	err := json.NewDecoder(r.Body).Decode(&todoId)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	byteValue, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(byteValue, &todos)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(todos); i++ {
		if todos[i].Id == todoId.Id {
			todos[i].IsComplited = !todos[i].IsComplited
			break
		}
	}

	jsonTodos, err := json.MarshalIndent(&todos, "", "    ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("todos.json", jsonTodos, 0666)

	fmt.Println(todoId.Id)
}

func DeleteAllTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	path := "./todos.json"
	err := os.WriteFile(path, []byte("[]"), 0666)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func main() {
	mux := http.NewServeMux()
	templ = template.Must(template.ParseFiles("templates/index.gohtml"))
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", GetTodoList)
	mux.HandleFunc("/todo/create", CreateNewTodo)
	mux.HandleFunc("/todo/delete", DeleteTodo)
	mux.HandleFunc("/todo/status", ChangeStatus)
	mux.HandleFunc("/todo/clear-all", DeleteAllTodo)
	port := ":3001"

	log.Fatal(http.ListenAndServe(port, mux))
}
